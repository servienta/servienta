// Package ntpsend queries a receiving application's NTP server (R13) and
// returns its stratum and transmit time.
package ntpsend

import (
	"context"
	"encoding/binary"
	"net"
	"time"
)

const ntpEpochOffset = 2208988800

type Sender struct{}

func (Sender) Name() string { return "ntp" }

func (Sender) Send(ctx context.Context, target string, payload map[string]any) (map[string]any, error) {
	c, err := net.DialTimeout("udp", target, 3*time.Second)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	req := make([]byte, 48)
	req[0] = 0x23 // LI=0, VN=4, Mode=3 (client)
	c.SetDeadline(time.Now().Add(3 * time.Second))
	if _, err := c.Write(req); err != nil {
		return nil, err
	}
	resp := make([]byte, 48)
	n, err := c.Read(resp)
	if err != nil {
		return nil, err
	}
	if n < 48 {
		return map[string]any{"sent": true, "reply": false}, nil
	}
	secs := binary.BigEndian.Uint32(resp[40:44])
	txTime := time.Unix(int64(secs)-ntpEpochOffset, 0).UTC()
	return map[string]any{
		"sent":         true,
		"stratum":      int(resp[1]),
		"transmit_utc": txTime.Format(time.RFC3339),
	}, nil
}
