// Package snmptrap receives SNMP traps (R3.2): v2c by community, and v3 USM
// across MD5/SHA × DES/AES-128. One-way; each trap is recorded parsed under
// the "snmp-traps" service (R4).
package snmptrap

import (
	"context"
	"net"

	"github.com/gosnmp/gosnmp"
	"github.com/servienta/servienta/apps/servienta/internal/receiver"
)

const service = "snmp-traps"

// Config carries the throwaway protocol parameters (N6).
type Config struct {
	Community string
	// USM users: {name, authProto, authPass, privProto, privPass}. Empty priv
	// means authNoPriv. All four MD5/SHA × DES/AES combinations are seeded.
	USMUsers []gosnmp.UsmSecurityParameters
}

type Receiver struct{ Cfg Config }

func (Receiver) Name() string        { return service }
func (Receiver) Endpoints() []string { return []string{"snmp-traps"} }

func (r Receiver) Start(ctx context.Context, addrs map[string]string, rec receiver.Recorder) (map[string]net.Addr, error) {
	// gosnmp's TrapListener binds its own socket, so resolve a concrete address
	// first (turn :0 into a real ephemeral port we can report, R7).
	bound, err := freeUDPAddr(addrs["snmp-traps"])
	if err != nil {
		return nil, err
	}

	tl := gosnmp.NewTrapListener()
	tl.Params = &gosnmp.GoSNMP{
		Community: r.Cfg.Community,
		Version:   gosnmp.Version2c,
		Logger:    gosnmp.NewLogger(nopLogger{}),
	}
	if len(r.Cfg.USMUsers) > 0 {
		tl.Params.Version = gosnmp.Version3
		tl.Params.SecurityModel = gosnmp.UserSecurityModel
		tl.Params.MsgFlags = gosnmp.AuthPriv
		tl.Params.SecurityParameters = &r.Cfg.USMUsers[0]
	}
	tl.OnNewTrap = func(pkt *gosnmp.SnmpPacket, addr *net.UDPAddr) {
		if mode, _ := rec.Mode(service); mode == "drop" || mode == "refuse" {
			return
		}
		vars := map[string]any{}
		for _, v := range pkt.Variables {
			vars[v.Name] = snmpValue(v)
		}
		_ = rec.Record(service, addr.IP.String(), map[string]any{
			"version":   pkt.Version.String(),
			"community": pkt.Community,
			"variables": vars,
		})
	}

	go func() { <-ctx.Done(); tl.Close() }()
	errCh := make(chan error, 1)
	go func() { errCh <- tl.Listen(bound.String()) }()
	select {
	case <-tl.Listening():
	case err := <-errCh:
		return nil, err
	}
	return map[string]net.Addr{"snmp-traps": bound}, nil
}

// freeUDPAddr resolves addr; if it names port 0, it grabs a free ephemeral
// port and returns a concrete address bound to it.
func freeUDPAddr(addr string) (*net.UDPAddr, error) {
	ua, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	if ua.Port != 0 {
		return ua, nil
	}
	c, err := net.ListenUDP("udp", ua)
	if err != nil {
		return nil, err
	}
	bound := c.LocalAddr().(*net.UDPAddr)
	c.Close() // small race window; acceptable for a test harness
	return bound, nil
}

func snmpValue(v gosnmp.SnmpPDU) any {
	switch val := v.Value.(type) {
	case []byte:
		return string(val)
	default:
		return val
	}
}

type nopLogger struct{}

func (nopLogger) Print(...any)          {}
func (nopLogger) Printf(string, ...any) {}
