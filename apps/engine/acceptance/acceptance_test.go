// The acceptance suite (docs/acceptance-suite.md): talks to the engine only
// through public interfaces — the versioned HTTP contract and network
// listeners. Requirement mapping is in the doc; a requirement is closed only
// when its test here passes.
package acceptance

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/servienta/servienta/apps/engine/internal/app"
)

type engine struct {
	control string // base URL
	files   string
	refAddr string
	close   func()
}

func startEngine(t *testing.T) *engine {
	t.Helper()
	fixtures := t.TempDir()
	writeFixtures(t, fixtures)
	ctx, cancel := context.WithCancel(context.Background())
	a, err := app.Start(ctx, app.Config{
		ControlAddr:   "127.0.0.1:0",
		FilesHTTPAddr: "127.0.0.1:0",
		ReferenceAddr: "127.0.0.1:0",
		FixturesDir:   fixtures,
	})
	if err != nil {
		cancel()
		t.Fatalf("start: %v", err)
	}
	e := &engine{
		control: "http://" + a.Endpoints["control"],
		files:   "http://" + a.Endpoints["files-http"],
		refAddr: a.Endpoints["reference"],
		close:   func() { a.Close(); cancel() },
	}
	t.Cleanup(e.close)
	return e
}

var fixtureData = map[string][]byte{}

func writeFixtures(t *testing.T, dir string) {
	t.Helper()
	if len(fixtureData) == 0 {
		big := make([]byte, 8<<20) // scaled-down large.bin for CI; N6-sized run is a flag away
		rand.Read(big)
		fixtureData["small.txt"] = []byte("hello, servienta\n")
		fixtureData["large.bin"] = big
		fixtureData["malformed.bin"] = []byte("PK\x03\x04 truncated-on-purpose")
	}
	for name, data := range fixtureData {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func (e *engine) do(t *testing.T, method, path string, body any) (*http.Response, []byte) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, e.control+path, &buf)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer res.Body.Close()
	var out bytes.Buffer
	out.ReadFrom(res.Body)
	return res, out.Bytes()
}

func (e *engine) sendLines(t *testing.T, fromIP string, lines ...string) {
	t.Helper()
	d := net.Dialer{LocalAddr: &net.TCPAddr{IP: net.ParseIP(fromIP)}}
	conn, err := d.Dial("tcp", e.refAddr)
	if err != nil {
		t.Fatalf("dial from %s: %v", fromIP, err)
	}
	defer conn.Close()
	for _, l := range lines {
		fmt.Fprintln(conn, l)
	}
}

