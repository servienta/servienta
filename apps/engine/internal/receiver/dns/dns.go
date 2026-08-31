// Package dns answers DNS queries (R3.5) and lets a test control the reply
// (R8): a valid record, NXDOMAIN, SERVFAIL, or a delayed answer. Each query is
// also recorded (R4). Default: resolve to a fixed valid A record.
package dns

import (
	"context"
	"net"
	"time"

	"github.com/miekg/dns"
	"github.com/servienta/servienta/apps/engine/internal/receiver"
)

const service = "dns"

// DefaultA is the successful answer when no R8 control is set (N6: fixed).
const DefaultA = "192.0.2.1"

type Receiver struct{}

func (Receiver) Name() string        { return service }
func (Receiver) Endpoints() []string { return []string{"dns"} }

func (Receiver) Start(ctx context.Context, addrs map[string]string, rec receiver.Recorder) (map[string]net.Addr, error) {
	pc, err := net.ListenPacket("udp", addrs["dns"])
	if err != nil {
		return nil, err
	}
	srv := &dns.Server{PacketConn: pc, Handler: dns.HandlerFunc(func(w dns.ResponseWriter, req *dns.Msg) {
		handle(w, req, rec)
	})}
	go func() { <-ctx.Done(); srv.Shutdown() }()
	go srv.ActivateAndServe()
	return map[string]net.Addr{"dns": pc.LocalAddr()}, nil
}

func handle(w dns.ResponseWriter, req *dns.Msg, rec receiver.Recorder) {
	if mode, _ := rec.Mode(service); mode == "drop" || mode == "refuse" {
		return
	}
	name := ""
	if len(req.Question) > 0 {
		name = req.Question[0].Name
	}
	src := host(w.RemoteAddr().String())
	_ = rec.Record(service, src, map[string]any{"qname": name})

	m := new(dns.Msg)
	m.SetReply(req)

	spec := rec.Response(service)
	switch str(spec, "outcome") {
	case "nxdomain":
		m.Rcode = dns.RcodeNameError
	case "servfail":
		m.Rcode = dns.RcodeServerFailure
	case "delay":
		time.Sleep(time.Duration(intv(spec, "delay_ms")) * time.Millisecond)
		answerA(m, name, addrOr(spec, DefaultA))
	default:
		answerA(m, name, addrOr(spec, DefaultA))
	}
	w.WriteMsg(m)
}

func answerA(m *dns.Msg, name, ip string) {
	if len(m.Question) == 0 {
		return
	}
	rr, err := dns.NewRR(name + " 60 IN A " + ip)
	if err == nil {
		m.Answer = append(m.Answer, rr)
	}
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

func addrOr(m map[string]any, def string) string {
	if s := str(m, "address"); s != "" {
		return s
	}
	return def
}

func intv(m map[string]any, k string) int {
	if m == nil {
		return 0
	}
	if f, ok := m[k].(float64); ok {
		return int(f)
	}
	return 0
}
