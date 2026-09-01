package license

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

func signed(t *testing.T, priv ed25519.PrivateKey, p Payload) File {
	t.Helper()
	b, _ := json.Marshal(p)
	return File{
		PayloadB64: base64.StdEncoding.EncodeToString(b),
		Signature:  base64.StdEncoding.EncodeToString(ed25519.Sign(priv, b)),
	}
}

func spki(t *testing.T, pub ed25519.PublicKey) string {
	t.Helper()
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(der)
}

func TestVerifyValid(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	f := signed(t, priv, Payload{V: 1, Name: "Acme", Stands: []string{"http", "syslog"}, Exp: time.Now().Add(time.Hour).UnixMilli()})
	p, err := Verify(f, spki(t, pub), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "Acme" || len(p.Stands) != 2 {
		t.Fatalf("payload wrong: %+v", p)
	}
}

func TestVerifyExpired(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	f := signed(t, priv, Payload{Exp: time.Now().Add(-time.Hour).UnixMilli()})
	if _, err := Verify(f, spki(t, pub), time.Now()); err != ErrExpired {
		t.Fatalf("want ErrExpired, got %v", err)
	}
}

func TestVerifyTampered(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	f := signed(t, priv, Payload{Stands: []string{"http"}, Exp: time.Now().Add(time.Hour).UnixMilli()})
	// escalate: swap the payload for one granting everything, keep the old signature
	tampered, _ := json.Marshal(Payload{Stands: []string{"http", "https", "ftp", "kafka"}, Exp: time.Now().Add(time.Hour).UnixMilli()})
	f.PayloadB64 = base64.StdEncoding.EncodeToString(tampered)
	if _, err := Verify(f, spki(t, pub), time.Now()); err != ErrBadSignature {
		t.Fatalf("want ErrBadSignature, got %v", err)
	}
}

func TestVerifyWrongKey(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	otherPub, _, _ := ed25519.GenerateKey(rand.Reader)
	f := signed(t, priv, Payload{Stands: []string{"http"}, Exp: time.Now().Add(time.Hour).UnixMilli()})
	if _, err := Verify(f, spki(t, otherPub), time.Now()); err != ErrBadSignature {
		t.Fatalf("want ErrBadSignature, got %v", err)
	}
}
