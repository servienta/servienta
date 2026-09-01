# The delivered container (D20): engine + embedded management UI + user guide
# in one image. N7: reproducible build from a tagged source tree. Build
# context is the REPO ROOT — the SPA build imports docs/guide/*.md and
# scripts/walkthrough.sh from outside apps/servienta.
FROM node:24-alpine AS ui
WORKDIR /repo/apps/servienta/webui
COPY apps/servienta/webui/package.json ./
RUN npm install
COPY docs/guide /repo/docs/guide
COPY scripts /repo/scripts
COPY apps/servienta/webui /repo/apps/servienta/webui
RUN npm run build

FROM golang:1.26-alpine AS build
WORKDIR /src
COPY apps/servienta/go.mod apps/servienta/go.sum ./
RUN go mod download
COPY apps/servienta/ ./
COPY --from=ui /repo/apps/servienta/internal/webui/dist ./internal/webui/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /servienta ./cmd/servienta

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/servienta/servienta"
COPY --from=build /servienta /servienta
ENTRYPOINT ["/servienta"]
