---
title: Servienta — Stack
type: reference
status: active
updated: 2026-08-31
---

# Stack

The technology baseline for the product. Decisions behind it: D6 (engine runtime), D7 (topology,
domains, free-first), D8 (console and site stack), D9 (data layer) in [decisions.md](decisions.md).

**Naming rule:** the Cloudflare database product is always written **Cloudflare D1**, fully
qualified — bare "D1" in this repository means the decision-log entry.

## Principle: free at start

Every component runs on a free tier. The only unavoidable spend is the domain registration
(~$10–12/yr at cost via Cloudflare Registrar). Anything that would introduce a recurring cost enters
the stack only through a decision-log entry. Free-tier ceilings and their upgrade triggers are
tabled at the bottom; verify current limits at signup — providers move them.

## Topology

Three deployables ([D7](decisions.md) as amended by [D20](decisions.md)):

| Deployable | What it is | Stack | Runs on | Domain |
| --- | --- | --- | --- | --- |
| **Engine** | The test-services data plane: receivers, file server, versioned HTTP contract (R1–R11), embedded console + guide | Go, one Docker container (`ghcr.io/servienta/servienta`) | Any container host — customer machine, CI agent. Works with no internet (N2) | none (customer-addressed) |
| **Admin** | Vendor back office: license issuance, customers, plans | Vue 3 SPA + Hono API on Workers + Cloudflare D1; single-user email+password auth (D14) | Cloudflare (free plan) | admin.servienta.com |
| **Marketing site** | Public website | Static HTML + Tailwind on Workers static assets | Cloudflare (free plan) | servienta.com, www.servienta.com |

The console talks to engines **only** through the engine's versioned HTTP contract (R11) — it is the
contract's first generated client. The engine never depends on the console: a self-hosted engine is
fully operable with the contract and its documentation alone (N2 demands this).

## Engine (Go)

- **Language:** Go, current stable (1.26+ at time of writing); modules; `golangci-lint`.
- **HTTP contract:** stdlib `net/http` (1.22+ routing); OpenAPI schema is authored first and shipped
  in the delivery (R11).
- **Images:** multi-stage build (SPA → static binary) → `scratch`; a single `docker run` is the
  delivery entry point (R7, D20). Published to ghcr.io as `ghcr.io/servienta/servienta` by CI.
- **Protocol library map** (re-verified for Go under D6):

| Surface | Library | Note |
| --- | --- | --- |
| HTTP/HTTPS file server | stdlib `net/http` | streaming, no full-file buffering (R6) |
| FTP | `fclairamb/ftpserverlib` | |
| TFTP | `pin/tftp` | |
| SCP | `gliderlabs/ssh` + `x/crypto/ssh` | |
| Syslog UDP/TCP | `mcuadros/go-syslog` | RELP: small TCP framing, implemented in-repo |
| SNMP traps | `gosnmp/gosnmp` | v2c; v3 USM MD5/SHA × DES/AES-128 |
| RADIUS | `layeh/radius` | |
| TACACS+ | `nwaples/tacplus` | server side thin; gaps implemented in-repo |
| DNS | `miekg/dns` | |
| NTP | in-repo | RFC 5905 server subset — small wire format |
| Kafka | in-process simulated broker: `twmb/franz-go/pkg/kfake` (BSD-3) | enough wire protocol for producer clients (ApiVersions, Metadata, Produce/Fetch); no broker container in the delivery — the engine is the only container (D12). Fallback for real-broker semantics: `apache/kafka` container (Apache-2.0); Redpanda is BSL — local dev only |
| IPFIX | `vmware/go-ipfix` | collector |

- **Logging:** stdlib `log/slog`, JSON. **Metrics:** Prometheus text format on the control endpoint
  — scraped later, free now.

## Console (built into the engine)

The management UI is part of the delivered container ([D20](decisions.md)): the engine serves the
embedded SPA on **:5000** (`SERVIENTA_UI_ADDR`), the contract stays on **:5001**. There is no
vendor-hosted console and no console.servienta.com — delivery is one `docker run`, full stop.

- **Frontend:** Vue 3 + TypeScript, Vite, Pinia, Vue Router, Tailwind CSS v4 — source at
  `apps/servienta/webui/`, built into `internal/webui/dist` and embedded with `go:embed`. The user
  guide (`docs/guide/*.md`, rendered with `markdown-it`) and the walkthrough script are baked into
  the bundle at image build time, so the stand documents itself offline (N2); the image builds
  from the repo-root context for exactly this reason.
- **Contract discipline:** the SPA is a browser client of `/api/v1` — the UI port answers `/api/*`
  with the same handler the API port uses, so there is no second truth and no CORS. Unknown
  `/api/*` paths return an explicit JSON 404.
- **State:** none. The license file lives inside the container (`/license.json`); a new license
  arrives via `PUT /api/v1/license` and is verified at the in-place restart that applies it.
