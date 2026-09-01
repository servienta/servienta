---
title: Console
type: guide
updated: 2026-09-01
---

# Console

The console is the management UI, built into the delivered container (D20).
It is a browser client of the versioned control API — the contract's first
client — and can do nothing your test suite cannot do with `curl`.

## Opening it

Run the container ([Quickstart](quickstart.md)) with port `5000` mapped and
open **http://localhost:5000**. No install, no login, works offline (N2) —
the SPA, this guide, and the walkthrough script are embedded in the binary.

The console's own tabs:

- **Stand** — the live address and state of every running service, the
  license card (mode, customer, expiry, granted stands), received messages,
  and reset.
- **Getting started** — the core `curl` loop and a copyable walkthrough
  script.
- **Docs** — this entire user guide, rendered offline.

## What you can do

- **See the license** — mode (free/licensed), customer, expiry, enabled stands.
- **Apply / remove a license** — paste an issued license; Servienta restarts
  in place to apply it. See [Licensing](licensing.md).
- **Stand endpoints** — the live address of every running service.
- **Received messages** — read what a receiver recorded, per run or all.
- **Reset** — clear messages, faults, and response controls across the stand.

## Configuration

The console is served by Servienta itself:

| Env var | Default | Purpose |
| --- | --- | --- |
| `SERVIENTA_UI_ADDR` | `:5000` | Console listen address (empty disables it) |
| `SERVIENTA_CONTROL_ADDR` | `:5001` | Control API the console (and your tests) talk to |

The browser calls the same `/api/v1/…` contract you curl; the console port
answers `/api/*` with the identical handler so no CORS setup is needed.
