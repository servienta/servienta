# Servienta demo

A one-file capability walkthrough against a running engine (the free image is
enough — it uses the file server and the demo receiver).

```bash
docker run -d -p 8080:8080 -p 8081:8081 -p 9000:9000 \
  -v "$PWD/demo/fixtures:/fixtures:ro" ghcr.io/servienta/engine:latest

./demo/demo.sh
```

It discovers endpoints, serves fixtures byte-for-byte, records and reads back a
run, injects a file fault, and resets — narrating each step. See the same loop
live at `GET /api/v1/try`, and the getting-started guide at `GET /`.
