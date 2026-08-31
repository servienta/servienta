// Package snmpsend sends an SNMP v2c trap to a receiving application (R13).
package snmpsend

import (
	"context"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/servienta/servienta/apps/engine/internal/sender"
)

type Sender struct{}

func (Sender) Name() string { return "snmp-traps" }

func (Sender) Send(ctx context.Context, target string, payload map[string]any) (map[string]any, error) {
	host, port := splitHostPort(target, 162)
	g := &gosnmp.GoSNMP{
		Target:    host,
		Port:      port,
		Community: sender.StrOr(payload, "community", "throwaway-public"),
		Version:   gosnmp.Version2c,
		Timeout:   2 * time.Second,
		Retries:   0,
	}
	if err := g.Connect(); err != nil {
		return nil, err
	}
	defer g.Conn.Close()
	oid := sender.StrOr(payload, "oid", "1.3.6.1.4.1.9999.1")
	value := sender.StrOr(payload, "value", "servienta")
	trap := gosnmp.SnmpTrap{Variables: []gosnmp.SnmpPDU{
		{Name: "1.3.6.1.6.3.1.1.4.1.0", Type: gosnmp.ObjectIdentifier, Value: oid},
		{Name: oid, Type: gosnmp.OctetString, Value: value},
	}}
	if _, err := g.SendTrap(trap); err != nil {
		return nil, err
	}
	return map[string]any{"sent": true}, nil
}
