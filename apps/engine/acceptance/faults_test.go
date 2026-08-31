package acceptance

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

// --- R2: each file fault yields a protocol-level error, not wrong content ---
func TestFileFaults(t *testing.T) {
	e := startEngine(t)

	setFault := func(fixture, kind string) {
		if res, body := e.do(t, "PUT", "/api/v1/faults/files/"+fixture, map[string]string{"kind": kind}); res.StatusCode != 204 {
			t.Fatalf("set %s fault: %d %s", kind, res.StatusCode, body)
		}
	}

	setFault("small.txt", "auth-reject")
	if code := httpGetStatus(t, e.files+"/small.txt"); code != http.StatusUnauthorized {
		t.Fatalf("auth-reject: want 401, got %d", code)
	}

	setFault("small.txt", "missing")
	if code := httpGetStatus(t, e.files+"/small.txt"); code != http.StatusNotFound {
		t.Fatalf("missing: want 404, got %d", code)
	}

	setFault("small.txt", "corrupt")
	body := httpGetBody(t, e.files+"/small.txt")
	if bytes.Equal(body, fixtureData["small.txt"]) {
		t.Fatal("corrupt: content matches original — fault not applied")
	}

	setFault("large.bin", "truncate")
	// A truncated transfer must fail as a short read, not deliver wrong bytes cleanly.
	res, err := http.Get(e.files + "/large.bin")
	if err != nil {
		return // connection error is an acceptable protocol-level failure
	}
	defer res.Body.Close()
	_, err = io.Copy(io.Discard, res.Body)
	if err == nil {
		t.Fatal("truncate: full body read succeeded — transfer was not cut")
	}
}

// --- R9: receiver modes; reset lifts them (R5) ---
func TestReceiverModes(t *testing.T) {
	e := startEngine(t)
	e.do(t, "PUT", "/api/v1/runs/run-1", map[string]any{"sources": []string{"127.0.0.1"}})

	setMode := func(mode string, delay int) {
		if res, _ := e.do(t, "PUT", "/api/v1/faults/receivers/reference", map[string]any{"mode": mode, "delay_ms": delay}); res.StatusCode != 204 {
			t.Fatalf("set mode %s failed", mode)
		}
	}

	// drop: connection accepted, nothing recorded
	setMode("drop", 0)
	e.sendLines(t, "127.0.0.1", "into-the-void")
	time.Sleep(150 * time.Millisecond)
	if msgs := e.receivedNoWait(t, "reference", "run-1"); len(msgs) != 0 {
		t.Fatalf("drop: expected 0 messages, got %d", len(msgs))
	}

	// reset lifts the mode (R5); normal recording resumes
	if res, _ := e.do(t, "POST", "/api/v1/reset", nil); res.StatusCode != 204 {
		t.Fatal("reset failed")
	}
	e.do(t, "PUT", "/api/v1/runs/run-1", map[string]any{"sources": []string{"127.0.0.1"}})
	e.sendLines(t, "127.0.0.1", "after-reset")
	if msgs := e.received(t, "reference", "run-1"); len(msgs) != 1 {
		t.Fatalf("after reset: expected 1 message, got %d", len(msgs))
	}
}

func httpGetStatus(t *testing.T, url string) int {
	t.Helper()
	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	return res.StatusCode
}

func httpGetBody(t *testing.T, url string) []byte {
	t.Helper()
	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var buf bytes.Buffer
	buf.ReadFrom(res.Body)
	return buf.Bytes()
}

var _ = net.Dial
