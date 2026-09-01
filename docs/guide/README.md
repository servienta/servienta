---
title: Servienta — User Guide
type: guide
status: active
updated: 2026-09-01
---

# Servienta user guide

How to use the whole system: run the engine, drive it from your tests, inject
faults, control replies, apply a license, and operate the console and admin.

New to Servienta? Read [What Servienta is](#what-servienta-is), then jump to the
[Quickstart](quickstart.md).

## What Servienta is

Servienta gives your test suite a **controlled instance of the network services
your application talks to** — a file server on five transports and receivers for
syslog, SNMP traps, RADIUS, TACACS+, DNS, NTP, Kafka, and IPFIX. Your tests
bring it up with one command, generate traffic, read back exactly what arrived,
inject faults, and reset — all reproducibly, with no human in the loop.

The **application under test is out of scope**: the engine never reaches into
it. It only accepts, serves, and reports what it observed.

## The pieces

| Piece | What it is | Where it runs |
| --- | --- | --- |
| **Engine** | The data plane: receivers, file server, control API | Your infrastructure (Docker), offline-capable |
| **Console** | Management UI: view the stand, apply a license, reset | Docker on :5000, next to the engine |
| **Admin** | Vendor back office: issue licenses, manage customers | Cloudflare (vendor-operated) |

## Guide contents

| Page | Covers |
| --- | --- |
| [Quickstart](quickstart.md) | Run the engine, record and read your first message |
| [Engine API reference](engine-api.md) | Every control endpoint |
| [Receivers](receivers.md) | Each service: how to send to it and read it back |
| [Response control](response-control.md) | Steering RADIUS/TACACS+/DNS/NTP replies |
| [Fault injection](faults.md) | Forcing file and receiver failures |
| [Licensing](licensing.md) | Free vs licensed, applying a license |
| [Console](console.md) | Operating the management UI |
| [Configuration](configuration.md) | Every environment variable |
| [Extending](extending.md) | Adding a new receiver |

## Core concepts

- **Instance** — one bring-up of the engine (one compose project). Mutations
  (reset, faults, response controls) are instance-wide.
- **Run** — one logical test execution inside an instance, identified by a
  **run id** you choose. Recorded messages are read back per run.
- **Source claim** — a run declares the source addresses whose traffic belongs
  to it; the engine attributes each message to a run by source. This is what
  lets two test runs share one instance without reading each other's traffic.
- **Stand** — one licensable service (a file transport or a receiver). A license
  enables a set of stands; the free image runs the file server (HTTP) and a demo
  receiver.
