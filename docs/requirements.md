---
title: Test Harness — Requirements
type: requirements
status: active
updated: 2026-08-31
tags: [tooling, testing, docker]
---

# Test Harness — Requirements

The source of truth for scope. Changing a requirement means changing this file plus an entry in
[decisions.md](decisions.md) — not a verbal agreement.

## Problem

The application under test talks to a set of external network services and downloads files from a
file server over five different transports. No controlled instance of any of these endpoints exists
in the development environment or in CI.

The consequence: every screen that touches such an integration is verified by reading code, and
every import check rests on a hand-assembled file server. Neither is reproducible, and neither can
run without a human.

## Scope

The harness is a set of **services and processes**: endpoints that accept data from the application
under test or serve data to it, and a control layer that lets an automated test prepare the
environment, read observations off it, and return it to a known state.

Out of scope: the application under test itself and the system it works with. Both already exist and
are assumed running. The harness never polls them, never configures them, and never depends on them —
it only accepts, serves, and reports what it observed.

These requirements govern the **engine** — the deployable that provides the endpoints. The hosted
console and the marketing site are separate deliverables with their own stack and constraints; see
[stack.md](stack.md).

## Vocabulary

Two units of isolation exist, and every requirement below uses them precisely ([D3](decisions.md)):

- **Instance** — one bring-up of the harness: one container on one network. Mutations of
  shared behavior — reset (R5), fault injection (R2, R9), response controls (R8) — are
  instance-wide.
- **Run** — one logical test execution inside an instance, identified by a caller-chosen **run id**.
  Message recording and reads are run-scoped (R4). Concurrent runs on one instance may generate and
  read traffic freely; a run that mutates instance-wide state must have the instance to itself for
  the duration of the mutation.

## What the customer provides

The harness is built against these, not in isolation from them:

1. **A fixture tree** — an opaque directory of files, mounted into the harness at runtime. The
   harness serves it byte for byte and never parses, validates, or rewrites its contents.
2. **An acceptance suite** — an executable test suite that verifies the requirements below through
   the harness's public interfaces. It is the acceptance instrument: a requirement is met when the
   part of the suite covering it passes.
3. **Protocol parameters** — the specific credentials, ports, community strings, USM users, topic
   names, and collectors each service must be configured with.

## Functional requirements

Every requirement carries a verification method. A requirement for which no verification method can
be stated is not ready to be implemented.

**R1 — Serve the fixture tree over five transports.** One host serves one tree over HTTP, HTTPS,
FTP, TFTP, and SCP. HTTPS, FTP, and SCP require credentials; the credential set is supplied by
configuration, not by code.
*Verification:* the acceptance suite fetches the same file over each of the five transports and
compares all five byte for byte against the original.

**R2 — Fault injection on file retrieval.** On request, the harness makes a named fixture fail in a
specified way: host unreachable, authentication rejected, file missing, transfer cut off midway,
content corrupted. A fault injection is an instance-wide mutation (N4) and stays in force until
lifted explicitly or by `POST /reset`.
*Verification:* the suite triggers each fault and confirms the transport returns a protocol-level
error rather than a successful transfer of wrong content.

**R3 — Receive what the application sends outward.** One requirement per service; each is
implemented, delivered, and accepted independently of the others.

| | Service | Protocol surface |
| --- | --- | --- |
| R3.1 | Syslog | UDP, TCP, RELP |
| R3.2 | SNMP traps | v2c community; v3 USM across MD5/SHA × DES/AES-128 |
| R3.3 | RADIUS | authentication |
| R3.4 | TACACS+ | authentication |
| R3.5 | DNS | name resolution |
| R3.6 | NTP | time synchronization |
| R3.7 | Kafka | broker |
| R3.8 | IPFIX | collector |

*Verification, per row:* the suite sends a message to the service and reads it back through R4 in
parsed form.

R3.1, R3.2, R3.7, and R3.8 are one-way: the application sends, the harness receives. R3.3–R3.6 are
request–response, and for those the fact of receipt alone is not enough: the test must control the
content of the reply. That is carried by R8.

**R4 — Test access to what was received.** `GET /received/<service>` returns the recorded messages as
JSON — timestamp, source address, parsed content — in arrival order.

Recording and reading are run-scoped. Because the application under test cannot carry a run id
inside its own protocols, attribution works by **claimed source addresses** ([D4](decisions.md)):
a run is declared to the control layer before traffic is generated, and the declaration claims the
source addresses whose traffic belongs to that run. The harness attributes each recorded message to
a run by its source address. A source can be claimed by at most one active run; a claim conflict is
an explicit error. Traffic from unclaimed sources is recorded without run attribution and is visible
only to unfiltered reads. `/received/<service>?run=<id>` returns only that run's messages; a read
for an undeclared run id is an explicit error, not an empty list.

Two runs that cannot be separated by source address (one shared application instance) cannot share a
harness instance — they use separate instances (N4).

*Verification:* two runs of the suite execute concurrently against one instance, each claiming its
own source; each asserts on its own messages and sees none of the other's.