- **Licensing hygiene:** the UI is delivered software, so the permissive-licenses rule
  (invariant 11) applies: Vue/Pinia/Vue Router/Tailwind/markdown-it — all MIT.

## Admin (admin.servienta.com)

A Workers app ([D13](decisions.md)) — ships independently of everything else, free tier.

- **API:** Hono on Cloudflare Workers, `zod` validation, OpenAPI via `@hono/zod-openapi`.
- **Frontend:** Vue 3 + TypeScript, Vite, Pinia, Vue Router, Tailwind CSS v4 — served as Workers
  static assets from the same deploy; API routes take precedence.
- **Auth:** single-user email+password ([D14](decisions.md)): PBKDF2-SHA-256 hash in Cloudflare
  D1, stateless HMAC-signed session cookie (SESSION_SECRET Worker secret), login/logout/change,
  and forgot/reset via a hashed one-time token emailed through Resend (disabled until
  RESEND_API_KEY is set; password change from the UI works regardless).
- **Data:** Cloudflare D1 under [D9](decisions.md)'s portability rules; Drizzle ORM migrations as
  the single schema source.
- **License issuance (D10/R12):** Ed25519 via WebCrypto in the Worker; the signing key is a
  Worker secret (`wrangler secret put`), never in the repository or any delivery.
- **License upload:** the customer console (not admin) accepts an issued license, writes it to a
  volume shared with the engine, and triggers an engine reload (restart) to apply it — the engine
  still verifies the signature at startup, so the console never gains signing power.
- **Deploys:** Wrangler; workers.dev URL until the servienta.com zone exists, then the custom
  domain attaches.

## Marketing site (servienta.com)

Static HTML + Tailwind served as Workers static assets (unlimited free requests); `www` redirects to
the apex. No framework until page count justifies one (Astro is the named candidate — revisit in
D8). Cloudflare Web Analytics (free, cookieless).

## Repository & delivery

Monorepo (this repository):

```
apps/servienta/     Go module — the engine + embedded console SPA (webui/), the delivered container
apps/admin/      vendor admin: Vue 3 SPA + Hono Worker API + Cloudflare D1 (admin.servienta.com)
apps/web/        marketing static site on Cloudflare Workers (servienta.com)
packages/        shared libraries (@servienta/*), extracted only when a second consumer exists
docs/            requirements, roadmap, decisions, stack
```

- **VCS/CI:** GitHub + GitHub Actions (free tier). JS tooling: pnpm workspaces + Turborepo (`turbo run build/deploy`).
- **Images:** ghcr.io (free for public images).
- **DNS/domains:** Cloudflare Registrar + DNS. Records: apex → marketing Worker, `www` → redirect,
  `admin` → admin Worker. (No `console` record — the console ships inside the container, D20.)
- **Transactional email** (auth flows, when needed): Resend free tier *(proposed)*; inbound via
  Cloudflare Email Routing (free).
- **Error tracking:** Sentry free tier *(proposed, optional at start)*.

## Licensing and delivery of the engine ([D10](decisions.md))

- **Mechanism:** a signed license file (Ed25519), mounted like configuration; the engine validates
  signature, expiry, and edition limits at startup, fully offline (N2). Public key embedded in the
  binary; the signing key is a Cloudflare Worker secret readable only by the admin Worker, never
  part of any delivery; licenses issued and managed from the admin panel (admin.servienta.com).
  See R12.
- **Free edition:** runs without a license file within documented limits — free at start (D7).
- **Distribution:** per-customer pull access to the ghcr.io image; air-gapped hosts get it as a `docker save` tar (D20).
- **Container runtime:** any OCI runtime — Docker Engine (Apache-2.0), Podman, containerd. Docker
  Desktop is never required; its per-company subscription terms are the customer's affair.
- **Dependency licenses:** the delivery redistributes what it contains — engine dependencies are
  permissive only (MIT/BSD/Apache-2.0); BSL/GPL/AGPL enters only via a decision-log entry.
- Honestly stated: offline licensing of self-hosted software is compliance clarity, not DRM.

## Free-tier budget and upgrade triggers

As of 2026-08 — verify at signup:

| Resource | Free ceiling | First upgrade trigger |
| --- | --- | --- |
| Workers (marketing site) | static asset requests uncounted; 100k req/day if dynamic endpoints appear | dynamic traffic near the ceiling → Workers Paid $5/mo |
| Cloudflare D1 (admin data) | 5 GB total, 100k row writes/day | vendor back office will not approach this for years; if it does → Postgres (D9) |
| R2 (if needed for artifacts) | 10 GB, zero egress fees | — |
| GitHub Actions | 2,000 min/mo (private repos) | CI time → self-hosted runner (free) |
| ghcr.io | free public images | private images → GitHub paid storage |
| Resend | ~3k emails/mo | volume → paid tier |
| Domain | **not free** — ~$10–12/yr | — |

The engine itself costs nothing to run: it is the customer's container host.
