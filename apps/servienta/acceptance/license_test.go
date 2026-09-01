package acceptance

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/servienta/servienta/apps/servienta/internal/app"
)

// startWithLicense boots an engine that resolves a real mounted license file
// against a given public key — exercising the full R12 startup path.
func startWithLicense(t *testing.T, licenseJSON []byte, pubKeyB64 string) *engine {
	t.Helper()
	fixtures := t.TempDir()
	writeFixtures(t, fixtures)
	licPath := ""
	if licenseJSON != nil {
		licPath = filepath.Join(t.TempDir(), "license.json")
		if err := os.WriteFile(licPath, licenseJSON, 0o644); err != nil {
			t.Fatal(err)
		}
	} else {
		licPath = filepath.Join(t.TempDir(), "absent.json") // missing => Free mode
	}
	ctx, cancel := context.WithCancel(context.Background())
	a, err := app.Start(ctx, app.Config{
		ControlAddr:    "127.0.0.1:0",
		FilesHTTPAddr:  "127.0.0.1:0",
		FilesHTTPSAddr: "127.0.0.1:0",
		FilesFTPAddr:   "127.0.0.1:0",
		FilesTFTPAddr:  "127.0.0.1:0",
		FilesSCPAddr:   "127.0.0.1:0",
		FilesUser:      filesUser,
		FilesPassword:  filesPassword,
		ReferenceAddr:  "127.0.0.1:0",
		SyslogUDPAddr:  "127.0.0.1:0",
		SyslogTCPAddr:  "127.0.0.1:0",
		SyslogRELPAddr: "127.0.0.1:0",
		SNMPTrapAddr:   "127.0.0.1:0",
		FixturesDir:    fixtures,
		LicensePath:    licPath,
		LicensePubKey:  pubKeyB64,
		// LicensedStands intentionally nil: force real license resolution.
	})
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	e := &engine{control: "http://" + a.Endpoints["control"], endpoints: a.Endpoints, close: func() { a.Close(); cancel() }}
	if fh, ok := a.Endpoints["files-http"]; ok {
		e.files = "http://" + fh
	}
	t.Cleanup(e.close)
	return e
}

func mkLicense(t *testing.T, priv ed25519.PrivateKey, stands []string, exp time.Time) []byte {
	t.Helper()
	payload, _ := json.Marshal(map[string]any{"v": 1, "name": "Acme", "stands": stands, "exp": exp.UnixMilli()})
	f := map[string]string{
		"payload_b64": base64.StdEncoding.EncodeToString(payload),
		"signature":   base64.StdEncoding.EncodeToString(ed25519.Sign(priv, payload)),
	}
	b, _ := json.Marshal(f)
	return b
}

func spkiB64(t *testing.T, pub ed25519.PublicKey) string {
	t.Helper()
	der, _ := x509.MarshalPKIXPublicKey(pub)
	return base64.StdEncoding.EncodeToString(der)
}

func (e *engine) licenseStatus(t *testing.T) map[string]any {
	t.Helper()
	res, body := e.do(t, "GET", "/api/v1/license", nil)
	if res.StatusCode != 200 {
		t.Fatalf("license status: %d", res.StatusCode)
	}
	var m map[string]any
	json.Unmarshal(body, &m)
	return m
}

// --- R12: a license grants exactly its stands; unlicensed ones do not start ---
func TestLicenseGatesStands(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	lic := mkLicense(t, priv, []string{"http", "syslog"}, time.Now().Add(time.Hour))
	e := startWithLicense(t, lic, spkiB64(t, pub))

	st := e.licenseStatus(t)
	if st["mode"] != "licensed" {
		t.Fatalf("mode: %v", st["mode"])
	}
	// http licensed -> file server up; https NOT licensed -> absent
	if _, ok := e.endpoints["files-http"]; !ok {
		t.Fatal("http licensed but files-http not started")
	}
	if _, ok := e.endpoints["files-https"]; ok {
		t.Fatal("https not licensed but files-https started")
	}
	if _, ok := e.endpoints["snmp-traps"]; ok {
		t.Fatal("snmp-traps not licensed but started")
	}
	if _, ok := e.endpoints["syslog-udp"]; !ok {
		t.Fatal("syslog licensed but not started")
	}
}

// --- R12: no license file => Free mode, http only ---
func TestFreeMode(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	e := startWithLicense(t, nil, spkiB64(t, pub))
	st := e.licenseStatus(t)
	if st["mode"] != "free" {
		t.Fatalf("want free mode, got %v", st["mode"])
	}
	if _, ok := e.endpoints["files-http"]; !ok {
		t.Fatal("free mode must serve http")
	}
	if _, ok := e.endpoints["files-ftp"]; ok {
		t.Fatal("free mode must not serve ftp")
	}
}

// --- R12: expired license is refused explicitly, engine falls back to Free ---
func TestExpiredLicenseRefused(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	lic := mkLicense(t, priv, []string{"http", "ftp"}, time.Now().Add(-time.Hour))
	e := startWithLicense(t, lic, spkiB64(t, pub))
	st := e.licenseStatus(t)
	if st["mode"] != "free" || st["error"] == nil {
		t.Fatalf("expired license must be refused with an error, got %v", st)
	}
	if _, ok := e.endpoints["files-ftp"]; ok {
		t.Fatal("expired license must not grant ftp")
	}
}

var _ = http.StatusOK
