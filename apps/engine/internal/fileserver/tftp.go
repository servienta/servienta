package fileserver

import (
	"context"
	"io"
	"net"
	"os"

	"github.com/pin/tftp/v3"
)

// StartTFTP serves the tree read-only over TFTP (no credentials in the
// protocol — R1 requires them only for HTTPS, FTP, and SCP).
func StartTFTP(ctx context.Context, addr, dir string) (net.Addr, error) {
	srv := tftp.NewServer(func(filename string, rf io.ReaderFrom) error {
		f, err := os.Open(securePath(dir, filename))
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = rf.ReadFrom(f) // streams; never the whole file in memory (R6)
		return err
	}, nil)
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		srv.Shutdown()
		conn.Close()
	}()
	go srv.Serve(conn.(*net.UDPConn))
	return conn.LocalAddr(), nil
}
