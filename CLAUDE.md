# CLAUDE.md

Working rules for the Servienta repository. These extend the global principles; on conflict, this
file wins.

## What is being built

Servienta — a provider of testing services. Four deployables (D7, details in
[docs/stack.md](docs/stack.md)):

- **Engine** — the data plane: protocol receivers, five-transport file server, versioned HTTP
  contract. Go, Docker compose, runs on customer/CI infrastructure, operable with no internet (N2).
  Scope: [docs/requirements.md](docs/requirements.md). All protocols, Kafka and FTP included, are
  simulated in-process (D6, D12) — the engine is the delivery's only container.
- **Console** — management, settings, statistics. One Go service in Docker on the vendor host,
  listening on :80 (TLS at Cloudflare's edge); serves the Vue 3 + Tailwind SPA and the REST API.
  SQLite now, Postgres later (D9 as amended by D11). console.servienta.com. Talks to engines only
  through the versioned engine contract — it is the contract's first client.
- **Admin** — vendor back office (license issuance per D10, customers, plans). Its own Workers
  app at admin.servienta.com: Vue 3 SPA + Hono API + Cloudflare D1 (D13); single-user
  email+password auth with session cookies and email reset (D14).
- **Marketing site** — static HTML + Tailwind on Workers. servienta.com, www.servienta.com.

**The boundary that must not be crossed.** The application under test and the system it works with
do not belong to the engine. The engine does not poll them, does not configure them, and does not
depend on them. It accepts, serves, and reports what it observed. Code that reaches into the
application under test is a design error, not a convenient shortcut.

**Naming:** the Cloudflare database is always written **Cloudflare D1**, fully qualified — bare
"D1" means the decision-log entry.

## Invariants

Breaking any of these is a reason to stop and discuss, not to work around:

1. **Run attribution is mandatory from the first commit.** Runs are declared, sources are claimed,
   and every recorded message is attributed by claimed source (D4) — or explicitly unattributed.
   Never guessed. Not optional, not "we'll add it when we need concurrency": optional attribution
   degenerates into a shared log, and N4 breaks under load.
2. **Mutation is instance-wide; reading is run-scoped (D3).** Reset (R5), fault injection (R2, R9),
   and response controls (R8) affect the whole instance, and a mutating run owns the instance for
   the duration. Recording and `/received` are scoped by run. Never blur the two levels.
3. **`POST /reset` is complete by construction.** Any new state — recorded messages, an injected
   fault, a response mode, a run declaration, a write into the served tree — is registered with
   reset by the same change that introduces it. State that outlives a run fails the next test for no
   visible reason.
4. **A receiver is added only together with the test that reads it.** A receiver whose observations
   nobody reads is infrastructure without an owner.
5. **The fixture tree is opaque.** Bytes are served as they are. No parsing, no schema validation, no
   encoding detection, no name normalization. Large files stream; they are never held in memory in
   full.
6. **Throwaway credentials only.** Anything that looks like a secret is fixed, publicly known, and
   explicitly documented as unusable for real access (N6).
7. **No hard-coded host ports.** Two instances on one host must not collide (N4). Addresses and
   ports are reported machine-readably (R7).
8. **Empty ≠ not running.** "No messages" and "no receiver" are distinguishable answers (N5), and a
   read for an undeclared run is an error, not an empty list (R4). The same holds for a deliberate
   failure mode (R9): the control layer reports that the receiver is alive and in the configured
   mode.
9. **The contract is versioned.** A breaking change to `/received`, `/reset`, run declarations, the
   R8 response controls, or the R9 failure modes requires a version bump and an update to the schema
   in the delivery (R11). The console consumes engines through this contract only.
10. **Free tier by default.** A component that cannot run on a free tier enters the stack only
    through a decision-log entry (D7). The domain is the single accepted spend.
11. **Permissive dependencies only in the engine.** The delivery redistributes what it contains:
    MIT/BSD/Apache-2.0 only. BSL, GPL, or AGPL components enter only through a decision-log entry
    (D10). Kafka is simulated in-process by `kfake` (BSD-3, D12); the only sanctioned container
    fallback is `apache/kafka` (Apache-2.0) — never Redpanda (BSL).

## Definition of done

An engine requirement is closed if and only if the part of the **acceptance suite** covering it
passes. Not "the code is written," not "I checked it by hand." On closing, update the row in the
readiness matrix in `docs/requirements.md` with a link to the specific test. R4, R5, and R11 are
cross-cutting: every phase re-verifies them over everything delivered so far.

## How work proceeds

- Scope changes by a diff to `docs/requirements.md`, not by verbal agreement.
- A non-trivial decision gets an entry at the top of `docs/decisions.md`, using the template there.
- The phase order in `docs/roadmap.md` holds: R5 is confirmed first, before anything else comes to
  depend on it.
- A requirement for which no verification method can be stated is not ready to implement — say so
  instead of starting.

## Current state

Implementation has started with the marketing site and the admin panel (D13). Engine code is not
written until the phase 0 inputs are closed.
Decided: D1→D6 (engine runtime: Go), D3 (concurrency model), D4 (source-claim attribution), D5 (R4
core in phase 0), D7 (topology, domains, free-first), D8 (console/site stack), D9 (data-layer
portability), D10 (licensing: offline signed license, Free edition, R12), D11 (console and admin in
Docker on the vendor host; SQLite → Postgres), D12 (Kafka simulated in-process by kfake — the
engine is the only delivered container), D13 (implementation starts with the marketing site and
the admin panel on Cloudflare), D14 (admin auth: single-user email+password). Outstanding phase 0 inputs: the fixture tree, the executable acceptance
suite, the protocol parameters — see `docs/roadmap.md` → "Before phase 0 starts".
