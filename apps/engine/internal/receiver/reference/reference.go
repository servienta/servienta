// Package reference is the deliberately trivial receiver required by R10's
// verification: a line-based TCP listener. It is the worked example for the
// receiver documentation and honors the R9 failure modes.
package reference

import (
	"bufio"
	"context"
	"net"
	"time"

	"github.com/servienta/servienta/apps/engine/internal/receiver"
)

type Receiver struct{}

func (Receiver) Name() string        { return "reference" }
func (Receiver) Endpoints() []string { return []string{"reference"} }

func (Receiver) Start(ctx context.Context, addrs map[string]string, rec receiver.Recorder) (map[string]net.Addr, error) {
	ln, err := net.Listen("tcp", addrs["reference"])
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
			go handle(conn, rec)
		}
	}()
	return map[string]net.Addr{"reference": ln.Addr()}, nil
}

func handle(conn net.Conn, rec receiver.Recorder) {
	defer conn.Close()
	mode, delayMs := rec.Mode("reference")
	switch mode {
	case "refuse", "drop":
		return
	case "delay":
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}
	host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return
	}
	sc := bufio.NewScanner(conn)
	for sc.Scan() {
		if mode == "cut" {
			return
		}
		_ = rec.Record("reference", host, map[string]any{"line": sc.Text()})
	}
}
