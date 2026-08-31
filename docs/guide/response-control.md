---
title: Response control (R8)
type: guide
updated: 2026-09-01
---

# Response control

For the request-response services (RADIUS, TACACS+, DNS, NTP) you set, for the
run, exactly what the engine answers. The default for every service is a
successful, valid reply. Controls are instance-wide and lifted by
`POST /api/v1/reset`.

Set a control:
```bash
curl -X PUT http://localhost:8080/api/v1/responses/<service> -d '<spec>'
```
Clear it (restore default): `PUT .../responses/<service> -d '{}'`.

## RADIUS
```json
{ "outcome": "accept" }                      // default
{ "outcome": "reject", "reason": "locked" }  // Access-Reject with a reply message
```

## TACACS+
```json
{ "outcome": "pass" }                     // default
{ "outcome": "fail", "reason": "denied" } // authentication FAIL
```

## DNS
```json
{ "outcome": "record", "address": "192.0.2.9" } // A record (default address if omitted)
{ "outcome": "nxdomain" }                        // NXDOMAIN
{ "outcome": "servfail" }                        // SERVFAIL
{ "outcome": "delay", "delay_ms": 500, "address": "192.0.2.9" } // delayed answer
```

## NTP
```json
{ "outcome": "normal" }                     // default: correct time, stratum 1
{ "outcome": "offset", "offset_ms": 5000 }  // time shifted by an offset
{ "outcome": "stratum", "stratum": 5 }      // altered stratum
{ "outcome": "refuse" }                     // refuse to serve (silence)
```

## Example: verify your login screen handles a reject
```bash
curl -X PUT .../api/v1/responses/radius -d '{"outcome":"reject","reason":"bad password"}'
# now drive your app's login; it should surface the failure
curl -X POST .../api/v1/reset   # back to accept
```
