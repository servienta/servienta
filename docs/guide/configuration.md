---
title: Configuration
type: guide
updated: 2026-09-01
---

# Engine configuration

Every setting is an environment variable with a documented default. In-container
ports are fixed; host ports are yours to map (compose assigns them, and
`GET /api/v1/endpoints` reports the live addresses).

## Addresses

| Variable | Default | Service |
| --- | --- | --- |
| `SERVIENTA_CONTROL_ADDR` | `:8080` | Control API |
| `SERVIENTA_FILES_HTTP_ADDR` | `:8081` | Files: HTTP |
| `SERVIENTA_FILES_HTTPS_ADDR` | `:8443` | Files: HTTPS |
| `SERVIENTA_FILES_FTP_ADDR` | `:2121` | Files: FTP |
| `SERVIENTA_FILES_TFTP_ADDR` | `:6969` | Files: TFTP |
| `SERVIENTA_FILES_SCP_ADDR` | `:2222` | Files: SCP |
| `SERVIENTA_REFERENCE_ADDR` | `:9000` | Demo receiver |
| `SERVIENTA_SYSLOG_UDP_ADDR` | `:5514` | Syslog UDP |
| `SERVIENTA_SYSLOG_TCP_ADDR` | `:5515` | Syslog TCP |
| `SERVIENTA_SYSLOG_RELP_ADDR` | `:5516` | Syslog RELP |
| `SERVIENTA_SNMP_ADDR` | `:1162` | SNMP traps |
| `SERVIENTA_RADIUS_ADDR` | `:1812` | RADIUS |
| `SERVIENTA_TACACS_ADDR` | `:49` | TACACS+ |
| `SERVIENTA_DNS_ADDR` | `:5353` | DNS |
| `SERVIENTA_NTP_ADDR` | `:1123` | NTP |
| `SERVIENTA_KAFKA_ADDR` | `:9092` | Kafka |
| `SERVIENTA_IPFIX_ADDR` | `:4739` | IPFIX |

## Credentials (all throwaway, N6)

| Variable | Default | Used by |
| --- | --- | --- |
| `SERVIENTA_FILES_USER` | `servienta` | HTTPS, FTP, SCP |
| `SERVIENTA_FILES_PASSWORD` | `throwaway-not-a-secret` | HTTPS, FTP, SCP |
| `SERVIENTA_SNMP_COMMUNITY` | `throwaway-public` | SNMP v2c |
| `SERVIENTA_RADIUS_SECRET` | `throwaway-radius` | RADIUS |
| `SERVIENTA_TACACS_SECRET` | `throwaway-tacacs` | TACACS+ |

SNMP v3 USM users (auth pass `throwaway-auth`, priv pass `throwaway-priv`):
`usm-md5-des`, `usm-md5-aes`, `usm-sha-des`, `usm-sha-aes`.

**Every credential above is fixed, public, and documented so it can never be
mistaken for a real one. Never put a real secret in a Servienta config.**

## Fixtures and license

| Variable | Default | Purpose |
| --- | --- | --- |
| `SERVIENTA_FIXTURES` | `/fixtures` | Mounted fixture tree (read-only) |
| `SERVIENTA_LICENSE` | `/license.json` | Mounted license file (absent → free mode) |
| `SERVIENTA_LICENSE_PUBKEY` | (embedded) | Override the license public key |