R4 is the requirement that makes the harness a program rather than a bundle of off-the-shelf
containers. Stock images accept traffic, but none of them lets a test ask what arrived.

**R5 — Reset to a known state.** `POST /reset` is instance-wide: it clears the recorded messages of
every receiver, lifts every fault injection from R2 and R9, returns the R8 response controls to
their defaults, releases all run declarations, and restores the served tree to the mounted fixtures,
discarding anything written into it during the run. Reset is a mutation in the sense of N4: it must
not race concurrent runs on the same instance.
*Verification:* the acceptance suite passes twice in a row with nothing but `POST /reset` between the
runs and no process restarted.

**R6 — Serve arbitrary content without interpreting it.** The harness imposes no assumptions about
schema, encoding, or naming on the fixture tree, handles text and binary files identically, and
serves files of at least several hundred megabytes without holding them in memory in full.
*Verification:* the suite fetches a small text file, a large binary file, and a deliberately
malformed file, and compares each byte for byte against the original.

**R7 — One-command startup.** A single `docker run` (D20) brings up every harness service on one
network and provides a documented, machine-readable way to discover each one's address and port.
*Verification:* on a machine holding nothing but a container runtime and this delivery, one command
yields a harness that passes the acceptance suite for every delivered requirement.

**R8 — Control the reply of request–response services.** For R3.3–R3.6, the test sets exactly what
the harness will answer: for RADIUS and TACACS+, success or a rejection with a stated reason; for
DNS, a valid record, `NXDOMAIN`, `SERVFAIL`, or a reply after a specified delay; for NTP, a time
with an offset, an altered stratum, or a refusal to serve. The default for each service is a
successful, valid reply. A response control is an instance-wide mutation (N4) and stays in force
until changed or until `POST /reset`.
*Verification:* for each of the four services the suite sets a negative outcome, issues the protocol
request itself, and receives exactly the reply that was set.

**R9 — Fault injection on the receiving side.** Any receiver from R3 can be put into a failure mode:
refuse connections, accept and silently drop, reply after a specified delay, cut the connection
mid-message, return a protocol error where the protocol allows one. A failure mode is an
instance-wide mutation (N4) and stays in force until lifted explicitly or by `POST /reset`.
*Verification:* the suite enables each mode, generates traffic, and confirms the sending side
observes the corresponding failure and that `/received` matches the mode.

A deliberately configured failure mode must be distinguishable from "the receiver is not running" in
N5: the control layer reports that the receiver is alive and in the configured mode.

**R10 — Receiver interface.** A new protocol is added by implementing a documented contract —
lifecycle, message recording, run attribution, participation in reset and in failure modes — with no
changes to the harness core.
*Verification:* using nothing but the public documentation, a deliberately trivial **reference
receiver** is added; it appears in `/received`, obeys `POST /reset`, and requires no core edits. At
phase 0 the reference receiver is the first receiver and doubles as the worked example in the
documentation; once R3 is delivered in full, it is the ninth.

**R11 — HTTP contract described and versioned.** The `/received` and `/reset` endpoints, run
declarations (R4), the response controls (R8), and the failure modes (R9) are described by a
machine-readable schema shipped with the delivery. A breaking change to the contract requires a
version bump; the current version is available on request.
*Verification:* a client generated from the schema passes the acceptance suite; a version query
returns the version the delivery declares.

**R12 — Offline license validation.** The engine validates a signed license file at startup with no
network access (N2): signature, expiry, and the licensed stands (D15) — only they are started. A missing, expired, or tampered license is
an explicit, documented refusal — never silent degradation (the N5 spirit). The Free mode runs
without a license file within its documented limits (stand set per D10/D15, still open). The engine embeds only the public verification
key; the signing key never ships ([D10](decisions.md)).
*Verification:* the suite starts the engine with a valid, an expired, a tampered, and no license
file, and observes exactly the documented behavior in each case.

**R13 — Active send to a receiving application.** The mirror of R3: for an
application that *receives* a protocol (a syslog server, a DNS server, a Kafka
broker, …), the test asks the engine to send it a message. `POST
/api/v1/send/<service>` takes a `target` (`host:port`) and the payload; the
engine forms the protocol message and sends it to that target. For
request-response protocols the engine is the client and returns the
application's reply; for one-way protocols it confirms the send. **The target is
supplied by the test on every request** — the engine never discovers, stores, or
depends on the application's address (the boundary; see [D18](decisions.md)).
Availability follows the stand grant (D15): sending a protocol requires that
stand to be licensed.
*Verification:* for each protocol the suite stands up a listener (standing in for
the application), asks the engine to send to it, and confirms the listener
received exactly the message (and, for request-response, that the returned reply
matches what the listener answered).

## Non-functional requirements

**N1 — Startup time.** The limit is set from measurement after phase 0, not assigned in advance. It
must stay small enough that the harness remains usable inside an edit–run loop.
*Verification:* the time from the run command to every service reporting ready is measured on
every CI run; after phase 0 the agreed limit is recorded in this file, and the suite fails when it
is exceeded.

