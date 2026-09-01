// Package control is the engine's versioned HTTP contract (R11): run
// declarations and reads (R4), reset (R5), endpoint discovery (R7), version.
package control

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/servienta/servienta/apps/servienta/internal/core"
	"github.com/servienta/servienta/apps/servienta/internal/sender"
	"github.com/servienta/servienta/apps/servienta/internal/webui"
)

// ContractVersion is bumped on any breaking change to this surface (R11).
// 0.2.0: additive — PUT/DELETE /api/v1/license (D20: the UI ships in the
// engine, so the license is applied through the contract).
const ContractVersion = "0.2.0"

type Server struct {
	store     *core.Store
	endpoints map[string]string
	license   any // app.LicenseStatus, kept as any to avoid an import cycle
	senders   map[string]sender.Sender
	licPath   string // where PUT /api/v1/license persists the file
}

func New(store *core.Store, endpoints map[string]string, licenseStatus any, senders map[string]sender.Sender, licensePath string) *Server {
	return &Server{store: store, endpoints: endpoints, license: licenseStatus, senders: senders, licPath: licensePath}
}

func (s *Server) Start(ctx context.Context, addr string) (net.Addr, error) {
	return serve(ctx, addr, s.handler())
}

// StartUI serves the embedded management SPA (D20) on its own port. The SPA
// is a browser client of the contract, so /api/* is answered by the same
// handler the API port uses — same origin for the browser, no second truth.
func (s *Server) StartUI(ctx context.Context, addr string) (net.Addr, error) {
	api := s.handler()
	spaFS := webui.FS()
	spa := http.FileServer(http.FS(spaFS))
	mux := http.NewServeMux()
	mux.Handle("/api/", api)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if p == "/" {
			p = "/index.html"
		}
		if _, err := fs.Stat(spaFS, p[1:]); err != nil {
			r.URL.Path = "/" // deep links fall back to index.html
		}
		spa.ServeHTTP(w, r)
	})
	return serve(ctx, addr, mux)
}

func serve(ctx context.Context, addr string, h http.Handler) (net.Addr, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	srv := &http.Server{Handler: h}
	go func() {
		<-ctx.Done()
		srv.Close()
	}()
	go srv.Serve(ln)
	return ln.Addr(), nil
}

