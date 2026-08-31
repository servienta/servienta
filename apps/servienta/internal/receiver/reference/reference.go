// Package reference is the deliberately trivial receiver required by R10's
// verification: a line-based TCP listener. It is the worked example for the
// receiver documentation — at phase 0 the first receiver, after R3 the ninth.
package reference

import (
	"bufio"
	"context"
	"net"

	"github.com/servienta/servienta/apps/servienta/internal/receiver"
)

type Receiver struct{}

func (Receiver) Name() string { return "reference" }

func (Receiver) Start(ctx context.Context, addr string, rec receiver.Recorder) (net.Addr, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		ln.Close()
	}()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return // listener closed
			}
			go handle(conn, rec)
		}
	}()
	return ln.Addr(), nil
}

func handle(conn net.Conn, rec receiver.Recorder) {
	defer conn.Close()
	host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return
	}
	sc := bufio.NewScanner(conn)
	for sc.Scan() {
		// Parsed content for R4: the reference protocol is one line, one message.
		_ = rec.Record("reference", host, map[string]any{"line": sc.Text()})
	}
}