**N2 — No external dependencies at runtime.** No internet access, no shared or corporate network, no
services outside the customer's control. The harness must come up on an isolated host.
*Verification:* the acceptance suite passes against an instance brought up on a network with no
external egress, with images built or loaded beforehand.

**N3 — One artifact everywhere.** The same image and run command work on a developer machine and in
CI, differing only in environment parameters rather than diverging into two configurations.
*Verification:* the byte-identical image and documented run command are used by the developer instructions and by
CI; the suite passes in both, with differences expressed only through environment parameters.

**N4 — Safe under concurrency.** Isolation is two-level ([D3](decisions.md)). *Instance level:* two
developers on one host, or two CI jobs on one agent, run their own instances, which must not
interfere — no hard-coded host ports, no shared mutable state outside the container. *Run
level:* within one instance, recording and reads are run-scoped per R4, and concurrent runs may
generate and read traffic freely; instance-wide mutations (R2, R5, R8, R9) require the instance to
the mutating run alone.
*Verification:* two instances come up on one host simultaneously and each passes the acceptance
suite; within one instance, R4's concurrent-run verification passes.

**N5 — Honest failure.** If a receiver is not running, calls against it fail explicitly rather than
returning an empty success. An empty `/received/<service>` result must be distinguishable from "this
receiver was never started."
*Verification:* with one receiver disabled by configuration, protocol traffic to it is refused and
`/received/<service>` for it returns an explicit "not running" error; the same read against a
running receiver with no traffic returns an empty list.

**N6 — Throwaway credentials only.** Every credential, key, and secret shipped with the harness is
fixed, publicly known, and documented as such, so that none of them can be mistaken for a real one.
*Verification:* a secret scanner run over the delivery tree finds nothing resembling a live credential.

**N7 — Reproducible builds.** Every delivered image builds reproducibly from a tagged state of the
source tree on a clean host. Source access itself is a licensing-tier matter ([D10](decisions.md)),
not a blanket delivery requirement.
*Verification:* on a clean host holding only a container runtime and the tagged source tree, the
build instructions produce images that pass the acceptance suite without pulling prebuilt harness
images.

## Readiness matrix

Filled in as implementation proceeds: a requirement is closed if and only if the part of the
acceptance suite covering it passes.

**R4, R5, and R11 are cross-cutting:** each later phase enlarges the surface they must cover, and a
phase is not accepted until they are re-verified over everything delivered so far. Their status
names the latest phase they have been verified against.

| Requirement | Phase | Status | Covering test |
| --- | --- | --- | --- |
| R1 | 0 | passing — all five transports | `TestFixtureByteCompareAllTransports`, `TestFilesAuthRequired` |
| R2 | 1 | passing (HTTP surface) | `TestFileFaults` |
| R3.1 | 2 | passing — UDP, TCP, RELP | `TestSyslogAllTransports` |
| R3.2 | 2 | passing — v2c; v3 USM seeded | `TestSNMPv2cTrap` |
| R3.3 | 3 | passing — accept/reject | `TestRADIUSResponseControl` |
| R3.4 | 3 | passing — pass/fail | `TestTACACSResponseControl` |
| R3.5 | 3 | passing — record/NXDOMAIN/SERVFAIL | `TestDNSResponseControl` |
| R3.6 | 3 | passing — time/stratum | `TestNTPResponseControl` |
| R3.7 | 4 | passing — kfake broker | `TestKafkaProduce` |
| R3.8 | 4 | passing — collector | `TestIPFIXExport` |
| R4 | 0, cross-cutting | passing (phase 0 surface) | `TestRunIsolation`, `TestReferenceReceiverRoundTrip` |
| R5 | 0, cross-cutting | passing (phase 0 surface) | `TestResetTwice` |
| R6 | 0 | passing — opacity + streaming; large-file at CI scale | `TestFixtureByteCompareAllTransports` |
| R7 | 0 | in progress — single-container delivery (D20), clean-host check pending | — |
| R8 | 3 | passing — all four services | `TestDNSResponseControl`, `TestNTPResponseControl`, `TestRADIUSResponseControl`, `TestTACACSResponseControl` |
| R9 | 1 | passing (reference surface) | `TestReceiverModes` |
| R10 | 0 | passing | `TestReferenceReceiverRoundTrip` |
| R11 | 0, cross-cutting | in progress — schema ships, version passing; generated-client check pending | `TestVersion` |
| R12 | 5 | passing | `TestLicenseGatesStands`, `TestFreeMode`, `TestExpiredLicenseRefused`, license unit tests |
| R13 | 6 | passing — all 8 protocols | `TestSend*` |
| N1 | measured after 0 | not measured | — |
| N2 | cross-cutting | not started | — |
| N3 | cross-cutting | not started | — |
| N4 | cross-cutting | not started | — |
| N5 | 0 | passing | `TestEmptyVsUnknown` |
| N6 | cross-cutting | not started | — |
| N7 | cross-cutting | not started | — |
