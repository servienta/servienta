// Package control is the engine's versioned HTTP contract (R11): run
// declarations and reads (R4), reset (R5), endpoint discovery (R7), version.
package control

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/servienta/servienta/apps/engine/internal/core"
	"github.com/servienta/servienta/apps/engine/internal/sender"
)

// ContractVersion is bumped on any breaking change to this surface (R11).
const ContractVersion = "0.1.0"

type Server struct {
	store     *core.Store
	endpoints map[string]string
	license   any // app.LicenseStatus, kept as any to avoid an import cycle
	senders   map[string]sender.Sender
	guide     string
}

func New(store *core.Store, endpoints map[string]string, licenseStatus any, senders map[string]sender.Sender, guide string) *Server {
	return &Server{store: store, endpoints: endpoints, license: licenseStatus, senders: senders, guide: guide}
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
	mux.HandleFunc("GET /api/v1/license", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, s.license)
	})
	// A one-call self-demo: declare a run, send a line to the demo receiver,
	// read it back, and narrate the round-trip. Lets a new user see the whole
	// record → read → reset loop work without wiring anything up.
	mux.HandleFunc("GET /api/v1/try", func(w http.ResponseWriter, r *http.Request) {
		s.runTry(w)
	})
	// Reload re-reads the mounted license by restarting: the engine still
	// validates the license only at startup (R12), so applying a new one means
	// starting up again. Docker's restart policy brings the process back.
	mux.HandleFunc("POST /api/v1/reload", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		go func() {
			time.Sleep(200 * time.Millisecond)
			os.Exit(0)
		}()
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
	// The first thing a new user curls: return the getting-started guide, not a
	// 404. Any unmatched path falls here; only "/" gets the guide.
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			writeError(w, http.StatusNotFound, "not found; see GET / for where to start")
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(s.guide))
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

// runTry drives the reference receiver end to end and reports each step.
func (s *Server) runTry(w http.ResponseWriter) {
	refAddr := s.endpoints["reference"]
	if refAddr == "" {
		writeError(w, http.StatusServiceUnavailable, "the demo receiver is not running")
		return
	}
	runID := "try-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	steps := []string{}

	// 1. Declare the run, claiming the loopback source we are about to send from.
	if err := s.store.DeclareRun(runID, []string{"127.0.0.1"}); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":   false,
			"note": "127.0.0.1 is already claimed by another run — try again once it releases, or use your own run id",
		})
		return
	}
	defer s.store.ReleaseRun(runID)
	steps = append(steps, "declared run "+runID+", claiming source 127.0.0.1")

	// 2. Send a line to the demo receiver (as your application would).
	_, port, _ := net.SplitHostPort(refAddr)
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", port), 2*time.Second)
	if err != nil {
		writeError(w, http.StatusBadGateway, "could not reach the demo receiver: "+err.Error())
		return
	}
	msg := "hello from the try endpoint"
	conn.Write([]byte(msg + "\n"))
	conn.Close()
	steps = append(steps, "sent \""+msg+"\" to the reference receiver on "+refAddr)

	// 3. Read it back, scoped to this run.
	var got []core.Message
	deadline := time.Now().Add(2 * time.Second)
	for {
		got, _ = s.store.Received("reference", runID)
		if len(got) > 0 || time.Now().After(deadline) {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	steps = append(steps, "read it back at GET /api/v1/received/reference?run="+runID)

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       len(got) > 0,
		"steps":    steps,
		"received": got,
		"next":     "now do it yourself: GET / shows the commands",
	})
}
