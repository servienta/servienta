---
title: Decision Log
type: decision-log
status: active
updated: 2026-09-01
---

# Decision Log

New entries go on top. The format is mandatory; a decision without a revisit trigger hardens into
an unexamined given. A decision that is raised but not yet made carries `Status: open` in place of
`Choice`; every other field still applies.

```
## D<N> — <short title>
Date: YYYY-MM-DD
Context: <problem, constraints>
Options considered: <A, B, C>
Choice: <X>, because <reason>
Reversibility: <one-way / two-way door>
Revisit trigger: <metric / date / condition>
```

---

## D19 — The compose stand publishes the engine 1:1; the console moves to :5000

Date: 2026-09-01
Context: the compose stand originally published only the console, on host
:8080 — the same number as the engine's control port. The engine had no host
ports at all: the Stand view showed receivers as `[::]:8081 · up` that the
host could not reach, and every quickstart curl against :8080 hit the console
instead of the contract. The stand could be managed but not exercised.
Options considered: (A) publish the engine 1:1 — contract included — and move
the console off :8080; (B) publish only the data plane, keep control behind a
console proxy; (C) publish through env-overridable host ports; (D) keep the
engine unpublished and only label the addresses as internal.
Choice: **A** — the stand exists to run the loop against, and 1:1 keeps every
address the engine reports (R7) true on the host, contract at `:8080` exactly
as the site quickstart and the walkthrough assume. The console listens on
`:5000`. One exception: `dns` maps to host
15353 because mDNS owns 5353/udp on desktops.
Known limits: TFTP transfers reply from ephemeral ports and don't survive the
NAT (use a standalone engine); host-originated traffic arrives with the Docker
gateway as its source (172.18.0.1), which is what source-claim attribution
(D4) will see.
Reversibility: two-way door — a compose edit.
Revisit trigger: two stands on one host (host-port collisions → option C), or
the first user tripped up by NAT'd source attribution.

## D18 — Active send: the test supplies the target every time

Date: 2026-09-01
Context: the harness was built to *receive* what the application sends outward
(R3). A new need is the mirror — sending a protocol message *into* an
application that receives it (a syslog/DNS server, a Kafka broker). This touches
the core boundary: "the engine never reaches into the application under test."
Options considered: (A) the engine is configured with the application's address
and sends to it — makes the engine depend on and discover the application,
breaking the boundary; (B) the test supplies the `target` on every send request,
so the engine only executes a test-directed action and never holds or discovers
the application's address.
Choice: **B**. The initiator is always the test, exactly as with reads and
resets; the engine forms and sends the protocol message to the target the test
names, and holds nothing about the application between requests. Serving files
and answering DNS were already "sending" in response to app-initiated requests
(R1, R8); R13 adds test-initiated sends, still without the engine depending on
the application. New capability recorded as R13.
Reversibility: two-way — senders are stateless and additive; removing them
touches nothing else.
Revisit trigger: a real need for the engine to send on its own schedule
(unprompted by a test) — that would genuinely cross the boundary and needs its
own decision, not this one.

---

## D17 — Pricing model: annual subscription, tiered by stand bundles

Date: 2026-09-01
Context: the license mechanism (D15/D16) can price on exactly two axes — the set
of stands and the expiry — and nothing else: the engine is offline (N2), so
usage cannot be metered or observed, and instance/seat counts cannot be
enforced in a source-available binary. A pricing model has to live inside those
limits.
Options considered: (A) usage/consumption pricing — impossible to meter offline;
(B) per-instance or per-seat — unenforceable offline, contract-only at best;
(C) annual subscription tiered by stand bundles, instance count governed by
contract. This is the self-hosted-infra standard (Elastic, HashiCorp, GitLab
self-managed) precisely because they share these constraints.
Choice: **C**. Plans (presets over the stand set):
- **Free** — files-http + demo receiver, no license file (D16). Funnel/trial.
- **Files** — all five file transports (http, https, ftp, tftp, scp).
- **Standard** — Files + syslog, snmp-traps, dns, ntp.
- **Enterprise** — everything (adds radius, tacacs, kafka, ipfix).
- **Custom** — an arbitrary stand set, for negotiated deals (the admin already
  supports freeform selection).
