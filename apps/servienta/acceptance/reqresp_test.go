package acceptance

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// setResponse configures an R8 control for a service.
func (e *engine) setResponse(t *testing.T, service string, spec map[string]any) {
	t.Helper()
	if res, body := e.do(t, "PUT", "/api/v1/responses/"+service, spec); res.StatusCode != 204 {
		t.Fatalf("set response %s: %d %s", service, res.StatusCode, body)
	}
}

// --- R3.5 + R8: DNS default record, NXDOMAIN, SERVFAIL ---
func TestDNSResponseControl(t *testing.T) {
	e := startEngine(t)
	q := func() *dns.Msg {
		m := new(dns.Msg)
		m.SetQuestion("example.test.", dns.TypeA)
		c := &dns.Client{Timeout: 2 * time.Second}
		r, _, err := c.Exchange(m, e.endpoints["dns"])
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	// default: a valid A record
	if r := q(); r.Rcode != dns.RcodeSuccess || len(r.Answer) == 0 {
		t.Fatalf("default: want an A record, got rcode %d", r.Rcode)
	}
	// NXDOMAIN
	e.setResponse(t, "dns", map[string]any{"outcome": "nxdomain"})
	if r := q(); r.Rcode != dns.RcodeNameError {
		t.Fatalf("nxdomain: want %d, got %d", dns.RcodeNameError, r.Rcode)
	}
	// SERVFAIL
	e.setResponse(t, "dns", map[string]any{"outcome": "servfail"})
	if r := q(); r.Rcode != dns.RcodeServerFailure {
		t.Fatalf("servfail: want %d, got %d", dns.RcodeServerFailure, r.Rcode)
	}
	// request recorded via R4
	e.do(t, "PUT", "/api/v1/runs/r", map[string]any{"sources": []string{"127.0.0.1"}})
}

// --- R3.6 + R8: NTP default time, and an altered stratum ---
func TestNTPResponseControl(t *testing.T) {
	e := startEngine(t)
	query := func() (byte, error) {
		c, err := net.Dial("udp", e.endpoints["ntp"])
		if err != nil {
			return 0, err
		}
		defer c.Close()
		req := make([]byte, 48)
		req[0] = 0x23 // LI=0 VN=4 Mode=3 (client)
		c.SetDeadline(time.Now().Add(2 * time.Second))
		c.Write(req)
		resp := make([]byte, 48)
		n, err := c.Read(resp)
		if err != nil {
			return 0, err
		}
		if n < 48 {
			return 0, fmt.Errorf("short ntp response")
		}
		return resp[1], nil // stratum
	}
	if s, err := query(); err != nil || s != 1 {
		t.Fatalf("default stratum: want 1, got %d (err %v)", s, err)
	}
	e.setResponse(t, "ntp", map[string]any{"outcome": "stratum", "stratum": float64(5)})
	if s, err := query(); err != nil || s != 5 {
		t.Fatalf("altered stratum: want 5, got %d (err %v)", s, err)
	}
}

// --- R3.3 + R8: RADIUS Access-Accept by default, Access-Reject on demand ---
func TestRADIUSResponseControl(t *testing.T) {
	e := startEngine(t)
	auth := func() radius.Code {
		p := radius.New(radius.CodeAccessRequest, []byte("throwaway-radius"))
		rfc2865.UserName_SetString(p, "alice")
		rfc2865.UserPassword_SetString(p, "secret")
		ctx, cancel := timeoutCtx(2 * time.Second)
		defer cancel()
		resp, err := radius.Exchange(ctx, p, e.endpoints["radius"])
		if err != nil {
			t.Fatal(err)
		}
		return resp.Code
	}
	if c := auth(); c != radius.CodeAccessAccept {
		t.Fatalf("default: want Access-Accept, got %v", c)
	}
	e.setResponse(t, "radius", map[string]any{"outcome": "reject", "reason": "account locked"})
	if c := auth(); c != radius.CodeAccessReject {
		t.Fatalf("reject: want Access-Reject, got %v", c)
	}
}

// --- R3.4 + R8: TACACS+ PASS by default, FAIL on demand ---
func TestTACACSResponseControl(t *testing.T) {
	e := startEngine(t)
	if status := tacacsAuth(t, e.endpoints["tacacs"], "throwaway-tacacs", "bob"); status != 0x01 {
		t.Fatalf("default: want PASS(0x01), got 0x%02x", status)
	}
	e.setResponse(t, "tacacs", map[string]any{"outcome": "fail", "reason": "denied"})
	if status := tacacsAuth(t, e.endpoints["tacacs"], "throwaway-tacacs", "bob"); status != 0x02 {
		t.Fatalf("fail: want FAIL(0x02), got 0x%02x", status)
	}
}
