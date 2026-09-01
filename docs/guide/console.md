---
title: Console
type: guide
updated: 2026-09-01
---

# Console

The console is the management UI for a running engine. It talks to the engine
only through the versioned control API — it is that contract's first client.

## Running it

The root `docker-compose.yml` runs the engine and console together:
```bash
docker compose up -d
# console on http://localhost:5000
```
The engine is **not published to the host** — only the console reaches it, over
the internal network. A shared volume holds the license.

## What you can do

- **See the license** — mode (free/licensed), customer, expiry, enabled stands.
- **Apply / remove a license** — paste an issued license; the engine restarts to
  apply it. See [Licensing](licensing.md).
- **Stand endpoints** — the live address of every running service.
- **Received messages** — read what a receiver recorded.
- **Reset** — clear messages, faults, and response controls across the stand.

## Configuration

| Env var | Default | Purpose |
| --- | --- | --- |
| `CONSOLE_ADDR` | `:5000` | Listen address |
| `CONSOLE_ENGINE_BASE` | `http://engine:8080` | Engine control API base |
| `CONSOLE_LICENSE_PATH` | `/license/license.json` | Shared-volume license path |
