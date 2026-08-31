# Servienta

A provider of testing services: a controlled set of network endpoints and a file server that make
an application's integrations verifiable by automated tests rather than by reading code.

**Status:** requirements baselined, stack and topology decided (D0–D13); implementation started
with the marketing site and the admin panel (D13).

## Product

| Deployable | What it is | Where |
| --- | --- | --- |
| **Engine** | The test-services data plane: protocol receivers, five-transport file server, versioned HTTP contract | Go + Docker compose, runs on customer/CI infrastructure |
| **Console** | Management, settings, statistics | Vue 3 + Tailwind + Go API, one Docker service on :80, TLS at Cloudflare edge — console.servienta.com |
| **Admin** | Vendor back office: license issuance, customers, plans | Vue 3 + Hono on Cloudflare Workers + Cloudflare D1, single-user login — admin.servienta.com |
| **Marketing site** | Public website | Static on Workers — servienta.com, www.servienta.com |

The engine is a set of **services and processes**: endpoints that accept data from the application
under test or serve data to it, and a control layer that lets an automated test prepare the
environment, read observations off it, and return it to a known state. The application under test
and the system it talks to are **out of scope**: the engine never polls them, never configures
them, and never depends on them.

Isolation is two-level: an **instance** (one compose project — the unit that mutations like reset
and fault injection apply to) and a **run** (one test execution inside an instance, whose recorded
traffic is read back by run id). See D3/D4 in the decision log.

Everything starts on free tiers; the domain is the only unavoidable spend (D7).

## Navigation

| Document | Contents |
| --- | --- |
| [docs/requirements.md](docs/requirements.md) | Engine requirements R1–R11, N1–N7 with verification methods — source of truth |
| [docs/roadmap.md](docs/roadmap.md) | Phases 0–4, phase 0 acceptance criteria, inputs, risks |
| [docs/stack.md](docs/stack.md) | Stack: engine libraries, console, site, hosting, free-tier budget |
| [docs/decisions.md](docs/decisions.md) | Decision log D0–D9 |
| [CLAUDE.md](CLAUDE.md) | Working rules for this repository |

## Next step

Produce the three phase 0 inputs — the fixture tree, the executable acceptance suite, the protocol
parameters (see "Before phase 0 starts" in the roadmap). Once they exist, phase 0 begins with
`POST /reset`.
