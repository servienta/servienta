package acceptance

import (
	"net"
	"testing"
	"time"

	"context"

	"github.com/miekg/dns"
	"layeh.com/radius"
)

// R13: for each protocol, stand up a listener (the application under test),
// ask the engine to send to it, and confirm receipt (and reply for
// request-response). The engine only sends to the target we name (D18).

// --- syslog send (one-way, UDP) ---
func TestSendSyslogUDP(t *testing.T) {
	e := startEngine(t)
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer pc.Close()

	res, body := e.do(t, "POST", "/api/v1/send/syslog", map[string]any{
		"target": pc.LocalAddr().String(), "transport": "udp", "message": "<34>hello app",
	})
	if res.StatusCode != 200 {
		t.Fatalf("send: %d %s", res.StatusCode, body)
	}
	pc.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 256)
	n, _, err := pc.ReadFrom(buf)
	if err != nil {
		t.Fatalf("listener did not receive: %v", err)
	}
	if string(buf[:n]) != "<34>hello app" {
		t.Fatalf("wrong message: %q", buf[:n])
	}
}

// --- DNS send (request-response): engine queries our server, returns its answer ---
func TestSendDNS(t *testing.T) {
	e := startEngine(t)
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := &dns.Server{PacketConn: pc, Handler: dns.HandlerFunc(func(w dns.ResponseWriter, req *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(req)
		rr, _ := dns.NewRR(req.Question[0].Name + " 60 IN A 10.11.12.13")
		m.Answer = append(m.Answer, rr)
		w.WriteMsg(m)
	})}
	go srv.ActivateAndServe()
	defer srv.Shutdown()

	res, body := e.do(t, "POST", "/api/v1/send/dns", map[string]any{
		"target": pc.LocalAddr().String(), "qname": "example.test",
	})
	if res.StatusCode != 200 {
		t.Fatalf("send: %d %s", res.StatusCode, body)
	}
	var out struct {
		Answers []string `json:"answers"`
		Rcode   string   `json:"rcode"`
	}
	decode(t, body, &out)
	if len(out.Answers) != 1 || out.Answers[0] != "10.11.12.13" {
		t.Fatalf("expected the app's answer back, got %+v", out)
	}
}

// --- RADIUS send (request-response): engine authenticates against our server ---
func TestSendRADIUS(t *testing.T) {
	e := startEngine(t)
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := radius.PacketServer{
		SecretSource: radius.StaticSecretSource([]byte("throwaway-radius")),
		Handler: radius.HandlerFunc(func(w radius.ResponseWriter, r *radius.Request) {
			w.Write(r.Response(radius.CodeAccessReject)) // the app rejects
		}),
	}
	go srv.Serve(pc)
	defer srv.Shutdown(context.Background())
	time.Sleep(50 * time.Millisecond)

	res, body := e.do(t, "POST", "/api/v1/send/radius", map[string]any{
		"target": pc.LocalAddr().String(), "secret": "throwaway-radius", "username": "bob",
	})
	if res.StatusCode != 200 {
		t.Fatalf("send: %d %s", res.StatusCode, body)
	}
	var out struct {
		Code string `json:"code"`
	}
	decode(t, body, &out)
	if out.Code != "Access-Reject" {
		t.Fatalf("expected the app's reply (Access-Reject), got %q", out.Code)
	}
}

// --- unlicensed sender is refused ---
func TestSendUnlicensed(t *testing.T) {
	e := startEngineFree(t) // free mode: only http licensed, no syslog sender
	res, _ := e.do(t, "POST", "/api/v1/send/syslog", map[string]any{"target": "127.0.0.1:514", "message": "x"})
	if res.StatusCode != 404 {
		t.Fatalf("unlicensed sender must be 404, got %d", res.StatusCode)
	}
}

func startEngineFree(t *testing.T) *engine {
	// A licensed set without any sender-capable stand beyond files: use http only.
	return startEngineWithStands(t, []string{"http"})
}

// --- NTP send: engine queries our NTP server, returns stratum ---
func TestSendNTP(t *testing.T) {
	e := startEngine(t)
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer pc.Close()
	go func() {
		buf := make([]byte, 48)
		for {
			n, addr, err := pc.ReadFrom(buf)
			if err != nil {
				return
			}
			if n < 48 {
				continue
			}
			resp := make([]byte, 48)
			resp[0] = 0x24
			resp[1] = 7 // stratum 7
			pc.WriteTo(resp, addr)
		}
	}()
	res, body := e.do(t, "POST", "/api/v1/send/ntp", map[string]any{"target": pc.LocalAddr().String()})
	if res.StatusCode != 200 {
		t.Fatalf("send: %d %s", res.StatusCode, body)
	}
	var out struct {
		Stratum int `json:"stratum"`
	}
	decode(t, body, &out)
	if out.Stratum != 7 {
		t.Fatalf("expected the app's stratum 7, got %d", out.Stratum)
	}
}