func (s *Server) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"contract": ContractVersion})
	})
	mux.HandleFunc("GET /api/v1/endpoints", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, s.endpoints)
	})
	mux.HandleFunc("GET /api/v1/license", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, s.license)
	})
	// Reload re-reads the license by restarting in place: the engine still
	// validates the license only at startup (R12), so applying a new one
	// means starting up again.
	mux.HandleFunc("POST /api/v1/reload", func(w http.ResponseWriter, r *http.Request) {
		restart(w)
	})
	// D20: the UI ships inside the engine, so the license arrives through the
	// contract. The body is stored as-is and verified at the restart that
	// applies it — this endpoint grants no signing power (D10).
	mux.HandleFunc("PUT /api/v1/license", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil || !json.Valid(body) {
			writeError(w, http.StatusBadRequest, "body must be the license JSON as issued")
			return
		}
		if err := os.MkdirAll(filepath.Dir(s.licPath), 0o755); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if err := os.WriteFile(s.licPath, body, 0o644); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		restart(w)
	})
	mux.HandleFunc("DELETE /api/v1/license", func(w http.ResponseWriter, r *http.Request) {
		if err := os.Remove(s.licPath); err != nil && !os.IsNotExist(err) {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		restart(w)
	})
	mux.HandleFunc("POST /api/v1/reset", func(w http.ResponseWriter, r *http.Request) {
		s.store.Reset()
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("PUT /api/v1/runs/{id}", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Sources []string `json:"sources"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Sources) == 0 {
			writeError(w, http.StatusBadRequest, "body must be {\"sources\": [\"<ip>\", ...]}")
			return
		}
		switch err := s.store.DeclareRun(r.PathValue("id"), body.Sources); {
		case errors.Is(err, core.ErrClaimConflict), errors.Is(err, core.ErrRunExists):
			writeError(w, http.StatusConflict, err.Error())
		case err != nil:
			writeError(w, http.StatusInternalServerError, err.Error())
		default:
			w.WriteHeader(http.StatusCreated)
		}
	})
	mux.HandleFunc("DELETE /api/v1/runs/{id}", func(w http.ResponseWriter, r *http.Request) {
		if err := s.store.ReleaseRun(r.PathValue("id")); err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	// R2: force a file fault for a named fixture (instance-wide, D3).
	mux.HandleFunc("PUT /api/v1/faults/files/{fixture...}", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Kind string `json:"kind"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "body must be {\"kind\": \"...\"}")
			return
		}
		switch core.FileFaultKind(body.Kind) {
		case core.FileAuthReject, core.FileMissing, core.FileTruncate, core.FileCorrupt:
			s.store.Faults.SetFileFault(r.PathValue("fixture"), core.FileFaultKind(body.Kind))
			w.WriteHeader(http.StatusNoContent)
		default:
			writeError(w, http.StatusBadRequest, "unknown fault kind: "+body.Kind)
		}
	})
	// R13: send a protocol message to a target the caller names (D18).
	mux.HandleFunc("POST /api/v1/send/{service}", func(w http.ResponseWriter, r *http.Request) {
		svc := r.PathValue("service")
		snd, ok := s.senders[svc]
		if !ok {
			writeError(w, http.StatusNotFound, "no licensed sender for service "+svc)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "body must be a JSON object")
			return
		}
		target, _ := body["target"].(string)
		if target == "" {
			writeError(w, http.StatusBadRequest, "target (host:port) is required")
			return
		}
		result, err := snd.Send(r.Context(), target, body)
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, result)
	})
	// R8: control the reply of a request-response service (instance-wide, D3).
	mux.HandleFunc("PUT /api/v1/responses/{service}", func(w http.ResponseWriter, r *http.Request) {
		var spec map[string]any
		if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
			writeError(w, http.StatusBadRequest, "body must be a JSON object")
			return
		}
		s.store.Responses.Set(r.PathValue("service"), spec)
		w.WriteHeader(http.StatusNoContent)
	})
	// R9: put a receiver into a failure mode (instance-wide, D3).
	mux.HandleFunc("PUT /api/v1/faults/receivers/{service}", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Mode    string `json:"mode"`
			DelayMs int    `json:"delay_ms"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "body must be {\"mode\": \"...\", \"delay_ms\": N}")
			return
		}
		switch core.ReceiverMode(body.Mode) {
		case core.ModeNormal, core.ModeRefuse, core.ModeDrop, core.ModeDelay, core.ModeCut, core.ModeProtoEr:
			s.store.Faults.SetReceiverMode(r.PathValue("service"), core.ReceiverMode(body.Mode), body.DelayMs)
			w.WriteHeader(http.StatusNoContent)
		default:
			writeError(w, http.StatusBadRequest, "unknown receiver mode: "+body.Mode)
		}
	})
	mux.HandleFunc("GET /api/v1/received/{service}", func(w http.ResponseWriter, r *http.Request) {
		msgs, err := s.store.Received(r.PathValue("service"), r.URL.Query().Get("run"))
		switch {
		case errors.Is(err, core.ErrUnknownService):
			writeError(w, http.StatusNotFound, err.Error()) // N5: not an empty 200
		case errors.Is(err, core.ErrUnknownRun):
			writeError(w, http.StatusNotFound, err.Error()) // R4: not an empty 200
		case err != nil:
			writeError(w, http.StatusInternalServerError, err.Error())
		default:
			writeJSON(w, http.StatusOK, msgs)
		}
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})
	// An unknown API path is a visible error — a JSON client must not get an
	// empty 404 or an HTML page for a typo'd path (N5 in spirit).
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "unknown contract path: "+r.URL.Path+" (see GET /api/v1/endpoints, /api/v1/version)")
	})
	return mux
}

// restart acknowledges the request, then re-executes the binary in place so
// the startup license validation (R12) runs again — no docker restart policy
// required, a plain `docker run --rm` survives it. If exec fails, exit and
// let whatever supervises the container bring it back.
func restart(w http.ResponseWriter) {
	w.WriteHeader(http.StatusAccepted)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	go func() {
		time.Sleep(200 * time.Millisecond)
		if exe, err := os.Executable(); err == nil {
			_ = syscall.Exec(exe, os.Args, os.Environ())
		}
		os.Exit(0)
	}()
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
