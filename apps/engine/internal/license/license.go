// Package license verifies the engine's offline license (R12, D10, D15):
// an Ed25519-signed payload enumerating the licensed stands. The engine
// embeds only the public key; the signing key never ships. Verification is
// fully offline (N2).
package license

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// File is what the admin issues and the customer mounts (base64 payload +
// Ed25519 signature over the exact payload bytes).
type File struct {
	PayloadB64 string `json:"payload_b64"`
	Signature  string `json:"signature"`
}

// Payload is the signed grant. Times are epoch millis UTC (D9).
type Payload struct {
	V      int      `json:"v"`
	JTI    string   `json:"jti"`
	Sub    string   `json:"sub"`
	Name   string   `json:"name"`
	Stands []string `json:"stands"`
	IAT    int64    `json:"iat"`
	Exp    int64    `json:"exp"`
}

var (
	ErrBadSignature = errors.New("license signature does not verify")
	ErrExpired      = errors.New("license has expired")
	ErrPublicKey    = errors.New("engine has no valid license public key")
)

// Verify checks the signature against pubKeyB64 (SPKI or raw, base64) and the
// expiry, returning the grant. The payload bytes are verified as-is, so
// canonicalization never matters (D10).
func Verify(f File, pubKeyB64 string, now time.Time) (*Payload, error) {
	pub, err := parsePublicKey(pubKeyB64)
	if err != nil {
		return nil, err
	}
	payloadBytes, err := base64.StdEncoding.DecodeString(f.PayloadB64)
	if err != nil {
		return nil, fmt.Errorf("payload: %w", err)
	}
	sig, err := base64.StdEncoding.DecodeString(f.Signature)
	if err != nil {
		return nil, fmt.Errorf("signature: %w", err)
	}
	if !ed25519.Verify(pub, payloadBytes, sig) {
		return nil, ErrBadSignature
	}
	var p Payload
	if err := json.Unmarshal(payloadBytes, &p); err != nil {
		return nil, fmt.Errorf("payload json: %w", err)
	}
	if now.UnixMilli() > p.Exp {
		return nil, ErrExpired
	}
	return &p, nil
}

// parsePublicKey accepts a raw 32-byte Ed25519 key or a DER SPKI wrapper, both
// base64. The admin's generate-signing-key script emits SPKI.
func parsePublicKey(b64 string) (ed25519.PublicKey, error) {
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, ErrPublicKey
	}
	if len(raw) == ed25519.PublicKeySize {
		return ed25519.PublicKey(raw), nil
	}
	// SPKI DER for Ed25519 is a fixed 12-byte prefix + the 32-byte key.
	if len(raw) == 12+ed25519.PublicKeySize {
		return ed25519.PublicKey(raw[12:]), nil
	}
	return nil, ErrPublicKey
}
