// Package tacacs answers TACACS+ authentication (R3.4, RFC 8907) and lets a
// test control the outcome (R8): PASS, or FAIL with a stated reason. Each
// authentication is recorded (R4). Default: PASS. The body obfuscation
// (MD5 pseudo-pad over the shared secret) is implemented in-repo (D6).
package tacacs

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"io"
	"net"

	"github.com/servienta/servienta/apps/engine/internal/receiver"
)

const service = "tacacs"

const (
	authenStatusPass = 0x01
	authenStatusFail = 0x02
	typeReply        = 0x02 // authentication REPLY
)

type Receiver struct{ Secret string }

func (Receiver) Name() string        { return service }
func (Receiver) Endpoints() []string { return []string{"tacacs"} }

func (rv Receiver) Start(ctx context.Context, addrs map[string]string, rec receiver.Recorder) (map[string]net.Addr, error) {
	ln, err := net.Listen("tcp", addrs["tacacs"])
	if err != nil {
		return nil, err
	}
	go func() { <-ctx.Done(); ln.Close() }()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go rv.handle(conn, rec)
		}
	}()
	return map[string]net.Addr{"tacacs": ln.Addr()}, nil
}

func (rv Receiver) handle(conn net.Conn, rec receiver.Recorder) {
	defer conn.Close()
	if mode, _ := rec.Mode(service); mode == "drop" || mode == "refuse" {
		return
	}
	var hdr [12]byte
	if _, err := io.ReadFull(conn, hdr[:]); err != nil {
		return
	}
	length := binary.BigEndian.Uint32(hdr[8:12])
	body := make([]byte, length)
	if _, err := io.ReadFull(conn, body); err != nil {
		return
	}
	sessionID := binary.BigEndian.Uint32(hdr[4:8])
	seq := hdr[2]
	deobfuscate(body, []byte(rv.Secret), hdr[4:8], hdr[0], seq)

	user := parseUser(body)
	_ = rec.Record(service, host(conn.RemoteAddr().String()), map[string]any{"username": user})

	status := byte(authenStatusPass)
	spec := rec.Response(service)
	if str(spec, "outcome") == "fail" {
		status = authenStatusFail
	}
	reply := buildReply(status, str(spec, "reason"))
	sendReply(conn, hdr, sessionID, seq+1, []byte(rv.Secret), reply)
}

// buildReply is an authentication REPLY body: status, flags, server_msg_len(2),
// data_len(2), then server_msg.
func buildReply(status byte, msg string) []byte {
	b := make([]byte, 6+len(msg))
	b[0] = status
	b[1] = 0 // flags
	binary.BigEndian.PutUint16(b[2:4], uint16(len(msg)))
	binary.BigEndian.PutUint16(b[4:6], 0)
	copy(b[6:], msg)
	return b
}

func sendReply(conn net.Conn, reqHdr [12]byte, sessionID uint32, seq byte, secret, body []byte) {
	var hdr [12]byte
	hdr[0] = reqHdr[0] // version
	hdr[1] = 0x01      // type = authentication
	hdr[2] = seq
	hdr[3] = 0 // flags (encrypted body)
	binary.BigEndian.PutUint32(hdr[4:8], sessionID)
	binary.BigEndian.PutUint32(hdr[8:12], uint32(len(body)))
	obfuscate(body, secret, hdr[4:8], hdr[0], seq)
	conn.Write(hdr[:])
	conn.Write(body)
}

// pad builds the MD5 pseudo-pad and XORs it over data (RFC 8907 §4.5).
func pad(data, secret, sessionID []byte, version, seq byte) {
	var prev []byte
	for i := 0; i < len(data); i += 16 {
		h := md5.New()
		h.Write(sessionID)
		h.Write(secret)
		h.Write([]byte{version, seq})
		h.Write(prev)
		sum := h.Sum(nil)
		for j := 0; j < 16 && i+j < len(data); j++ {
			data[i+j] ^= sum[j]
		}
		prev = sum
	}
}

func obfuscate(data, secret, sessionID []byte, version, seq byte) {
	pad(data, secret, sessionID, version, seq)
}
func deobfuscate(data, secret, sessionID []byte, version, seq byte) {
	pad(data, secret, sessionID, version, seq)
}

// parseUser reads the user field of an authentication START body:
// action, priv_lvl, authen_type, service, user_len, port_len, rem_addr_len,
// data_len, then user.
func parseUser(b []byte) string {
	if len(b) < 8 {
		return ""
	}
	userLen := int(b[4])
	if 8+userLen > len(b) {
		return ""
	}
	return string(b[8 : 8+userLen])
}

func host(addr string) string {
	if h, _, err := net.SplitHostPort(addr); err == nil {
		return h
	}
	return addr
}

func str(m map[string]any, k string) string {
	if m == nil {
		return ""
	}
	s, _ := m[k].(string)
	return s
}
