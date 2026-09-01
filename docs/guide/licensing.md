---
title: Licensing
type: guide
updated: 2026-09-01
---

# Licensing

A license enables a set of **stands** (file transports and receivers). It is an
Ed25519-signed file, validated by Servienta at startup, fully offline — no
network, no phone-home.

## Free vs licensed

- **Free** — no license file. Servienta runs the **HTTP file server** and the
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

Three ways, all offline:

### A. Through the contract
```bash
curl -X PUT http://localhost:5001/api/v1/license -d @license.json
```
Servienta stores the file and restarts in place; the signature is verified at
that startup — the endpoint grants no signing power. `DELETE /api/v1/license`
reverts to free mode. The stored license lives inside the container: it
survives `docker restart`, and is simply re-applied after a `docker run --rm`
recreation.

### B. From the console
Open the console (http://localhost:5000), **License** card → **Apply a
license**, paste the file, **Apply** — the console makes exactly the call
above. **Remove license** reverts to free.

### C. Mount the file and start
```bash
docker run … -v "$PWD/license.json:/license.json:ro" ghcr.io/servienta/servienta:latest
```
Servienta reads `/license.json` (or `SERVIENTA_LICENSE`) at startup, verifies
signature and expiry, and enables the licensed stands.

## What Servienta checks

- **Signature** — against the embedded public key. A tampered or wrongly-signed
  file is refused.
- **Expiry** — a past `exp` is refused.
- On refusal Servienta runs **free mode and reports the error** in
  `GET /api/v1/license` — never a silent partial start.

## Honest limits

Servienta is source-available and runs offline, so the license mechanism is for
**compliance clarity, not DRM**: it makes "what is enabled" signed and
verifiable, and prevents third-party forgery, but it cannot stop a determined
licensee from patching an open binary. Instance counts are a contract matter,
not a technical lock. Every shipped credential is throwaway and documented as
such.
