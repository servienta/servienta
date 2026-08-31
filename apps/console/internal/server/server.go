// Package server is the console's HTTP layer: it serves the SPA and a thin
// API that talks to the engine ONLY through the engine's versioned contract
// (R11, D11) — the console is that contract's first client and never reaches
// into the engine's internals.
package server

import (
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"time"
)

type Config struct {
	// EngineBase is the engine control API base URL, e.g. http://engine:8080.
	EngineBase string
	// SPA is the built Vue app, embedded at build time.
	SPA fs.FS
}

type Server struct {
	cfg    Config
	engine *http.Client
}

func New(cfg Config) *Server {
	return &Server{cfg: cfg, engine: &http.Client{Timeout: 10 * time.Second}}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Console's own health — distinct from the engine's.
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "service": "console", "time": time.Now().UTC()})
	})

	// Engine reachability + endpoint map, proxied through the contract (R7/R11).
	mux.HandleFunc("GET /api/engine/endpoints", s.proxyGet("/api/v1/endpoints"))
	mux.HandleFunc("GET /api/engine/version", s.proxyGet("/api/v1/version"))
	mux.HandleFunc("GET /api/engine/license", s.proxyGet("/api/v1/license"))

	// Manage the whole stand: reset, received, runs — all via the contract.
	mux.HandleFunc("POST /api/engine/reset", func(w http.ResponseWriter, r *http.Request) {
		s.proxy(w, r, http.MethodPost, "/api/v1/reset", nil)
	})
	mux.HandleFunc("GET /api/engine/received/{service}", func(w http.ResponseWriter, r *http.Request) {
		q := ""
		if run := r.URL.Query().Get("run"); run != "" {
			q = "?run=" + run
		}
		s.proxy(w, r, http.MethodGet, "/api/v1/received/"+r.PathValue("service")+q, nil)
	})

	// SPA (deep links fall back to index.html).
	spa := http.FileServer(http.FS(s.cfg.SPA))
	mux.Handle("/", spaFallback(s.cfg.SPA, spa))
	return mux
}

func (s *Server) proxyGet(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { s.proxy(w, r, http.MethodGet, path, nil) }
}

func (s *Server) proxy(w http.ResponseWriter, _ *http.Request, method, path string, body io.Reader) {
	req, err := http.NewRequest(method, s.cfg.EngineBase+path, body)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	res, err := s.engine.Do(req)
	if err != nil {
		// N5 in spirit: the engine being unreachable is an explicit error.
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "engine unreachable: " + err.Error()})
		return
	}
	defer res.Body.Close()
	w.Header().Set("Content-Type", res.Header.Get("Content-Type"))
	w.WriteHeader(res.StatusCode)
	io.Copy(w, res.Body)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// spaFallback serves static files, falling back to index.html for client routes.
func spaFallback(spaFS fs.FS, fileServer http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fs.Stat(spaFS, trimLeadingSlash(r.URL.Path)); err != nil {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})
}

func trimLeadingSlash(p string) string {
	if p == "/" || p == "" {
		return "index.html"
	}
	return p[1:]
}

