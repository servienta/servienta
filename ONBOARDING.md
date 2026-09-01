# Servienta — team onboarding

Start here. This page gets you from clone to a running stand and tells you
where everything lives. Deep dives: [docs/guide/](docs/guide/README.md) (user
guide), [docs/requirements.md](docs/requirements.md) (engine scope),
[docs/decisions.md](docs/decisions.md) (why things are the way they are).

## What we build

Servienta gives a test suite a controlled instance of the network services an
application talks to: a file server on five transports plus receivers for
syslog, SNMP traps, RADIUS, TACACS+, DNS, NTP, Kafka, IPFIX. The engine never
reaches into the application under test — it accepts, serves, and reports.

Three deployables, one repo (D20):

| App | What | Stack |
| --- | --- | --- |
| `apps/servienta` | The whole delivered product: data plane, versioned contract, embedded console + guide (`webui/`) | Go + Vue 3, ONE container |
| `apps/admin` | Vendor back office (licenses, customers) | Hono + Vue 3 on Cloudflare |
| `apps/web` | Marketing site + public docs | Static HTML on Cloudflare Workers |

## Run the stand

One container is the whole product — **no docker compose** (D20):

```bash
mkdir -p fixtures && echo hello > fixtures/hello.txt
docker run --rm -p 5000:5000 -p 5001:5001 -p 8080:8080 -p 9000:9000 \
  -v "$PWD/fixtures:/fixtures:ro" ghcr.io/servienta/servienta:latest
./scripts/walkthrough.sh   # exercises the whole loop, prints each step
```

Local dev build: `pnpm install && (cd apps/servienta/webui && pnpm run build)`
then `docker build -f apps/servienta/Dockerfile .` — **from the repo root**: the
SPA bakes in `docs/guide` and `scripts/walkthrough.sh`, so the image context
is the root — do not narrow it back to the app directory.

## ⚠ Port layout (changed 2026-09-01, D20 — supersedes every older note)

| Address | What |
| --- | --- |
| **http://localhost:5000** | Console (web UI): stand state, license, Getting started, the guide (Docs tab) |
| **localhost:5001** | The **contract** (`/api/v1/…`) — what tests curl |
| localhost:8080 | Files: HTTP (free tier) |
| localhost:9000 | Demo "reference" receiver (free tier) |
| 8443 / 2121+50000–50100 / 2222 / 6969·udp | Files: HTTPS / FTP(+passive) / SCP / TFTP |
| 5514·udp / 5515 / 5516 | Syslog UDP / TCP / RELP |
| 15353·udp | DNS (in-container default is 15353 so 1:1 works — mDNS owns 5353/udp on desktops) |
| 1123·udp / 1812·udp / 49 / 1162·udp / 4739·udp / 9092 | NTP / RADIUS / TACACS+ / SNMP traps / IPFIX / Kafka |

In-container ports are fixed; host ports are yours — map 1:1 normally, remap
for a second instance. Gotchas that cost us time already:

- Host traffic reaches the engine as the Docker gateway (`172.17.0.1`) — claim
  *that* source in run declarations, or read `/received/...` without `?run=`
  to see the real source.
- TFTP transfers reply from ephemeral ports and don't survive Docker NAT.
- Fixtures: any host dir, mounted `-v …:/fixtures:ro`; edits appear instantly.
- The license lives inside the container (`/license.json`), applied via
  `PUT /api/v1/license` (the engine restarts **in place** — plain
  `docker run --rm` survives it). Recreating the container = re-apply.

## Where the docs live (single source, three surfaces)

1. **`docs/guide/*.md`** — the canonical user guide.
2. **Console → Docs tab** — the same markdown, baked into the image at build
   time; editing the guide shows up there after an image rebuild. Same for
   `scripts/walkthrough.sh` (embedded on Getting started with a copy button).
3. **servienta.com/docs** — `apps/web/public/docs.html`, a condensed public
   version; deployed by CI on push to `main`.

## Rules that bite (short version)

- The engine contract is **versioned** (`GET /api/v1/version`, now 0.2.0) —
  breaking `/received`, `/reset`, runs, response controls, or failure modes
  bumps it and updates `apps/servienta/openapi.yaml` (invariant 9).
- Every new piece of state must be cleared by `POST /reset` in the same
  change that introduces it (invariant 3).
- A requirement is done when the **acceptance suite** covering it passes —
  then update the readiness matrix in `docs/requirements.md`.
- Non-trivial decisions get an entry at the top of `docs/decisions.md`
  (template inside). Read D20 for the single-container delivery above.
- The full list lives in [CLAUDE.md](CLAUDE.md) → Invariants; it applies to
  humans too.

## CI / deploy

Push to `main`: **CI** runs the engine suite, then builds and pushes the one
image `ghcr.io/servienta/servienta` (registry keeps the 5 newest versions);
**Deploy** ships `apps/web` and `apps/admin` to Cloudflare. Watch runs with
`gh run list`. The old `ghcr.io/servienta/{engine,console}` packages are
frozen leftovers — safe to delete in the GitHub UI.
