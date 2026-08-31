# Console: build the SPA, embed it into the Go binary, ship on scratch.
FROM node:24-alpine AS spa
WORKDIR /app
COPY package.json ./
RUN npm install
COPY . .
RUN npm run build

FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
COPY --from=spa /app/internal/webui/dist ./internal/webui/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /console ./cmd/console

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/servienta/servienta"
COPY --from=build /console /console
ENTRYPOINT ["/console"]
