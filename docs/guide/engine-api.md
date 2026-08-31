---
title: Engine API reference
type: guide
updated: 2026-09-01
---

# Engine control API

Base path `/api/v1`, served on the control port (default `:8080`). The machine-
readable schema ships with the delivery as `apps/engine/openapi.yaml`; a client
generated from it passes the acceptance suite. Breaking changes bump the
contract version.

## Discovery

### `GET /api/v1/version`
Returns the contract version.
```json
{ "contract": "0.1.0" }
```

### `GET /api/v1/endpoints`
The in-container address of every running service.
```json
{ "control": "[::]:8080", "files-http": "[::]:8081", "reference": "[::]:9000" }
```

### `GET /api/v1/license`
Current license status and the licensed stands.
```json
{ "mode": "licensed", "stands": ["http","syslog"], "customer": "Acme", "expires_at": 1893456000000 }
```
`mode` is `free` or `licensed`; `error` appears when a mounted license was
rejected (the engine then runs free).

## Runs (R4)

### `PUT /api/v1/runs/{id}`
Declare a run and claim its source addresses.
```bash
curl -X PUT .../api/v1/runs/run-1 -d '{"sources": ["10.0.0.5", "10.0.0.6"]}'
```
- `201` — declared
- `409` — a source is already claimed by another run, or the run already exists

### `DELETE /api/v1/runs/{id}`
Release a run's claims. Recorded messages stay.
- `204` — released; `404` — run not declared

## Reading (R4)

### `GET /api/v1/received/{service}[?run=<id>]`
Recorded messages in arrival order. With `run`, only that run's messages.
```json
[ { "ts": "2026-09-01T...", "source": "10.0.0.5", "run": "run-1", "content": { ... } } ]
```
- `200` — the messages (an empty `[]` means "running, nothing yet")
- `404` — unknown receiver, or an undeclared run id (never a silent empty `200`)

## Reset (R5)

### `POST /api/v1/reset`
Instance-wide: clears every receiver's messages, lifts every fault (R2, R9),
restores response controls (R8) to default, releases run declarations, and
restores the served tree. `204`.

## Response control (R8)

### `PUT /api/v1/responses/{service}`
Steer a request-response service's reply for the run. See
[Response control](response-control.md) for the per-service spec. An empty `{}`
restores the default (successful reply). `204`.

## Fault injection (R2, R9)

### `PUT /api/v1/faults/files/{fixture}`
Force a file fault for a fixture. Body `{"kind": "auth-reject|missing|truncate|corrupt"}`. `204`.

### `PUT /api/v1/faults/receivers/{service}`
Put a receiver into a failure mode. Body `{"mode": "refuse|drop|delay|cut|error", "delay_ms": N}`.
An empty `{"mode": ""}` clears it. `204`.

See [Fault injection](faults.md).

## Lifecycle

### `POST /api/v1/reload`
Restart the engine to re-read a newly mounted license (the console uses this
after writing a license file). The engine still validates the license only at
startup. `202`, then the process exits and its restart policy brings it back.

## All endpoints at a glance

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/api/v1/version` | Contract version |
| GET | `/api/v1/endpoints` | Service addresses |
| GET | `/api/v1/license` | License status |
| PUT | `/api/v1/runs/{id}` | Declare a run |
| DELETE | `/api/v1/runs/{id}` | Release a run |
| GET | `/api/v1/received/{service}` | Read messages |
| POST | `/api/v1/reset` | Reset the instance |
| PUT | `/api/v1/responses/{service}` | Steer a reply (R8) |
| PUT | `/api/v1/faults/files/{fixture}` | File fault (R2) |
| PUT | `/api/v1/faults/receivers/{service}` | Receiver fault (R9) |
| POST | `/api/v1/reload` | Restart to re-read the license |
