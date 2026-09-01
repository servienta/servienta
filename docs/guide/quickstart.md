---
title: Quickstart
type: guide
updated: 2026-09-01
---

# Quickstart

Run Servienta, serve a file, record a message, read it back, reset. No
dependencies, works offline. The free image needs no license.

## 1. Run Servienta

```bash
docker pull ghcr.io/servienta/servienta:latest

mkdir -p fixtures && echo hello > fixtures/hello.txt

docker run --rm \
  -p 5000:5000 -p 5001:5001 -p 8080:8080 -p 9000:9000 \
  -v "$PWD/fixtures:/fixtures:ro" \
  ghcr.io/servienta/servienta:latest
```

- `5000` — the management console, in your browser (always on)
- `5001` — the control API, the versioned contract (always on)
- `8080` — the HTTP file server (free)
- `9000` — the demo "reference" receiver (free)

One container is the whole product: Servienta, web console, and the full user
guide (console → Docs) — no compose, no second service. Licensed services
each listen on their own port; map the ones you use the same 1:1 way.

## 2. Discover the endpoints

```bash
curl http://localhost:5001/api/v1/endpoints
# {"control":"...","files-http":"...","reference":"..."}
```

In-container ports are fixed; **host ports are yours to map**. `/endpoints`
reports the in-container address of every running service.

## 3. Serve a fixture byte-for-byte

```bash
curl http://localhost:8080/hello.txt   # -> hello
```

Servienta serves your fixture tree exactly as-is: no parsing, no encoding
assumptions, text and binary alike, large files streamed.

The tree is **live** — it is your host directory, mounted read-only into the
container at `/fixtures`. Directory structure maps to URL paths one-to-one,
and edits on the host are visible immediately, no restart and no upload step:

```bash
mkdir -p fixtures/configs
printf 'interface eth0\n  link down\n' > fixtures/configs/router.cfg
curl http://localhost:8080/configs/router.cfg   # served at once
```

Every licensed file transport (HTTPS, FTP, TFTP, SCP) serves this same tree.
Servienta never writes to it: fault injection only masks entries, and
`POST /reset` unmasks them — your files are never touched.

## 4. Record and read a message

Declare a run, claiming the source address your traffic will come from, send a
message, and read it back:

```bash
# declare a run; claim the docker gateway as this run's source
curl -X PUT http://localhost:5001/api/v1/runs/run-1 \
  -d '{"sources": ["172.17.0.1"]}'

# send a line to the demo receiver
echo 'hello from my app' | nc localhost 9000

# read back what this run received, parsed
curl 'http://localhost:5001/api/v1/received/reference?run=run-1'
# [{"ts":"...","source":"172.17.0.1","run":"run-1","content":{"line":"hello from my app"}}]
```

> **Why the source address?** Your application can't put a run id inside a raw
> syslog packet, so a run instead *claims the source addresses* its traffic
> comes from. Servienta attributes each message to a run by source. One source
> belongs to one active run — a conflict is an explicit error. This is what
> keeps two concurrent runs from reading each other's traffic.

If the run-scoped read comes back empty, read without `?run=` — the unfiltered
view shows the actual `source` of every message; claim that address. (Traffic
sent from the Docker host typically arrives as the bridge gateway, `172.17.0.1`
for a plain `docker run`.)

## 5. Reset to a known state

```bash
curl -X POST http://localhost:5001/api/v1/reset
```

Reset clears every receiver's messages, lifts all faults and response controls,
releases run declarations, and restores the served tree. Run your suite twice
with only a reset between — it passes both times.

## Run it anywhere

Servienta is a single container with no runtime dependencies and no network
needs of its own. Three variations cover most environments:

**Air-gapped host** — no registry access. Move the image as a file:

```bash
docker save ghcr.io/servienta/servienta:latest -o servienta.tar
# carry the tar over, then on the target host:
docker load -i servienta.tar
```

**No host directory** — only Docker available, nowhere to keep files. Use a
named volume instead of a bind mount:

```bash
docker volume create fixtures
docker run --rm -v fixtures:/fixtures -v "$PWD:/src:ro" alpine cp -r /src/. /fixtures/
docker run --rm -p 5000:5000 -p 5001:5001 -p 8080:8080 -p 9000:9000 \
  -v fixtures:/fixtures:ro ghcr.io/servienta/servienta:latest
```

**Two instances on one host** — e.g. two CI jobs. In-container ports are
fixed; host ports are yours, so remap the second instance:

```bash
docker run --rm -p 15000:5000 -p 15001:5001 -p 18080:8080 -p 19000:9000 \
  -v "$PWD/other-fixtures:/fixtures:ro" ghcr.io/servienta/servienta:latest
```

Each instance is fully independent: its own fixtures, runs, faults, and reset.

## The console and the guide

Open **http://localhost:5000** — the management console ships inside the
container: stand state, license upload, reset, a Getting-started page with a
copyable walkthrough script, and this entire guide under **Docs** (works
offline). The console is a browser client of the same contract you curl —
it can do nothing your test suite cannot.

## Next

- Unlock the service receivers (syslog, SNMP, …) → [Licensing](licensing.md)
- Send to and read every service → [Receivers](receivers.md)
- Steer replies and inject failures → [Response control](response-control.md),
  [Fault injection](faults.md)
