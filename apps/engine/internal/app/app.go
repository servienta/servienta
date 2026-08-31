// Package app wires the engine together: core store, control API, file
// server, and every registered receiver. main() and the acceptance suite both
// start the engine through this package — the suite still talks to it only
// over the network.
package app

import (
	"context"
	"fmt"
	"net"

	"github.com/servienta/servienta/apps/engine/internal/control"
	"github.com/servienta/servienta/apps/engine/internal/core"
	"github.com/servienta/servienta/apps/engine/internal/fileserver"
	"github.com/servienta/servienta/apps/engine/internal/receiver"
	"github.com/servienta/servienta/apps/engine/internal/receiver/reference"
)

type Config struct {
	ControlAddr    string // e.g. ":8080"; ":0" for ephemeral
	FilesHTTPAddr  string
	FilesHTTPSAddr string
	FilesFTPAddr   string
	FilesTFTPAddr  string
	FilesSCPAddr   string
	FilesUser      string // credentials for HTTPS, FTP, SCP (R1); throwaway (N6)
	FilesPassword  string
	ReferenceAddr  string
	FixturesDir    string
}

type App struct {
	Endpoints map[string]string // service -> host:port (R7)
	cancel    context.CancelFunc
}

// Receivers is the engine's receiver registry (R10): adding a protocol means
// appending its Receiver here — nothing in the core changes.
var Receivers = []receiver.Receiver{
	reference.Receiver{},
}

func receiverAddr(cfg Config, name string) string {
	switch name {
	case "reference":
		return cfg.ReferenceAddr
	default:
		return ":0"
	}
}

func Start(parent context.Context, cfg Config) (*App, error) {
	ctx, cancel := context.WithCancel(parent)
	store := core.NewStore()
	endpoints := map[string]string{}

	starters := []struct {
		name  string
		start func() (net.Addr, error)
	}{
		{"files-http", func() (net.Addr, error) { return fileserver.StartHTTP(ctx, cfg.FilesHTTPAddr, cfg.FixturesDir, store.Faults) }},
		{"files-https", func() (net.Addr, error) {
			return fileserver.StartHTTPS(ctx, cfg.FilesHTTPSAddr, cfg.FixturesDir, cfg.FilesUser, cfg.FilesPassword)
		}},
		{"files-ftp", func() (net.Addr, error) {
			return fileserver.StartFTP(ctx, cfg.FilesFTPAddr, cfg.FixturesDir, cfg.FilesUser, cfg.FilesPassword)
		}},
		{"files-tftp", func() (net.Addr, error) { return fileserver.StartTFTP(ctx, cfg.FilesTFTPAddr, cfg.FixturesDir) }},
		{"files-scp", func() (net.Addr, error) {
			return fileserver.StartSCP(ctx, cfg.FilesSCPAddr, cfg.FixturesDir, cfg.FilesUser, cfg.FilesPassword)
		}},
	}
	for _, sv := range starters {
		addr, err := sv.start()
		if err != nil {
			cancel()
			return nil, fmt.Errorf("%s: %w", sv.name, err)
		}
		endpoints[sv.name] = addr.String()
	}

	for _, r := range Receivers {
		store.RegisterService(r.Name())
		addr, err := r.Start(ctx, receiverAddr(cfg, r.Name()), store)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("receiver %s: %w", r.Name(), err)
		}
		endpoints[r.Name()] = addr.String()
	}

	ctl, err := control.New(store, endpoints).Start(ctx, cfg.ControlAddr)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("control: %w", err)
	}
	endpoints["control"] = ctl.String()

	return &App{Endpoints: endpoints, cancel: cancel}, nil
}

func (a *App) Close() { a.cancel() }
