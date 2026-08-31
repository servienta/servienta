// Package control is the engine's versioned HTTP contract (R11): run
// declarations and reads (R4), reset (R5), endpoint discovery (R7), version.
package control

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"

	"github.com/servienta/servienta/apps/engine/internal/core"
)

// ContractVersion is bumped on any breaking change to this surface (R11).
const ContractVersion = "0.1.0"

type Server struct {
	store     *core.Store
	endpoints map[string]string
}

func New(store *core.Store, endpoints map[string]string) *Server {
	return &Server{store: store, endpoints: endpoints}
}

func (s *Server) Start(ctx context.Context, addr string) (net.Addr, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	srv := &http.Server{Handler: s.handler()}
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
	return mux
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
