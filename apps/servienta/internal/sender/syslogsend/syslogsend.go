// Package syslogsend sends a syslog message to a receiving application (R13),
// over UDP, TCP, or RELP. One-way: confirms the send.
package syslogsend

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/servienta/servienta/apps/servienta/internal/sender"
)

type Sender struct{}

func (Sender) Name() string { return "syslog" }

func (Sender) Send(ctx context.Context, target string, payload map[string]any) (map[string]any, error) {
	msg := sender.Str(payload, "message")
	transport := sender.StrOr(payload, "transport", "udp")
	switch transport {
	case "udp":
		return sendUDP(target, msg)
	case "tcp":
		return sendStream(target, msg+"\n")
	case "relp":
		return sendRELP(target, msg)
	default:
		return nil, fmt.Errorf("unknown transport %q (use udp, tcp, or relp)", transport)
	}
}

func sendUDP(target, msg string) (map[string]any, error) {
	c, err := net.Dial("udp", target)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	if _, err := c.Write([]byte(msg)); err != nil {
		return nil, err
	}
	return map[string]any{"sent": true, "transport": "udp", "bytes": len(msg)}, nil
}

func sendStream(target, msg string) (map[string]any, error) {
	c, err := net.DialTimeout("tcp", target, 3*time.Second)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	if _, err := c.Write([]byte(msg)); err != nil {
		return nil, err
	}
	return map[string]any{"sent": true, "transport": "tcp", "bytes": len(msg)}, nil
}

func sendRELP(target, msg string) (map[string]any, error) {
	c, err := net.DialTimeout("tcp", target, 3*time.Second)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	fmt.Fprintf(c, "1 open 0 \n")
	fmt.Fprintf(c, "2 syslog %d %s\n", len(msg), msg)
	fmt.Fprintf(c, "3 close 0 \n")
	c.SetReadDeadline(time.Now().Add(time.Second))
	buf := make([]byte, 256)
	c.Read(buf)
	return map[string]any{"sent": true, "transport": "relp", "bytes": len(msg)}, nil
}