// --- SNMP trap send (one-way): our trap listener receives it ---
func TestSendSNMPTrap(t *testing.T) {
	e := startEngine(t)
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer pc.Close()
	got := make(chan bool, 1)
	go func() {
		buf := make([]byte, 2048)
		if _, _, err := pc.ReadFrom(buf); err == nil {
			got <- true
		}
	}()
	res, body := e.do(t, "POST", "/api/v1/send/snmp-traps", map[string]any{
		"target": pc.LocalAddr().String(), "community": "throwaway-public", "value": "boom",
	})
	if res.StatusCode != 200 {
		t.Fatalf("send: %d %s", res.StatusCode, body)
	}
	select {
	case <-got:
	case <-time.After(2 * time.Second):
		t.Fatal("trap listener did not receive the trap")
	}
}

// --- TACACS+ send (request-response): our server replies PASS ---
func TestSendTACACS(t *testing.T) {
	e := startEngine(t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go tacacsServer(ln, "throwaway-tacacs", 0x01) // reply PASS
	res, body := e.do(t, "POST", "/api/v1/send/tacacs", map[string]any{
		"target": ln.Addr().String(), "secret": "throwaway-tacacs", "username": "carol",
	})
	if res.StatusCode != 200 {
		t.Fatalf("send: %d %s", res.StatusCode, body)
	}
	var out struct {
		Status string `json:"status"`
	}
	decode(t, body, &out)
	if out.Status != "pass" {
		t.Fatalf("expected the app's status 'pass', got %q", out.Status)
	}
}

// --- Kafka send (one-way): produce to a broker; here the harness's own broker ---
func TestSendKafka(t *testing.T) {
	e := startEngine(t)
	// Use the engine's own kafka broker as the "application" and confirm the produce succeeds.
	res, body := e.do(t, "POST", "/api/v1/send/kafka", map[string]any{
		"target": e.endpoints["kafka"], "topic": "servienta", "key": "k", "value": "v",
	})
	if res.StatusCode != 200 {
		t.Fatalf("send: %d %s", res.StatusCode, body)
	}
	var out struct {
		Sent bool `json:"sent"`
	}
	decode(t, body, &out)
	if !out.Sent {
		t.Fatal("kafka produce not confirmed")
	}
}

// --- IPFIX send (one-way): export to our collector via the engine's collector ---
func TestSendIPFIX(t *testing.T) {
	e := startEngine(t)
	res, body := e.do(t, "POST", "/api/v1/send/ipfix", map[string]any{
		"target": e.endpoints["ipfix"], "src": "9.9.9.9", "dst": "8.8.8.8",
	})
	if res.StatusCode != 200 {
		t.Fatalf("send: %d %s", res.StatusCode, body)
	}
	var out struct {
		Sent bool `json:"sent"`
	}
	decode(t, body, &out)
	if !out.Sent {
		t.Fatal("ipfix export not confirmed")
	}
}

func tacacsServer(ln net.Listener, secret string, status byte) {
	conn, err := ln.Accept()
	if err != nil {
		return
	}
	defer conn.Close()
	var hdr [12]byte
	if _, err := readFull(conn, hdr[:]); err != nil {
		return
	}
	blen := uint32(hdr[8])<<24 | uint32(hdr[9])<<16 | uint32(hdr[10])<<8 | uint32(hdr[11])
	body := make([]byte, blen)
	readFull(conn, body)
	sessionID := hdr[4:8]
	// reply body: status, flags, msglen(2), datalen(2)
	rbody := []byte{status, 0, 0, 0, 0, 0}
	var rhdr [12]byte
	rhdr[0] = hdr[0]
	rhdr[1] = 0x01
	rhdr[2] = 0x02 // seq 2
	rhdr[3] = 0x00
	copy(rhdr[4:8], sessionID)
	rhdr[11] = byte(len(rbody))
	pad(rbody, []byte(secret), sessionID, rhdr[0], rhdr[2])
	conn.Write(rhdr[:])
	conn.Write(rbody)
}

func readFull(c net.Conn, b []byte) (int, error) {
	total := 0
	for total < len(b) {
		n, err := c.Read(b[total:])
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}
