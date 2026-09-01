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

The tree is **live** — it is your host directory, mounted read-only into the
container at `/fixtures`. Directory structure maps to URL paths one-to-one,
and edits on the host are visible immediately, no restart and no upload step:

```bash
mkdir -p fixtures/configs
printf 'interface eth0\n  link down\n' > fixtures/configs/router.cfg
curl http://localhost:8081/configs/router.cfg   # served at once
```

Every licensed file transport (HTTPS, FTP, TFTP, SCP) serves this same tree.
The engine never writes to it: fault injection only masks entries, and
`POST /reset` unmasks them — your files are never touched.

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

If the run-scoped read comes back empty, read without `?run=` — the unfiltered
view shows the actual `source` of every message; claim that address. (Traffic
sent from the Docker host typically arrives as the bridge gateway, `172.17.0.1`
for a plain `docker run`.)

## 5. Reset to a known state

```bash
curl -X POST http://localhost:8080/api/v1/reset
```

Reset clears every receiver's messages, lifts all faults and response controls,
releases run declarations, and restores the served tree. Run your suite twice
with only a reset between — it passes both times.

## Run it anywhere

The engine is a single container with no runtime dependencies and no network
needs of its own. Three variations cover most environments:

**Air-gapped host** — no registry access. Move the image as a file:

```bash
docker save ghcr.io/servienta/engine:latest -o servienta-engine.tar
# carry the tar over, then on the target host:
docker load -i servienta-engine.tar
```

**No host directory** — only Docker available, nowhere to keep files. Use a
named volume instead of a bind mount:

```bash
docker volume create fixtures
docker run --rm -v fixtures:/fixtures -v "$PWD:/src:ro" alpine cp -r /src/. /fixtures/
docker run --rm -p 8080:8080 -p 8081:8081 -p 9000:9000 \
  -v fixtures:/fixtures:ro ghcr.io/servienta/engine:latest
```

**Two instances on one host** — e.g. two CI jobs. In-container ports are
fixed; host ports are yours, so remap the second instance:

```bash
docker run --rm -p 18080:8080 -p 18081:8081 -p 19000:9000 \
  -v "$PWD/other-fixtures:/fixtures:ro" ghcr.io/servienta/engine:latest
```

Each instance is fully independent: its own fixtures, runs, faults, and reset.

## The full stand: engine + console

The repository's `docker-compose.yml` brings up the engine together with the
management console:

```bash
docker compose up -d
```

- The engine is published 1:1 — the contract at `localhost:8080`, every
  receiver at the address it reports. One exception: `dns` answers on host
  `15353`, since mDNS owns 5353/udp on most desktops.
- The console is at **http://localhost:5000** — stand state, license upload,
  reset, the full guide, and a Getting-started page with a copyable
  walkthrough script.
- Fixtures live in `./fixtures` next to the compose file; point elsewhere with
  `SERVIENTA_FIXTURES=/path docker compose up -d`.

## Next

- Unlock the service receivers (syslog, SNMP, …) → [Licensing](licensing.md)
- Send to and read every service → [Receivers](receivers.md)
- Steer replies and inject failures → [Response control](response-control.md),
  [Fault injection](faults.md)
