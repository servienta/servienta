---
title: Fixture Tree
type: reference
status: active
updated: 2026-08-31
---

# Fixture tree

The opaque directory the engine serves (phase 0 input #1, roadmap). Mounted read-only at
`SERVIENTA_FIXTURES` (default `/fixtures`).

**Opacity rules (R6, invariant 5):** the engine serves bytes as they are — no parsing, no schema or
encoding assumptions, no name normalization; text and binary are identical; large files stream and
are never held in memory in full.

The tree's *content* is whatever the tests need; nothing is committed to this repository except the
acceptance minimum, which the acceptance suite generates into a temp dir at run time:

| File | Purpose |
| --- | --- |
| `small.txt` | small text file — byte-compare over every transport (R1) |
| `large.bin` | multi-hundred-MB random binary — streaming proof (R6); generated, never committed |
| `malformed.bin` | deliberately invalid content (truncated archive header) — opacity proof (R6) |

Anything written into the served tree during a run (future transports permitting writes) is
discarded by `POST /reset` (R5), restoring the mounted state.
