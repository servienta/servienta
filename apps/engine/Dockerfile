# N7: reproducible build from a tagged source tree; the engine is the
# delivery's only container (D12).
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /servienta ./cmd/servienta

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/servienta/servienta"
COPY --from=build /servienta /servienta
ENTRYPOINT ["/servienta"]
