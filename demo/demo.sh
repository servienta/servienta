#!/usr/bin/env bash
# Servienta capability demo. Runs against a running engine (free image is fine):
#   docker run -d -p 8080:8080 -p 8081:8081 -p 9000:9000 \
#     -v "$PWD/demo/fixtures:/fixtures:ro" ghcr.io/servienta/engine:latest
#   ./demo/demo.sh
set -euo pipefail
API="${API:-localhost:8080}"; FILES="${FILES:-localhost:8081}"; REF="${REF:-9000}"
step(){ printf '\n\033[1;32m# %s\033[0m\n' "$*"; }
run(){ printf '\033[36m$ %s\033[0m\n' "$*"; eval "$*"; }

step "1 · what's running, and where (R7)"
run "curl -s $API/api/v1/endpoints"

step "2 · serve the fixture tree byte-for-byte (R1/R6)"
run "curl -s $FILES/small.txt"
run "curl -s $FILES/large.bin | cmp -s - demo/fixtures/large.bin && echo 'large.bin: 1 MB, byte-identical'"

step "3 · one-call self-demo: record → read (R4)"
run "curl -s $API/api/v1/try"

step "4 · do it yourself: send traffic, read back what arrived (source + timestamp)"
run "printf 'auth failure on port 3\n' | nc -w1 localhost $REF"
run "curl -s $API/api/v1/received/reference"

step "5 · inject a fault — the file transfer must fail, not deliver wrong bytes (R2)"
run "curl -s -X PUT $API/api/v1/faults/files/small.txt -d '{\"kind\":\"missing\"}' -o /dev/null -w 'fault set: %{http_code}\n'"
run "curl -s -o /dev/null -w 'GET small.txt now: %{http_code}\n' $FILES/small.txt"

step "6 · reset to a known state — the suite can run again (R5)"
run "curl -s -X POST $API/api/v1/reset -o /dev/null -w 'reset: %{http_code}\n'"
run "curl -s -o /dev/null -w 'small.txt restored: %{http_code}\n' $FILES/small.txt"

step "done · every step ran against the real engine. offline, reproducible, one container."
