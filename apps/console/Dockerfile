# Console: build the SPA (user guide and walkthrough baked in), embed it into
# the Go binary, ship on scratch. Build context is the REPO ROOT — the SPA
# imports docs/guide/*.md and scripts/walkthrough.sh from outside apps/console.
FROM node:24-alpine AS spa
WORKDIR /repo/apps/console
COPY apps/console/package.json ./
RUN npm install
COPY docs/guide /repo/docs/guide
COPY scripts /repo/scripts
COPY apps/console /repo/apps/console
RUN npm run build

FROM golang:1.26-alpine AS build
WORKDIR /src
COPY apps/console/go.mod ./
RUN go mod download
COPY apps/console/ ./
COPY --from=spa /repo/apps/console/internal/webui/dist ./internal/webui/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /console ./cmd/console

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/servienta/servienta"
COPY --from=build /console /console
ENTRYPOINT ["/console"]
