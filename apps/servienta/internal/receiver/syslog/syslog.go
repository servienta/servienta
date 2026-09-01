// Package syslog receives syslog messages over UDP, TCP, and RELP (R3.1) and
// records each one parsed under the "syslog" service (R4). One-way: the
// application sends, the harness records.
package syslog

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/servienta/servienta/apps/servienta/internal/receiver"
)

const service = "syslog"

type Receiver struct{}

func (Receiver) Name() string        { return service }
func (Receiver) Endpoints() []string { return []string{"syslog-udp", "syslog-tcp", "syslog-relp"} }

func (Receiver) Start(ctx context.Context, addrs map[string]string, rec receiver.Recorder) (map[string]net.Addr, error) {
	out := map[string]net.Addr{}

	udp, err := net.ListenPacket("udp", addrs["syslog-udp"])
	if err != nil {
		return nil, err
	}
	out["syslog-udp"] = udp.LocalAddr()
	go serveUDP(ctx, udp, rec)

	tcp, err := net.Listen("tcp", addrs["syslog-tcp"])
	if err != nil {
		return nil, err
	}
	out["syslog-tcp"] = tcp.Addr()
	go serveStream(ctx, tcp, rec, "tcp")

	relp, err := net.Listen("tcp", addrs["syslog-relp"])
	if err != nil {
		return nil, err
	}
	out["syslog-relp"] = relp.Addr()
	go serveStream(ctx, relp, rec, "relp")

	return out, nil
}

func serveUDP(ctx context.Context, pc net.PacketConn, rec receiver.Recorder) {
	go func() { <-ctx.Done(); pc.Close() }()
	buf := make([]byte, 64<<10)
	for {
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			return
		}
		if mode, _ := rec.Mode(service); mode == "drop" || mode == "refuse" {
			continue
		}
		record(rec, host(addr.String()), "udp", string(buf[:n]))
	}
}

func serveStream(ctx context.Context, ln net.Listener, rec receiver.Recorder, transport string) {
	go func() { <-ctx.Done(); ln.Close() }()
	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		go func() {
			defer conn.Close()
			if mode, _ := rec.Mode(service); mode == "drop" || mode == "refuse" {
				return
			}
			if transport == "relp" {
				serveRELP(conn, rec)
				return
			}
			sc := bufio.NewScanner(conn)
			sc.Buffer(make([]byte, 64<<10), 1<<20)
			for sc.Scan() {
				record(rec, host(conn.RemoteAddr().String()), "tcp", sc.Text())
			}
		}()
	}
}

// serveRELP implements the RELP framing: "<txnr> <cmd> <datalen> <data>\n",
// answering each command with an "rsp" frame (200 OK) as the protocol requires.
func serveRELP(conn net.Conn, rec receiver.Recorder) {
	r := bufio.NewReader(conn)
	for {
		txnr, err := readToken(r)
		if err != nil {
			return
		}
		cmd, err := readToken(r)
		if err != nil {
			return
		}
		lenTok, err := readToken(r)
		if err != nil {
			return
		}
		n, _ := strconv.Atoi(strings.TrimSpace(lenTok))
		data := make([]byte, n)
		if n > 0 {
			if _, err := readFull(r, data); err != nil {
				return
			}
		}
		r.ReadByte() // trailing '\n'
		switch cmd {
		case "open":
			fmt.Fprintf(conn, "%s rsp %d 200 OK\n", txnr, len("200 OK"))
		case "syslog":
			record(rec, host(conn.RemoteAddr().String()), "relp", string(data))
			fmt.Fprintf(conn, "%s rsp %d 200 OK\n", txnr, len("200 OK"))
		case "close":
			fmt.Fprintf(conn, "%s rsp 0\n", txnr)
			return
		}
	}
}

func record(rec receiver.Recorder, src, transport, raw string) {
	content := map[string]any{"transport": transport, "raw": raw}
	if pri, rest, ok := parsePriority(raw); ok {
		content["facility"] = pri / 8
		content["severity"] = pri % 8
		content["message"] = rest
	}
	_ = rec.Record(service, src, content)
}

// parsePriority reads a leading "<PRI>" per RFC 3164/5424.
func parsePriority(s string) (pri int, rest string, ok bool) {
	if !strings.HasPrefix(s, "<") {
		return 0, "", false
	}
	end := strings.IndexByte(s, '>')
	if end < 0 {
		return 0, "", false
	}
	p, err := strconv.Atoi(s[1:end])
	if err != nil {
		return 0, "", false
	}
	return p, s[end+1:], true
}

func host(addr string) string {
	if h, _, err := net.SplitHostPort(addr); err == nil {
		return h
	}
	return addr
}

func readToken(r *bufio.Reader) (string, error) {
	tok, err := r.ReadString(' ')
	return strings.TrimSuffix(tok, " "), err
}

func readFull(r *bufio.Reader, buf []byte) (int, error) {
	total := 0
	for total < len(buf) {
		n, err := r.Read(buf[total:])
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}
