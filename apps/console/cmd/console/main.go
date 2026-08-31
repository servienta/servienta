// The Servienta console (D11): a Go service in Docker on :8080 that serves the
// management SPA and talks to the engine through its versioned contract.
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/servienta/servienta/apps/console/internal/server"
	"github.com/servienta/servienta/apps/console/internal/webui"
)

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	h := server.New(server.Config{
		EngineBase: envOr("CONSOLE_ENGINE_BASE", "http://engine:8080"),
		SPA:        webui.FS(),
	}).Handler()
	addr := envOr("CONSOLE_ADDR", ":8080")
	slog.Info("console listening", "addr", addr)
	if err := http.ListenAndServe(addr, h); err != nil {
		slog.Error("serve", "err", err)
		os.Exit(1)
	}
}
