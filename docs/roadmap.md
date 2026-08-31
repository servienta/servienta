---
title: Test Harness — Phases and Acceptance
type: roadmap
status: active
updated: 2026-08-31
---

# Phases

Building the whole thing before any of it becomes usable is a mistake: a receiver whose
observations no test reads is infrastructure without an owner.

| Phase | Contents | What it unblocks | Status |
| --- | --- | --- | --- |
| 0 | R5 (reset), R7 (compose), R1 and R6 (file server), R4 (recording core: `/received`, run declarations), R10 (receiver interface, reference receiver), R11 (contract) | Reproducible runs; import verification; extensibility from the start | **in progress** |
| 1 | R2 and R9 (fault injection on both sides) | Verification of error paths | in progress — HTTP + reference surface |
| 2 | R3.1 and R3.2 (syslog, SNMP traps) | First checks of outbound integrations | not started |
| 3 | R3.3–R3.6 with R8 (RADIUS, TACACS+, DNS, NTP) | Authentication and system-settings screens, failures included | not started |
| 4 | R3.7 and R3.8 (Kafka, IPFIX) | Streaming and flow-export screens | not started |
| 5 | R12 (offline license validation), per-customer registry access, license issuance in the admin panel | First paid delivery | not started |

**R5 is implemented and confirmed first, before anything else comes to depend on it.** An incomplete
reset produces flaky end-to-end runs, and flaky tests are worse than no tests: they get disabled,
and after that the harness is pointless.

**R4, R5, and R11 are cross-cutting.** Each phase enlarges the surface they cover — new receivers to
record and reset, new endpoints in the schema. A phase is accepted only when criteria 2, 3, and 7
below pass again over everything delivered so far, not just over the phase's additions.

## Phase 0 acceptance criteria

1. One command on a clean host brings up the harness; its file server answers over all five
   transports.
2. The acceptance suite fetches fixtures over each of the five transports and compares them byte for
   byte.
3. The same suite, run twice in a row with nothing but `POST /reset` between the runs, passes both
   times.
4. The delivery includes documentation: how to start the harness, how to mount the fixture tree, how
   to set the protocol parameters for each service, and how to add another receiver.
5. A machine-readable schema of the HTTP contract is part of the delivery, and a version query
   returns the version the delivery declares (R11).
6. The reference receiver has been added from the public documentation alone, with no core edits
   (R10).
7. Two concurrent runs against one instance, each with its own claimed source, send traffic to the
   reference receiver and each reads back only its own messages (R4).

## Before phase 0 starts

Phase 0 does not start until its inputs are closed:

| Input | Owner | Status |
| --- | --- | --- |
| D1 — control-layer runtime | contractor proposes, customer approves | decided 2026-08-31, revised by D6 — Go, see [decisions.md](decisions.md) |
| Fixture tree available to mount | vendor | closed — [fixture-tree.md](fixture-tree.md) |
| Acceptance suite in executable form | vendor | closed — [acceptance-suite.md](acceptance-suite.md), `apps/engine/acceptance/` |
| Protocol parameters (ports, credentials, community, USM, topics) | vendor | closed — [protocol-parameters.md](protocol-parameters.md) |

The last three rows are the "what the customer provides" list from the requirements. Without the
acceptance suite no requirement can be declared met: the suite *is* the definition of done.

## Risks

| Risk | Consequence | Mitigation |
| --- | --- | --- |
| `/reset` implemented incompletely | Flaky runs, tests get disabled, the harness gets abandoned | Phase 0 implements and confirms reset first (criterion 3) |
| The harness grows without an owner | One more system to maintain, with no users | A receiver is added only together with the test that reads its observations |
| Run-id discipline erodes | R4 degenerates into a shared log, N4 breaks under load | Run declarations and run attribution are mandatory in the API from the first commit, not optional |
| Source claims mistaken for heuristics | Attribution silently guesses, and runs read each other's traffic | A source is claimed by at most one active run, claim conflicts are explicit errors, unclaimed traffic is never attributed (D4) |
| The fixture tree is treated as structured data | The harness rejects or rewrites valid input it is supposed to pass through unchanged | R6 is verified with a binary file and a deliberately malformed one |
| R8 or R9 state outlives a run | The next test starts against a failing service and fails for no visible reason | R5 must lift both; acceptance criterion 3 runs the suite twice in a row |
| Extensibility implemented in form only | A ninth protocol still requires core edits, and the harness stays a kit for one fixed service set | R10 is verified by adding the reference receiver from the public documentation alone |
