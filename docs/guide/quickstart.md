---
title: Quickstart
type: guide
updated: 2026-09-01
---

# Quickstart

Run the engine, serve a file, record a message, read it back, reset. No
dependencies, works offline. The free image needs no license.

## 1. Run the engine

```bash
docker pull ghcr.io/servienta/engine:latest

mkdir -p fixtures && echo hello > fixtures/hello.txt

docker run --rm \
  -p 8080:8080 -p 8081:8081 -p 9000:9000 \
  -v "$PWD/fixtures:/fixtures:ro" \
  ghcr.io/servienta/engine:latest
```

- `8080` — the control API (always on)
- `8081` — the HTTP file server (free)
- `9000` — the demo "reference" receiver (free)

## 2. Discover the endpoints

```bash
curl http://localhost:8080/api/v1/endpoints
# {"control":"...","files-http":"...","reference":"..."}
```

In-container ports are fixed; **host ports are yours to map**. `/endpoints`
reports the in-container address of every running service.

## 3. Serve a fixture byte-for-byte

```bash
curl http://localhost:8081/hello.txt   # -> hello
```

The engine serves your fixture tree exactly as-is: no parsing, no encoding
assumptions, text and binary alike, large files streamed.

## 4. Record and read a message

Declare a run, claiming the source address your traffic will come from, send a
message, and read it back:

```bash
# declare a run; claim the docker gateway as this run's source
curl -X PUT http://localhost:8080/api/v1/runs/run-1 \
  -d '{"sources": ["172.17.0.1"]}'

# send a line to the demo receiver
echo 'hello from my app' | nc localhost 9000

# read back what this run received, parsed
curl 'http://localhost:8080/api/v1/received/reference?run=run-1'
# [{"ts":"...","source":"172.17.0.1","run":"run-1","content":{"line":"hello from my app"}}]
```

> **Why the source address?** Your application can't put a run id inside a raw
> syslog packet, so a run instead *claims the source addresses* its traffic
> comes from. The engine attributes each message to a run by source. One source
> belongs to one active run — a conflict is an explicit error. This is what
> keeps two concurrent runs from reading each other's traffic.

## 5. Reset to a known state

```bash
curl -X POST http://localhost:8080/api/v1/reset
```

Reset clears every receiver's messages, lifts all faults and response controls,
releases run declarations, and restores the served tree. Run your suite twice
with only a reset between — it passes both times.

## Next

- Unlock the service receivers (syslog, SNMP, …) → [Licensing](licensing.md)
- Send to and read every service → [Receivers](receivers.md)
- Steer replies and inject failures → [Response control](response-control.md),
  [Fault injection](faults.md)
