---
title: Receivers
type: guide
updated: 2026-09-01
---

# Receivers

Each receiver accepts one protocol and records what arrived under its service
name. Read messages back with `GET /api/v1/received/<service>?run=<id>`. All
service receivers except the demo one require a license (see
[Licensing](licensing.md)).

Every recorded message has the shape:
```json
{ "ts": "...", "source": "<ip>", "run": "<id>", "content": { ...parsed... } }
```

## reference — demo receiver (free)
A line-based TCP listener. Send a line, read it back. Used in the quickstart and
as the worked example for [adding a receiver](extending.md).
```bash
echo 'hello' | nc <host> 9000
# content: { "line": "hello" }
```

## syslog — service `syslog` (UDP, TCP, RELP)
One service, three transports (`syslog-udp`, `syslog-tcp`, `syslog-relp`). The
priority is parsed into facility and severity.
```bash
# UDP
printf '<34>hello' | nc -u -w1 <host> 5514
# content: { "transport":"udp", "raw":"<34>hello", "facility":4, "severity":2, "message":"hello" }
```

## snmp-traps — service `snmp-traps`
SNMP traps: v2c by community, and v3 USM across MD5/SHA × DES/AES-128. Each trap
is recorded with its variable bindings.
```
content: { "version":"2c", "community":"...", "variables": { "<oid>": <value>, ... } }
```
Default community and the four USM users are throwaway and documented in
[Configuration](configuration.md).

## radius / tacacs / dns / ntp — request-response
These reply, and you steer the reply. Each records the request too. See
[Response control](response-control.md).
- **radius** — authentication; records `{username}`
- **tacacs** — authentication; records `{username}`
- **dns** — name resolution; records `{qname}`
- **ntp** — time sync; records `{request:true}`

## kafka — service `kafka`
An in-process broker. Produce to the `servienta` topic; the engine records each
message.
```
content: { "topic":"servienta", "key":"...", "value":"..." }
```

## ipfix — service `ipfix`
An IPFIX collector. Export a template then data records; each data record's
fields are recorded by element name.
```
content: { "obs_domain_id": 1, "fields": { "sourceIPv4Address":"1.2.3.4", ... } }
```

## Reading back

```bash
curl 'http://localhost:8080/api/v1/received/syslog?run=run-1'
```
- Omit `?run=` to read everything (attributed or not).
- An unknown service is `404`, not an empty list (a stopped or unlicensed
  receiver is distinguishable from "running but empty").
