// Package fileserver serves the fixture tree (R1) without interpreting it
// (R6): stdlib streaming, bytes as they are. HTTP transport; HTTPS/FTP/TFTP/
// SCP follow in later iterations of phase 0.
package fileserver

import (
	"context"
	"net"
	"net/http"
)

func StartHTTP(ctx context.Context, addr, dir string) (net.Addr, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	srv := &http.Server{Handler: http.FileServer(http.Dir(dir))}
	go func() {
		<-ctx.Done()
		srv.Close()
	}()
	go srv.Serve(ln)
	return ln.Addr(), nil
}
