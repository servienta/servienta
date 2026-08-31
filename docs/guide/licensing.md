---
title: Licensing
type: guide
updated: 2026-09-01
---

# Licensing

A license enables a set of **stands** (file transports and receivers). It is an
Ed25519-signed file, validated by the engine at startup, fully offline — no
network, no phone-home.

## Free vs licensed

- **Free** — no license file. The engine runs the **HTTP file server** and the
  **demo receiver**. Enough to try Servienta and to verify file imports.
- **Licensed** — the mounted license enables exactly its stands: any of the five
  file transports and any of the service receivers (syslog, snmp-traps, radius,
  tacacs, dns, ntp, kafka, ipfix).

Plans (presets over the stand set): **Files**, **Standard**, **Enterprise**, or
a **Custom** stand selection. See the pricing model for what each includes.

## Getting a license

The vendor issues it in the admin panel (admin.servienta.com): pick the
customer, a plan or a custom stand set, and a term. You receive a license file —
JSON with `payload_b64` and `signature`.

## Applying a license

Two ways, both offline:

### A. Mount the file and start
Put the license next to your compose file and mount it:
```bash
SERVIENTA_LICENSE=./license.json docker compose up -d
```
The engine reads `/license.json` (or `SERVIENTA_LICENSE`) at startup, verifies
signature and expiry, and enables the licensed stands.

### B. From the console
Open the console (http://localhost:8080), **License** card → **Apply a
license**, paste the file, **Apply**. The console writes it to a shared volume
and restarts the engine to pick it up. **Remove license** reverts to free.

## What the engine checks

- **Signature** — against the embedded public key. A tampered or wrongly-signed
  file is refused.
- **Expiry** — a past `exp` is refused.
- On refusal the engine runs **free mode and reports the error** in
  `GET /api/v1/license` — never a silent partial start.

## Honest limits

The engine is source-available and runs offline, so the license mechanism is for
**compliance clarity, not DRM**: it makes "what is enabled" signed and
verifiable, and prevents third-party forgery, but it cannot stop a determined
licensee from patching an open binary. Instance counts are a contract matter,
not a technical lock. Every shipped credential is throwaway and documented as
such.
