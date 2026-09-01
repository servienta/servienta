#!/usr/bin/env bash
# Servienta walkthrough — run this against a running Servienta to see what it does.
#
#   1. start Servienta (free image is enough):
#        mkdir -p fixtures && printf 'link down eth0\n' > fixtures/hello.txt
#        docker run --rm -p 5001:5001 -p 8080:8080 -p 9000:9000 \
#          -v "$PWD/fixtures:/fixtures:ro" ghcr.io/servienta/servienta:latest
#   2. in another terminal:
#        ./walkthrough.sh   (this file — save it anywhere, chmod +x)
#
# Override hosts if you mapped different ports: API=host:port FILES=host:port REF=port
set -euo pipefail
API="${API:-localhost:5001}"; FILES="${FILES:-localhost:8080}"; REF="${REF:-9000}"

G='\033[1;32m'; C='\033[36m'; D='\033[2m'; R='\033[0m'
say(){ printf "\n${G}# %s${R}\n" "$*"; }
do_(){ printf "${C}\$ %s${R}\n" "$*"; eval "$*"; echo; }

command -v curl >/dev/null || { echo "curl is required"; exit 1; }
curl -sf "http://$API/api/v1/version" >/dev/null 2>&1 || {
  printf "${R}Servienta not reachable at %s — start it first (see the header of this script)\n" "$API"; exit 1; }

say "1 · what's running, and where  (host ports are whatever you mapped)"
do_ "curl -s http://$API/api/v1/endpoints"

say "2 · the file server returns your fixtures byte-for-byte"
do_ "curl -s http://$FILES/hello.txt"

say "3 · declare a test run, send it traffic, read back exactly what arrived"
do_ "curl -s -X PUT http://$API/api/v1/runs/run-1 -d '{\"sources\":[\"127.0.0.1\"]}' -o /dev/null -w 'run declared: %{http_code}\n'"
printf "${C}\$ printf 'auth failure on port 3\\\\n' | nc -w1 localhost $REF${R}\n"
printf 'auth failure on port 3\n' | nc -w1 localhost "$REF" || true
echo
do_ "curl -s http://$API/api/v1/received/reference"

say "4 · inject a fault — the transfer fails at protocol level, not with wrong bytes"
do_ "curl -s -X PUT http://$API/api/v1/faults/files/hello.txt -d '{\"kind\":\"missing\"}' -o /dev/null -w 'fault set: %{http_code}\n'"
do_ "curl -s -o /dev/null -w 'GET hello.txt now: %{http_code}\n' http://$FILES/hello.txt"

say "5 · reset returns everything to a known state — run your suite again"
do_ "curl -s -X POST http://$API/api/v1/reset -o /dev/null -w 'reset: %{http_code}\n'"
do_ "curl -s -o /dev/null -w 'hello.txt restored: %{http_code}\n' http://$FILES/hello.txt"

printf "${G}# done.${R} ${D}full reference: https://servienta.com/docs${R}\n\n"
