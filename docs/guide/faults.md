---
title: Fault injection (R2, R9)
type: guide
updated: 2026-09-01
---

# Fault injection

Force failures on both sides — file retrieval and receivers — to test your
error paths. Faults are instance-wide and lifted by `POST /api/v1/reset`.

## File faults (R2)

Make a named fixture fail on the file server:
```bash
curl -X PUT http://localhost:5001/api/v1/faults/files/hello.txt -d '{"kind":"<kind>"}'
```

| kind | Effect |
| --- | --- |
| `auth-reject` | Authentication rejected (401) |
| `missing` | File reported absent (404) |
| `truncate` | Transfer cut off midway — a short read, not clean content |
| `corrupt` | Content corrupted — a byte flipped, transfer "succeeds" |

Each yields a **protocol-level error**, never a successful transfer of wrong
content that a test might mistake for success. (Corrupt is the intentional
exception: it delivers altered bytes so your integrity check can catch it.)

## Receiver faults (R9)

Put a receiver into a failure mode:
```bash
curl -X PUT http://localhost:5001/api/v1/faults/receivers/syslog \
  -d '{"mode":"<mode>", "delay_ms": 0}'
```

| mode | Effect |
| --- | --- |
| `refuse` | Refuse connections |
| `drop` | Accept and silently drop (nothing recorded) |
| `delay` | Respond/accept after `delay_ms` |
| `cut` | Cut the connection mid-message |
| `error` | Return a protocol error where the protocol allows one |

Clear a mode: `-d '{"mode":""}'`.

A deliberate failure mode is **distinguishable from "not running"**: Servienta
reports the receiver as alive and in the configured mode.

## Example
```bash
# make the file server drop mid-transfer
curl -X PUT .../faults/files/large.bin -d '{"kind":"truncate"}'
curl http://localhost:8080/large.bin -o out.bin   # your client sees a short read

# make syslog silently drop
curl -X PUT .../faults/receivers/syslog -d '{"mode":"drop"}'
# send traffic; /received/syslog stays empty; your sender still "succeeds"

curl -X POST .../api/v1/reset   # lift everything
```
