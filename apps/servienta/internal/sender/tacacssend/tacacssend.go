// Package tacacssend sends a TACACS+ authentication START to a receiving
// application's server (R13, RFC 8907) and returns the reply status.
package tacacssend

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/servienta/servienta/apps/servienta/internal/sender"
)

type Sender struct{}

func (Sender) Name() string { return "tacacs" }

var statusName = map[byte]string{0x01: "pass", 0x02: "fail", 0x03: "getdata", 0x05: "getuser", 0x06: "getpass", 0x07: "restart", 0x08: "error"}

func (Sender) Send(ctx context.Context, target string, payload map[string]any) (map[string]any, error) {
	secret := []byte(sender.StrOr(payload, "secret", "throwaway-tacacs"))
	user := sender.StrOr(payload, "username", "user")

	conn, err := net.DialTimeout("tcp", target, 3*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(3 * time.Second))

	body := []byte{0x01, 0x00, 0x01, 0x01, byte(len(user)), 0x00, 0x00, 0x00}
	body = append(body, []byte(user)...)

	var hdr [12]byte
	hdr[0] = 0xc0 // major 12, minor 0
	hdr[1] = 0x01 // authentication
	hdr[2] = 0x01 // seq 1
	hdr[3] = 0x00 // encrypted
	sessionID := []byte{0x00, 0x00, 0x00, 0x2a}
	copy(hdr[4:8], sessionID)
	binary.BigEndian.PutUint32(hdr[8:12], uint32(len(body)))

	pad(body, secret, sessionID, hdr[0], hdr[2])
	conn.Write(hdr[:])
	conn.Write(body)

	var rhdr [12]byte
	if _, err := io.ReadFull(conn, rhdr[:]); err != nil {
		return nil, err
	}
	rlen := binary.BigEndian.Uint32(rhdr[8:12])
	rbody := make([]byte, rlen)
	if _, err := io.ReadFull(conn, rbody); err != nil {
		return nil, err
	}
	pad(rbody, secret, sessionID, rhdr[0], rhdr[2])
	if len(rbody) == 0 {
		return nil, fmt.Errorf("empty TACACS+ reply")
	}
	return map[string]any{"sent": true, "status": statusName[rbody[0]]}, nil
}

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
