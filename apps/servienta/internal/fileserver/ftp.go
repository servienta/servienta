package fileserver

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"log/slog"
	"net"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/spf13/afero"
)

// StartFTP serves the tree read-only over FTP with password auth (R1).
func StartFTP(ctx context.Context, addr, dir, user, password string) (net.Addr, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	srv := ftpserver.NewFtpServer(&ftpDriver{
		listener: ln,
		fs:       afero.NewReadOnlyFs(afero.NewBasePathFs(afero.NewOsFs(), dir)),
		user:     user,
		password: password,
	})
	srv.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := srv.Listen(); err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		srv.Stop()
	}()
	go srv.Serve()
	return ln.Addr(), nil
}

type ftpDriver struct {
	listener net.Listener
	fs       afero.Fs
	user     string
	password string
}

func (d *ftpDriver) GetSettings() (*ftpserver.Settings, error) {
	return &ftpserver.Settings{
		Listener:                 d.listener,
		PassiveTransferPortRange: &ftpserver.PortRange{Start: 50000, End: 50100},
	}, nil
}

func (d *ftpDriver) ClientConnected(ftpserver.ClientContext) (string, error) {
	return "servienta", nil
}

func (d *ftpDriver) ClientDisconnected(ftpserver.ClientContext) {}

func (d *ftpDriver) AuthUser(_ ftpserver.ClientContext, user, pass string) (ftpserver.ClientDriver, error) {
	if user != d.user || pass != d.password {
		return nil, errors.New("authentication rejected")
	}
	return d.fs, nil
}

func (d *ftpDriver) GetTLSConfig() (*tls.Config, error) {
	return nil, errors.New("TLS is not configured")
}
