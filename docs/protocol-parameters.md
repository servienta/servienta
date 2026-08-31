---
title: Protocol Parameters
type: reference
status: active
updated: 2026-08-31
---

# Protocol parameters

The canonical parameter set every engine instance is configured with (phase 0 input #3, roadmap).
All values are supplied by configuration (env vars, R1) with these defaults. **Every credential
below is throwaway by construction (N6): fixed, publicly known, documented here precisely so it can
never be mistaken for a real one.**

| Service | Parameter | Env var | Default |
| --- | --- | --- | --- |
| Control API | listen address | `SERVIENTA_CONTROL_ADDR` | `:8080` |
| Files: HTTP | listen address | `SERVIENTA_FILES_HTTP_ADDR` | `:8081` |
| Files: HTTPS | listen address / cert | `SERVIENTA_FILES_HTTPS_ADDR` | `:8443`, self-signed throwaway cert generated at startup |
| Files: HTTPS, FTP, SCP | username / password | `SERVIENTA_FILES_USER` / `SERVIENTA_FILES_PASSWORD` | `servienta` / `throwaway-not-a-secret` |
| Files: FTP | listen address | `SERVIENTA_FILES_FTP_ADDR` | `:2121` |
| Files: TFTP | listen address (no auth in protocol) | `SERVIENTA_FILES_TFTP_ADDR` | `:6969/udp` |
| Files: SCP | listen address / host key | `SERVIENTA_FILES_SCP_ADDR` | `:2222`, throwaway host key generated at startup |
| Fixture tree | mount path | `SERVIENTA_FIXTURES` | `/fixtures` (read-only) |
| Reference receiver | listen address | `SERVIENTA_REFERENCE_ADDR` | `:9000` |
| Syslog | UDP/TCP/RELP addresses | `SERVIENTA_SYSLOG_*_ADDR` | `:5514` per transport |
| SNMP traps | v2c community | `SERVIENTA_SNMP_COMMUNITY` | `throwaway-public` |
| SNMP traps | USM users | `SERVIENTA_SNMP_USM_*` | `usm-md5-des`, `usm-md5-aes`, `usm-sha-des`, `usm-sha-aes`; auth pass `throwaway-auth`, priv pass `throwaway-priv` |
| RADIUS | shared secret | `SERVIENTA_RADIUS_SECRET` | `throwaway-radius` |
| TACACS+ | shared secret | `SERVIENTA_TACACS_SECRET` | `throwaway-tacacs` |
| DNS / NTP | listen addresses | `SERVIENTA_DNS_ADDR` / `SERVIENTA_NTP_ADDR` | `:5353/udp` / `:1123/udp` |
| Kafka | listen address / default topic | `SERVIENTA_KAFKA_ADDR` / topic | `:9092` / any (recorded per topic) |
| IPFIX | collector address | `SERVIENTA_IPFIX_ADDR` | `:4739/udp` |

In-container ports are fixed (above); **host ports are never fixed** (N4) — compose publishes them
ephemerally, and `GET /api/v1/endpoints` reports the live address of every service (R7).
