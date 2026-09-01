---
title: Acceptance Suite
type: reference
status: active
updated: 2026-08-31
---

# Acceptance suite

The executable definition of done (phase 0 input #2, roadmap): a requirement is closed if and only
if the part of this suite covering it passes. Lives in `apps/servienta/acceptance/` (Go tests); it
talks to the engine **only through public interfaces** — network listeners and the versioned HTTP
contract — never through internals.

Run: `go test ./acceptance/...` in `apps/servienta/`.

## Phase 0 mapping

| Test | Covers | Criterion |
| --- | --- | --- |
| `TestVersion` | R11 — version query matches the declared contract | 5 |
| `TestFixtureByteCompareHTTP` | R1 (HTTP), R6 — small text, binary, malformed byte-compare | 1, 2 |
| `TestReferenceReceiverRoundTrip` | R10, R4 — record via reference receiver, read back parsed | 6 |
| `TestRunIsolation` | R4, N4 — two concurrent runs, claimed sources, no cross-reads | 7 |
| `TestClaimConflict` | R4/D4 — one source, one active run; conflict is an explicit error | 7 |
| `TestEmptyVsUnknown` | N5 — empty result ≠ unknown receiver ≠ undeclared run | — |
| `TestResetTwice` | R5 — the suite passes twice with only `POST /reset` between | 3 |

Transports FTP/TFTP/SCP/HTTPS extend `TestFixtureByteCompare*` as they land; every later phase adds
its tests here and re-runs the cross-cutting ones (R4, R5, R11) over the grown surface.
