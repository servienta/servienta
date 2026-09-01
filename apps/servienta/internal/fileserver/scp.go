package fileserver

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/gliderlabs/ssh"
)

// StartSCP serves single-file downloads over SCP source mode ("scp -f") with
// password auth (R1). The host key is generated at startup — throwaway by
// construction (N6). The tree is read-only, so sink mode ("-t") is refused.
func StartSCP(ctx context.Context, addr, dir, user, password string) (net.Addr, error) {
	srv := &ssh.Server{
		Handler: func(s ssh.Session) { handleSCP(s, dir) },
		PasswordHandler: func(sctx ssh.Context, pass string) bool {
			return sctx.User() == user && pass == password
		},
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		srv.Close()
	}()
	go srv.Serve(ln)
	return ln.Addr(), nil
}

func handleSCP(s ssh.Session, root string) {
	args := s.Command()
	var target string
	source := false
	for i, a := range args {
		switch a {
		case "-f":
			source = true
		default:
			if i > 0 && a[0] != '-' {
				target = a
			}
		}
	}
	if len(args) == 0 || args[0] != "scp" || !source || target == "" {
		fmt.Fprintln(s.Stderr(), "only scp -f <path> is supported")
		s.Exit(1)
		return
	}
	f, err := os.Open(securePath(root, target))
	if err != nil {
		fmt.Fprintln(s.Stderr(), "no such file")
		s.Exit(1)
		return
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil || st.IsDir() {
		s.Exit(1)
		return
	}
	ack := make([]byte, 1)
	if _, err := s.Read(ack); err != nil || ack[0] != 0 {
		s.Exit(1)
		return
	}
	fmt.Fprintf(s, "C0644 %d %s\n", st.Size(), filepath.Base(target))
	if _, err := s.Read(ack); err != nil || ack[0] != 0 {
		s.Exit(1)
		return
	}
	if _, err := io.Copy(s, f); err != nil { // streams (R6)
		s.Exit(1)
		return
	}
	s.Write([]byte{0})
	s.Read(ack)
	s.Exit(0)
}