Billing is an annual subscription: renewal is a reissued license with a new
`exp` (already implemented). Instance count is a **contract term, not code** —
offline enforcement is impossible and chasing it is wasted effort (see the
fingerprint discussion; D10's "compliance clarity, not DRM"). Prices themselves
are set out of band and are not encoded here.
Reversibility: two-way — plans are presets over the stand list (data), and the
license already carries the exact stands, so repricing never touches the engine.
Revisit trigger: a customer need a stand bundle cannot express (per-stand
metering, graduated limits), or a move to a hosted offering where usage *can* be
metered — either reopens the axes available to pricing.

---

## D16 — Free mode serves the HTTP file server only

Date: 2026-08-31
Context: D10 left open what the no-license Free mode includes; the engine now
enforces licensing (R12) and needs a concrete answer.
Options considered: (A) Free = nothing (license always required); (B) Free =
the HTTP file server only; (C) Free = all file transports.
Choice: **B** — the import-check use case (the original R1/R6 pain) needs only
file serving, and HTTP is the zero-credential transport, so Free mode is
immediately useful without handing over the paid surface. A rejected or expired
license falls back to Free and reports the error (R12: explicit, never silent).
Reversibility: two-way — the Free stand set is data (`app.FreeStands`).
Revisit trigger: product/pricing decides a different free tier.

---

## D15 — License semantics: a license enumerates the customer's stands

Date: 2026-08-31
Context: the scaffolded license carried a placeholder `edition` enum (free/standard/enterprise)
with no defined meaning. Owner decision: a license defines **which stands are available to the
customer** — the file server and each protocol receiver are the units of licensing.
Options considered: (A) named tiers mapping to stand bundles; (B) the license payload carries the
explicit list of licensed stand ids.
Choice: **B** — the payload's `stands` array is the grant itself; bundles, if ever needed, become a
convenience in the admin UI, not a license concept. Stand ids (file transports licensed
individually, per owner): `http`, `https`, `ftp`, `tftp`, `scp`, `syslog`, `snmp-traps`,
`radius`, `tacacs`, `dns`, `ntp`, `kafka`, `ipfix`. The engine (phase 5, R12) starts only licensed
stands; an unlicensed stand is reported as "not licensed" — distinguishable from "not running"
(N5). What the no-license Free mode includes stays open in D10 until the engine enforces licensing.
Amends: D10 ("edition limits are data" — the data is now the stands list); the admin's edition
field is replaced.
Reversibility: two-way until the first paid delivery; after that a payload change bumps the license
format version (`v`).
Revisit trigger: a real customer needs limits a stand list cannot express (instance counts, seats,
expiry per stand) — extend the payload then, not before.

---

## D14 — Admin auth: single-user email+password, no Cloudflare Access

Date: 2026-08-31
Context: owner decision — Cloudflare Access (D13) is unwanted ceremony for a one-person back
office; required instead: plain login for andrew@molyuk.com with a password-reset capability.
Options considered: (A) keep Access; (B) better-auth — a full auth framework for one fixed user;
(C) minimal app-level auth sized to the actual requirement.
Choice: **C**. One user row in Cloudflare D1 (PBKDF2-SHA-256 via WebCrypto — argon2 is not
Workers-native and heavy KDFs collide with the free tier's CPU budget), stateless HMAC-signed
session cookie (secret in a Worker secret), endpoints login/logout/change-password, and a
forgot/reset flow: one-time token, 15-minute expiry, stored hashed, delivered by email through
Resend (free tier; until RESEND_API_KEY is set, password change is available from inside the UI
after login). The initial password is generated at seed time and handed to the owner out of band —
no credential or its hash is committed.
Amends: D13 (auth only; the Workers + Cloudflare D1 placement stands).
Reversibility: two-way — the session cookie contract is internal to the admin.
Revisit trigger: a second admin user, or any external exposure beyond the one owner — that is the
moment to adopt better-auth or an IdP rather than grow this by hand.

---

## D13 — Implementation starts with the marketing site and the admin panel, both on Cloudflare

Date: 2026-08-31
Status: **amended by D14 (2026-08-31): admin auth is app-level, not Cloudflare Access.**
Context: owner decision — begin with the simple, independently shippable parts. The marketing site
was already Cloudflare-hosted (D7). The admin panel returns to Cloudflare, revising D11's
placement: the vendor back office has no tie to the customer-facing Docker host, and on Workers it
ships immediately at zero hosting cost.
Options considered: (A) admin stays inside the console's Go service per D11 — cannot ship until the
console exists; (B) admin as its own Workers app — Vue 3 SPA + Hono API + Cloudflare D1, the
original D8/D9 stack.
Choice: **B**, with one further simplification: at the start the admin's authentication *is*
Cloudflare Access — the Worker verifies the `Cf-Access-Jwt-Assertion` JWT, and there is no
application-level password store at all. Data: Cloudflare D1 under D9's portability rules
(Cloudflare D1 is SQLite, so they apply verbatim). The license signing key (D10) returns to
Cloudflare secrets; the console's Docker host never holds it. The console itself remains exactly
per D11 — a Go service in Docker on :80. Until the domain is registered, both apps deploy to
workers.dev URLs; routes for servienta.com, www, and admin attach once the zone exists.
Amends: D11 (admin placement only), D10 (signing-key location), D8 (admin auth: Cloudflare Access
instead of an app-level scheme).
Reversibility: two-way — the admin SPA talks to a REST API either way, and D9's rules keep the data
portable in every direction (Cloudflare D1 ↔ SQLite file ↔ Postgres).
Revisit trigger: the admin needing what Workers lacks (long-running jobs, direct engine
connectivity), or Cloudflare Access outgrowing the free 50 seats.

---

## D12 — Kafka is simulated in-process; no broker container in the delivery

Date: 2026-08-31
Context: owner question — can Kafka and FTP be simulated instead of shipped as real services? FTP
already is: `ftpserverlib` is a full in-process FTP server (D6). Kafka was planned as a real
`apache/kafka` container (D6, amended by D10) — the heaviest piece of the delivery: a JVM broker
dominating image size, memory, and startup time (N1), and needing a TCP proxy for R9 fault modes.
Options considered: (A) real `apache/kafka` container behind the engine; (B) in-process simulated
broker — `twmb/franz-go/pkg/kfake` (BSD-3, the same repository as the client library already
chosen), which implements enough of the wire protocol for standard clients: ApiVersions, Metadata,
Produce/Fetch, group management; (C) a hand-rolled Produce-path subset.
Choice: **B**. R3.7 is one-way — the application produces, the harness records — so produce-path
fidelity is what matters, and kfake provides it while making the engine the only container in the
delivery: startup drops to process start (N1), R5 reset becomes an in-memory operation, R9 fault
modes are native code instead of a TCP proxy, and the broker licensing question disappears (BSD-3
satisfies the permissive-only invariant). Honest limits, recorded: kfake is not a real broker —
transactional producers, rebalancing edge cases, and exact error semantics may diverge.
Amends: D6 and D10's delivered-broker corollary — no broker ships at all.
Reversibility: two-way — the engine's contract does not expose which Kafka sits behind the socket;
swapping in the `apache/kafka` container restores the D6 pattern unchanged.
Revisit trigger: the application's client requires broker behavior kfake lacks — that specific gap,
demonstrated by a failing acceptance test, brings back the `apache/kafka` container for exactly
that surface.

---

## D11 — Console and admin run in Docker on the vendor host, not on Cloudflare Workers

Date: 2026-08-31
Status: **amended by D13 (2026-08-31): the admin returns to Cloudflare Workers; everything about the console stands.**
Context: owner decision — console.servienta.com serves from Docker on port 80, not from Cloudflare
Workers. With the API off Workers, the Workers-specific stack (Hono, Cloudflare D1 bindings,
better-auth) loses its rationale, and one backend language across the whole product becomes
possible.
Options considered: (A) SPA in Docker but the API still on Workers + Cloudflare D1 — splits one
product across two runtimes and keeps JS on the backend for no gain; (B) the whole console — API
and both SPAs — as one Go service in Docker on the vendor host.
Choice: **B**. One Go service (module `console/`) serves the REST API and the static builds of both
SPAs; the container listens on :80; console.servienta.com and admin.servienta.com are proxied
Cloudflare DNS records pointing at the vendor host — TLS terminates at Cloudflare's edge, and
Cloudflare Access still fronts admin (`cloudflared` Tunnel is the free hardening option that closes
the inbound port entirely). Data: **SQLite now, Postgres later** — D9's portability rules carry
over verbatim, since they were written for exactly this SQLite→Postgres pair: driver
`modernc.org/sqlite` (CGO-free), migrations `pressly/goose`, queries generated by `sqlc` for both
dialects, all access behind a repository layer. Auth: Go-native sessions with argon2id hashing —
better-auth (JS) drops out with Workers. The license signing key (D10) becomes a mounted secret on
the vendor host, still never part of any delivery. Cloudflare keeps: DNS, edge TLS, Access, and the
marketing site on Workers static assets. Hosting cost stays zero on own hardware; a rented VPS
would be the first recurring cost and needs its own decision entry (D7's free-first rule).
Amends: D7 (topology), D8 (API runtime and auth — the Vue 3 + Tailwind frontend stack is
unchanged), D9 (the store is local SQLite instead of Cloudflare D1), D10 (signing-key location).
Reversibility: two-way — the SPAs talk to a REST API either way, and D9's rules keep the store
portable.
Revisit trigger: the vendor host becoming a bottleneck or a single point of failure for paying
customers — that reopens hosting (managed Postgres plus a second node, or edge hosting for the
stateless parts).

---

## D10 — Engine licensing and delivery model

Date: 2026-08-31
Context: the engine ships to customers as container images and runs on their infrastructure,
offline (N2). Phone-home activation is therefore impossible by our own requirement, and any
technical enforcement in self-hosted software is bypassable by a determined licensee — the
mechanism exists for compliance clarity, not as DRM. N7's blanket "delivered as source" came from
the single-customer proposal and contradicts a licensed product.
Options considered: enforcement — (A) offline-verifiable signed license file; (B) phone-home
activation, contradicts N2; (C) contract only, no mechanism. Source policy — (D) keep N7 blanket
source delivery; (E) rescope N7 to reproducible builds, source access becomes a licensing-tier
matter.
Choice: **A + E**, plus the contract itself. A signed license file (Ed25519), mounted like any
other configuration; the engine embeds only the public verification key and validates signature,
expiry, and edition limits at startup with no network access. The signing key never ships and lives
outside the delivery (location per D13: Cloudflare secrets, readable only by the admin Worker);
licenses are issued and managed from the admin panel (admin.servienta.com). The **Free edition runs without a license file** within
documented limits — the product is free at the start (D7). Distribution: per-customer pull access
to a private registry (ghcr.io to start). Recorded as requirement **R12**; **N7 rescoped** to
reproducible builds.
Two licensing corollaries, amending D6/stack: the **delivered Kafka broker is `apache/kafka`
(Apache-2.0)** — Redpanda's BSL restricts redistribution, so Redpanda stays a local-development
convenience only; and the delivery requires **any OCI runtime** (Docker Engine, Podman,
containerd) — Docker Desktop is never a dependency, so its per-company subscription terms remain
the customer's affair. Engine dependencies: permissive licenses only (MIT/BSD/Apache-2.0); BSL,
GPL, or AGPL components enter only through a decision-log entry.
Reversibility: two-way door — the license file format is versioned alongside the contract, and
edition limits are data, not code paths.
Revisit trigger: the first paid delivery (R12 must be closed by then); a customer requiring source
access defines the source-available terms at that point.

---

## D9 — Cloudflare D1 now, Postgres later: portability rules

Date: 2026-08-31
Status: **amended by D11 (2026-08-31): the store is local SQLite, not Cloudflare D1; the portability rules stand verbatim.**
Context: the console's data layer starts on Cloudflare D1 (free, zero-ops, pairs with Workers), but
the owner requires a real migration path to Postgres. SQLite and Postgres diverge enough that an
unplanned migration is a rewrite.
Options considered: (A) Cloudflare D1 with enforced portability rules and Drizzle ORM; (B) Postgres
from day one (Neon/Supabase free tiers) — adds latency and a second platform to a Workers-native
stack; (C) Cloudflare D1 with no discipline, migrate "when needed."
Choice: **A**. Drizzle ORM migrations are the single schema source; SQL stays inside the
SQLite∩Postgres common subset — text ULIDs for ids, integer epoch-millis UTC timestamps, JSON as
text, no SQLite pragmas, no AUTOINCREMENT reliance; every data access goes through a repository
layer, and raw Cloudflare D1 bindings never leak outside it. Migration then is: point Drizzle at
Postgres, run migrations, copy rows.
Reversibility: two-way door — that is the point of the rules.
Revisit trigger: approaching Cloudflare D1 free ceilings (5 GB storage, 100k row writes/day), a need
for multi-statement transactions or heavier relational features, or a self-hosted console
requirement — any of these starts the Postgres migration.

---

## D8 — Console and marketing-site stack

Date: 2026-08-31
Status: **amended by D11 (2026-08-31): the API runtime is Go in Docker, not Hono on Workers; better-auth is replaced by Go-native sessions. The Vue 3 + Tailwind frontend stack stands.**
Context: the owner set Vue 3 + Tailwind for the management/settings/statistics frontend; the API
runtime, auth, validation, and the marketing site's form remained open.
Options considered: API — (A) Hono on Workers, (B) itty-router, (C) Workers runtime raw; auth —
better-auth vs hosted (Clerk) vs hand-rolled; site — static HTML vs Astro vs Vue SSR.
Choice: frontend Vue 3 + TypeScript + Vite + Pinia + Vue Router + Tailwind CSS v4, headless
components reka-ui *(proposed, swappable)*, ECharts for statistics *(proposed)*. API: **Hono** +
zod + `@hono/zod-openapi` — schema-first, mirroring the engine's R11 discipline. Auth:
**better-auth** — open source, free, runs on Workers + Cloudflare D1; hosted auth adds a paid
dependency, hand-rolled auth is unjustifiable risk. Marketing site: **static HTML + Tailwind** on
Workers static assets — zero framework until page count justifies one (Astro is the named
candidate).
Reversibility: two-way door throughout; the console consumes the engine only through the versioned
contract, so console-stack changes never touch the engine.
Revisit trigger: marketing site exceeding ~5 hand-maintained pages → Astro; auth needing
enterprise SSO → revisit hosted providers.

---

## D7 — Product topology, domains, and the free-first rule

Date: 2026-08-31
Status: **amended by D11 (2026-08-31): console and admin run in Docker on the vendor host; Cloudflare keeps DNS, edge TLS, Access, and the marketing site.**
Context: the product (owner decision 2026-08-31) is a provider of testing services: the harness
becomes the **engine**, joined by a hosted management console and a public marketing site. The
engine's protocol surface is raw UDP/TCP listeners (syslog, SNMP traps, DNS, NTP, RADIUS…), which
serverless platforms do not offer — so "the system on Cloudflare" can only mean its control plane.
Options considered: (A) everything self-hosted on a VPS — recurring cost from day one; (B) control
plane and site on Cloudflare free plan, engine as Docker wherever the tests run; (C) engine also
"hosted" behind tunnels — dismissed: contradicts N2 and the raw-socket surface.
Choice: **B** — four deployables: **engine** (Go, Docker compose, customer/CI infrastructure,
operable with no internet per N2), **console** at console.servienta.com (customer-facing: Vue 3 SPA
+ Hono API on Cloudflare Workers + Cloudflare D1), **admin** at admin.servienta.com (vendor back
office: license issuance, customers, plans — its own Vue 3 SPA sharing the console's API and
database behind role gating, with Cloudflare Access (Zero Trust free tier) in front as a second
lock), and the **marketing site** at servienta.com with www redirecting to the apex (Workers static
assets). The console talks to engines only through the versioned engine
contract (R11) and is its first generated client; the engine never depends on the console.
**Free-first rule:** every component starts on a free tier; the domain is the only unavoidable
spend; a recurring cost enters the stack only through a decision-log entry.
Reversibility: two-way per component — the engine runs anywhere containers do, the console's
portability is guaranteed by D9, the site is static files.
Revisit trigger: Workers free ceilings (100k req/day, 10 ms CPU) or a console feature needing
websockets/Durable Objects — the paid-tier decision gets its own entry.

---

## D6 — Engine runtime: Go (supersedes D1)

Date: 2026-08-31
Context: D1 chose Python asyncio on library-maturity grounds. The product direction (owner decision
2026-08-31) sets Go for the service code: one static binary per service, scratch-based images,
trivial cross-compilation, and one backend language across the product.
Options considered: (A) keep Python per D1; (B) Go, with the protocol-library map re-verified.
Choice: **B** — owner decision, and the re-verified map holds: stdlib `net/http` (contract, HTTP/S
file serving), `miekg/dns`, `layeh/radius`, `gosnmp` (v3 USM MD5/SHA × DES/AES-128),
`mcuadros/go-syslog` (UDP/TCP), `nwaples/tacplus`, `vmware/go-ipfix`, `fclairamb/ftpserverlib`,
`pin/tftp`, `gliderlabs/ssh` (SCP), `twmb/franz-go` (Kafka consumer). The gaps — RELP framing, the
NTP server subset, thin spots in TACACS+ — are small, well-specified wire formats implemented
in-repo. The Kafka pattern from D1 stands: a real broker container behind the engine, faults
injected by a TCP proxy in front of it (amended by D10: the delivered broker is `apache/kafka`;
Redpanda is local-development only).
Reversibility: two-way door as long as the HTTP contract (R4, R5) stays stable — unchanged from D1;
the contract is fixed and versioned before implementation begins.
Revisit trigger: a protocol in phases 2–4 with no usable Go library — that receiver moves into its
own container behind the engine (the Kafka pattern) and this entry is amended.

---

## D5 — R4 recording core moves to phase 0; the reference receiver

Date: 2026-08-31
Context: phase 0 delivered R10 and R11, but both verify against `/received`, which belonged to R4 in
phase 2. Phase 0's acceptance criterion also spoke of a "ninth receiver" at a point where zero
receivers exist.
Options considered: (A) leave R4 in phase 2 and weaken the R10/R11 verifications to not touch
`/received`; (B) move the R4 core — recording pipeline, `/received`, run declarations — into phase 0
and prove it with a deliberately trivial reference receiver.
Choice: **B**. Weakened verifications (A) would leave extensibility and the contract unproven until
phase 2, which is exactly the "extensibility implemented in form only" risk. The reference receiver
is the first receiver at phase 0, doubles as the worked example in the receiver documentation, and
becomes the ninth once R3 is delivered in full. Phases 2–4 now deliver only protocol receivers.
Reversibility: two-way door.
Revisit trigger: if phase 0 grows past what one person delivers in one iteration, split the file
server (R1, R6) into its own sub-phase — the recording core stays in phase 0.

---

## D4 — Traffic is attributed to runs by claimed source addresses

Date: 2026-08-31
Context: R4 required every write path to carry a run id, but on the write path the caller is the
application under test, which cannot embed a run id in a syslog packet, an SNMP trap, or a RADIUS
request. Without a defined mechanism, run attribution — and with it N4 — was unimplementable.
Options considered: (A) claimed source addresses — a run declares itself and claims the sources
whose traffic belongs to it; (B) per-run dedicated receiver ports allocated at run declaration;
(C) content markers injected into the application's own traffic.
Choice: **A**, because it needs no cooperation from the application under test and no port
multiplication, and its failure mode is explicit: a source is claimed by at most one active run,
claim conflicts are errors, unclaimed traffic is recorded without attribution and never guessed at.
Known limitation, recorded deliberately: two runs driving one shared application instance cannot be
separated by source and must use separate harness instances.
Reversibility: two-way door — (B) can be added later as a non-breaking contract extension.
Revisit trigger: a real consumer needs two concurrent runs against one shared application instance;
that is the signal to add per-run port allocation (B), not to loosen claim semantics.

---

## D3 — Two-level concurrency model: instance-wide mutation, run-scoped reads

Date: 2026-08-31
Context: R5 reset "clears every receiver" and R2/R8/R9 flip shared listeners, while N4 and R4's
verification promised concurrent runs that do not interfere. A run resetting or degrading a shared
listener mid-flight destroys the other run's traffic — the requirements contradicted each other.
Options considered: (A) run-scope everything, including faults and reset — impossible for
listener-wide failure modes ("refuse connections") without per-run receiver instances; (B) drop run
ids and isolate only at instance level — loses cheap sequential reuse and R4's concurrent read
verification; (C) two levels: mutations of shared behavior (R2, R5, R8, R9) are instance-wide,
recording and reads are run-scoped, and a mutating run must have the instance to itself.
Choice: **C** — it is the only option that is both implementable over shared listeners and honest
about what concurrency is safe. Instances are cheap by construction (N4 already bans fixed host
ports precisely so several can coexist on one host).
Reversibility: two-way door until contract v1 ships; after that the scoping rules are contract
semantics and changing them is a breaking version bump (R11).
Revisit trigger: a real need for concurrent fault injection on one instance — the answer then is
per-run receiver allocation (see D4's option B), not implicit sharing.

---

## D1 — Control-layer runtime

Date: 2026-08-31
Status: **superseded by D6 (2026-08-31): the engine runtime is Go.** Kept for the reasoning trail.
Context: R1–R6 can be implemented by a single process if its runtime has mature libraries for all
five transports (HTTP, HTTPS, FTP, TFTP, SCP); otherwise the control layer acts as a facade in front
of several containers. The choice determines which part of the harness is a single deployable unit.
Options considered: (A) a single Python asyncio process, with a real broker container only where a
protocol cannot reasonably be served in-process; (B) a single Go binary; (C) a pure facade over
off-the-shelf containers.
Choice: **A**. Python is the one candidate with mature server-side libraries across the whole
surface: aiohttp (HTTP/HTTPS), pyftpdlib (FTP), asyncssh (SCP/SFTP), fbtftp (TFTP), pysnmp (SNMPv3
USM incl. MD5/SHA × DES/AES-128), pyrad (RADIUS), dnslib (DNS); the NTP, TACACS+, and IPFIX
receivers are small, well-specified wire formats implemented directly. Go (B) has weaker FTP/TFTP
and SNMPv3-USM server libraries, forcing hand-rolled protocol code or sidecar containers. A pure
facade (C) fails R4 by construction — stock images cannot report what arrived — and per-service
sidecars multiply the R10 contract across languages. The one exception: Kafka (R3.7) is not fakeable
in-process at fidelity; it runs as a real broker container, the control layer records by consuming
from it, and R9 failure modes for it are applied by a TCP proxy in front of the broker. Compose ties
the process and the broker together (R7).
Reversibility: two-way door as long as the HTTP contract (R4, R5) stays stable — the acceptance
suite depends on the contract, not on the implementation. The contract (R11) is fixed and versioned
before implementation begins, so the cost of switching runtimes is bounded by rewriting the
implementation.
Revisit trigger: a protocol in phases 2–4 turns out to have no usable Python library — that receiver
moves into its own container behind the control layer (the Kafka pattern) and this entry is amended;
the decision as a whole is not reopened.

---

## D2 — Requirements live in the repository, not in a standalone document

Date: 2026-08-31
Context: the requirements existed as a single proposal file at the repository root, with no trace of
implementation anywhere near it. Requirements kept apart from code drift from it silently.
Options considered: (A) leave the proposal document as is; (B) move it into `docs/` and split it
into requirements / phases / decisions; (C) keep requirements in an external tracker.
Choice: **B**. `docs/requirements.md` is the source of truth for scope, carrying a readiness matrix
that is filled in only when the acceptance suite actually passes. A change in scope becomes a diff
rather than a verbal agreement.
Reversibility: two-way door.
Revisit trigger: if the customer requires scope to be tracked in their own system, move it there and
leave a pointer here.

---

## D0 — Repository starter scaffold

Date: 2026-08-31
Context: an empty repository with no commits, whose only input artifact was a test-harness proposal.
Options considered: (A) a documentation scaffold with no code; (B) documentation plus a phase 0 code
skeleton.
Choice: **A**, because D1 was open at the time and the customer's acceptance suite has not been
received — any code written then would have targeted an unapproved runtime with no definition of
done. (D1 has since been decided; code starts when the phase 0 inputs in the roadmap are closed.)
Reversibility: two-way door.
Revisit trigger: the "Before phase 0 starts" table in the roadmap is fully closed.
