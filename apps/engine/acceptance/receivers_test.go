package acceptance

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/gosnmp/gosnmp"
)

// --- R3.1: syslog over UDP, TCP, and RELP, read back via R4 ---
func TestSyslogAllTransports(t *testing.T) {
	e := startEngine(t)
	e.do(t, "PUT", "/api/v1/runs/run-1", map[string]any{"sources": []string{"127.0.0.1"}})

	// UDP
	sendUDP(t, e.endpoints["syslog-udp"], "<34>udp hello")
	// TCP (newline-framed)
	sendTCP(t, e.endpoints["syslog-tcp"], "<13>tcp hello\n")
	// RELP (framed, with open/syslog/close handshake)
	sendRELP(t, e.endpoints["syslog-relp"], "<13>relp hello")

	msgs := e.receivedN(t, "syslog", "run-1", 3)
	byTransport := map[string]map[string]any{}
	for _, m := range msgs {
		c := m["content"].(map[string]any)
		byTransport[c["transport"].(string)] = c
	}
	for _, tr := range []string{"udp", "tcp", "relp"} {
		c, ok := byTransport[tr]
		if !ok {
			t.Fatalf("no %s message recorded", tr)
		}
		if c["message"] == nil {
			t.Fatalf("%s: message not parsed: %v", tr, c)
		}
	}
	// facility/severity parsed from <34> = facility 4, severity 2
	if u := byTransport["udp"]; u["facility"].(float64) != 4 || u["severity"].(float64) != 2 {
		t.Fatalf("udp priority parsed wrong: %v", u)
	}
}

// --- R3.2: SNMP v2c trap, read back via R4 ---
func TestSNMPv2cTrap(t *testing.T) {
	e := startEngine(t)
	e.do(t, "PUT", "/api/v1/runs/run-1", map[string]any{"sources": []string{"127.0.0.1"}})

	host, portStr, _ := net.SplitHostPort(e.endpoints["snmp-traps"])
	var port int
	fmt.Sscanf(portStr, "%d", &port)
	g := &gosnmp.GoSNMP{
		Target:    host,
		Port:      uint16(port),
		Community: "throwaway-public",
		Version:   gosnmp.Version2c,
		Timeout:   2 * time.Second,
		Retries:   1,
	}
	if err := g.Connect(); err != nil {
		t.Fatal(err)
	}
	defer g.Conn.Close()
	trap := gosnmp.SnmpTrap{Variables: []gosnmp.SnmpPDU{
		{Name: "1.3.6.1.6.3.1.1.4.1.0", Type: gosnmp.ObjectIdentifier, Value: "1.3.6.1.4.1.9999.1"},
		{Name: "1.3.6.1.4.1.9999.2", Type: gosnmp.OctetString, Value: "servienta-trap"},
	}}
	if _, err := g.SendTrap(trap); err != nil {
		t.Fatal(err)
	}

	msgs := e.receivedN(t, "snmp-traps", "run-1", 1)
	c := msgs[0]["content"].(map[string]any)
	vars := c["variables"].(map[string]any)
	if vars[".1.3.6.1.4.1.9999.2"] != "servienta-trap" {
		t.Fatalf("trap variable not recorded: %v", vars)
	}
}

func sendUDP(t *testing.T, addr, msg string) {
	t.Helper()
	c, err := net.Dial("udp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	c.Write([]byte(msg))
}

func sendTCP(t *testing.T, addr, msg string) {
	t.Helper()
	c, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	c.Write([]byte(msg))
}

func sendRELP(t *testing.T, addr, msg string) {
	t.Helper()
	c, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	fmt.Fprintf(c, "1 open 0 \n")
	fmt.Fprintf(c, "2 syslog %d %s\n", len(msg), msg)
	fmt.Fprintf(c, "3 close 0 \n")
	c.SetReadDeadline(time.Now().Add(time.Second))
	buf := make([]byte, 256)
	c.Read(buf) // drain responses
}
