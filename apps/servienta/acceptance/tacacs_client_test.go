package acceptance

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"
)

func timeoutCtx(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}

// tacacsAuth sends a minimal TACACS+ authentication START (ASCII) and returns
// the REPLY status byte. Mirrors the server's obfuscation.
func tacacsAuth(t *testing.T, addr, secret, user string) byte {
	t.Helper()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// START body: action(1)=LOGIN, priv_lvl(1), authen_type(1)=ASCII,
	// service(1)=LOGIN, user_len(1), port_len(1), rem_addr_len(1), data_len(1),
	// user, port, rem_addr, data.
	body := []byte{0x01, 0x00, 0x01, 0x01, byte(len(user)), 0x00, 0x00, 0x00}
	body = append(body, []byte(user)...)

	var hdr [12]byte
	hdr[0] = 0xc0 // major 12, minor 0
	hdr[1] = 0x01 // authentication
	hdr[2] = 0x01 // seq_no 1
	hdr[3] = 0x00 // flags: encrypted
	sessionID := []byte{0x00, 0x00, 0x00, 0x2a}
	copy(hdr[4:8], sessionID)
	binary.BigEndian.PutUint32(hdr[8:12], uint32(len(body)))

	pad(body, []byte(secret), sessionID, hdr[0], hdr[2])
	conn.SetDeadline(time.Now().Add(2 * time.Second))
	conn.Write(hdr[:])
	conn.Write(body)

	var rhdr [12]byte
	if _, err := io.ReadFull(conn, rhdr[:]); err != nil {
		t.Fatal(err)
	}
	rlen := binary.BigEndian.Uint32(rhdr[8:12])
	rbody := make([]byte, rlen)
	if _, err := io.ReadFull(conn, rbody); err != nil {
		t.Fatal(err)
	}
	pad(rbody, []byte(secret), sessionID, rhdr[0], rhdr[2])
	return rbody[0] // status
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