func (e *engine) received(t *testing.T, service, run string) []map[string]any {
	t.Helper()
	path := "/api/v1/received/" + service
	if run != "" {
		path += "?run=" + run
	}
	deadline := time.Now().Add(3 * time.Second)
	for {
		res, body := e.do(t, "GET", path, nil)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("received %s: %d %s", path, res.StatusCode, body)
		}
		var msgs []map[string]any
		json.Unmarshal(body, &msgs)
		if len(msgs) > 0 || time.Now().After(deadline) {
			return msgs
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// --- R11, criterion 5 ---
func TestVersion(t *testing.T) {
	e := startEngine(t)
	res, body := e.do(t, "GET", "/api/v1/version", nil)
	if res.StatusCode != 200 || !bytes.Contains(body, []byte(`"contract"`)) {
		t.Fatalf("version: %d %s", res.StatusCode, body)
	}
}

// --- R1 (HTTP), R6, criteria 1–2 ---
func TestFixtureByteCompareHTTP(t *testing.T) {
	e := startEngine(t)
	for name, want := range fixtureData {
		res, err := http.Get(e.files + "/" + name)
		if err != nil {
			t.Fatal(err)
		}
		var got bytes.Buffer
		got.ReadFrom(res.Body)
		res.Body.Close()
		if !bytes.Equal(got.Bytes(), want) {
			t.Fatalf("%s: served bytes differ from original", name)
		}
	}
}

// --- R10, R4, criterion 6 ---
func TestReferenceReceiverRoundTrip(t *testing.T) {
	e := startEngine(t)
	if res, body := e.do(t, "PUT", "/api/v1/runs/run-a", map[string]any{"sources": []string{"127.0.0.1"}}); res.StatusCode != 201 {
		t.Fatalf("declare: %d %s", res.StatusCode, body)
	}
	e.sendLines(t, "127.0.0.1", "alpha", "beta")
	msgs := e.received(t, "reference", "run-a")
	if len(msgs) != 2 {
		t.Fatalf("want 2 messages, got %d", len(msgs))
	}
	if c := msgs[0]["content"].(map[string]any); c["line"] != "alpha" {
		t.Fatalf("parsed content mismatch: %v", c)
	}
}

// --- R4, N4, criterion 7: two concurrent runs, no cross-reads ---
func TestRunIsolation(t *testing.T) {
	e := startEngine(t)
	e.do(t, "PUT", "/api/v1/runs/run-a", map[string]any{"sources": []string{"127.0.0.1"}})
	e.do(t, "PUT", "/api/v1/runs/run-b", map[string]any{"sources": []string{"127.0.0.2"}})
	e.sendLines(t, "127.0.0.1", "from-a")
	e.sendLines(t, "127.0.0.2", "from-b")
	a := e.received(t, "reference", "run-a")
	b := e.received(t, "reference", "run-b")
	if len(a) != 1 || len(b) != 1 {
		t.Fatalf("isolation: a=%d b=%d", len(a), len(b))
	}
	if a[0]["content"].(map[string]any)["line"] != "from-a" || b[0]["content"].(map[string]any)["line"] != "from-b" {
		t.Fatalf("cross-read: a=%v b=%v", a[0], b[0])
	}
}

// --- D4: one source, one active run ---
func TestClaimConflict(t *testing.T) {
	e := startEngine(t)
	e.do(t, "PUT", "/api/v1/runs/run-a", map[string]any{"sources": []string{"127.0.0.1"}})
	res, _ := e.do(t, "PUT", "/api/v1/runs/run-b", map[string]any{"sources": []string{"127.0.0.1"}})
	if res.StatusCode != http.StatusConflict {
		t.Fatalf("want 409, got %d", res.StatusCode)
	}
}

// --- N5, R4: empty ≠ unknown receiver ≠ undeclared run ---
func TestEmptyVsUnknown(t *testing.T) {
	e := startEngine(t)
	if res, _ := e.do(t, "GET", "/api/v1/received/reference", nil); res.StatusCode != 200 {
		t.Fatalf("running+empty must be 200 [], got %d", res.StatusCode)
	}
	if res, _ := e.do(t, "GET", "/api/v1/received/nonexistent", nil); res.StatusCode != 404 {
		t.Fatalf("unknown receiver must be 404, got %d", res.StatusCode)
	}
	if res, _ := e.do(t, "GET", "/api/v1/received/reference?run=ghost", nil); res.StatusCode != 404 {
		t.Fatalf("undeclared run must be 404, got %d", res.StatusCode)
	}
}

// --- R5, criterion 3: everything above passes twice with only POST /reset between ---
func TestResetTwice(t *testing.T) {
	e := startEngine(t)
	pass := func() {
		if res, _ := e.do(t, "PUT", "/api/v1/runs/run-a", map[string]any{"sources": []string{"127.0.0.1"}}); res.StatusCode != 201 {
			t.Fatalf("declare after reset must succeed: %d", res.StatusCode)
		}
		e.sendLines(t, "127.0.0.1", "ping")
		if msgs := e.received(t, "reference", "run-a"); len(msgs) != 1 {
			t.Fatalf("want exactly 1 message, got %d", len(msgs))
		}
	}
	pass()
	if res, _ := e.do(t, "POST", "/api/v1/reset", nil); res.StatusCode != http.StatusNoContent {
		t.Fatalf("reset: %d", res.StatusCode)
	}
	pass() // same run id redeclared, same assertions — nothing survived the reset
}
