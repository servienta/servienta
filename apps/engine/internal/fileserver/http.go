// Package fileserver serves the fixture tree (R1) without interpreting it
// (R6): stdlib streaming, bytes as they are. HTTP transport; the other four
// transports are in their own files. R2 file faults are honored here.
package fileserver

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/servienta/servienta/apps/engine/internal/core"
)

func StartHTTP(ctx context.Context, addr, dir string, faults *core.Faults) (net.Addr, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	srv := &http.Server{Handler: faultingFileHandler(dir, faults)}
	go func() {
		<-ctx.Done()
		srv.Close()
	}()
	go srv.Serve(ln)
	return ln.Addr(), nil
}

// faultingFileHandler serves files, applying any R2 fault set for the fixture.
func faultingFileHandler(dir string, faults *core.Faults) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if kind, ok := faults.FileFault(name); ok {
			switch kind {
			case core.FileAuthReject:
				http.Error(w, "authentication rejected", http.StatusUnauthorized)
				return
			case core.FileMissing:
				http.Error(w, "not found", http.StatusNotFound)
				return
			case core.FileCorrupt:
				serveCorrupt(w, dir, name)
				return
			case core.FileTruncate:
				serveTruncated(w, dir, name)
				return
			}
		}
		http.ServeFile(w, r, securePath(dir, name))
	})
}

func serveCorrupt(w http.ResponseWriter, dir, name string) {
	data, err := os.ReadFile(securePath(dir, name))
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if len(data) > 0 {
		data[0] ^= 0xFF // flip a byte: content corrupted, transfer "succeeds"
	}
	w.Write(data)
}

func serveTruncated(w http.ResponseWriter, dir, name string) {
	f, err := os.Open(securePath(dir, name))
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	defer f.Close()
	st, _ := f.Stat()
	// Declare the full length, then send only half and drop the connection —
	// the client sees a protocol-level short read, not a clean EOF.
	w.Header().Set("Content-Length", itoa(st.Size()))
	io.CopyN(w, f, st.Size()/2)
	if hj, ok := w.(http.Hijacker); ok {
		if conn, _, err := hj.Hijack(); err == nil {
			conn.Close()
		}
	}
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
