// Package app wires the engine together: core store, control API, file
// server, and every registered receiver. main() and the acceptance suite both
// start the engine through this package — the suite still talks to it only
// over the network.
package app

import (
	"context"
	"fmt"

	"github.com/servienta/servienta/apps/servienta/internal/control"
	"github.com/servienta/servienta/apps/servienta/internal/core"
	"github.com/servienta/servienta/apps/servienta/internal/fileserver"
	"github.com/servienta/servienta/apps/servienta/internal/receiver"
	"github.com/servienta/servienta/apps/servienta/internal/receiver/reference"
)

type Config struct {
	ControlAddr   string // e.g. ":8080"; ":0" for ephemeral
	FilesHTTPAddr string
	ReferenceAddr string
	FixturesDir   string
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

	files, err := fileserver.StartHTTP(ctx, cfg.FilesHTTPAddr, cfg.FixturesDir)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("files-http: %w", err)
	}
	endpoints["files-http"] = files.String()

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
