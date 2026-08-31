// The Servienta engine (docs/requirements.md). Configuration is environment
// variables with the defaults from docs/protocol-parameters.md; endpoints are
// reported machine-readably on stdout and at GET /api/v1/endpoints (R7).
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/servienta/servienta/apps/engine/internal/app"
)

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	cfg := app.Config{
		ControlAddr:    envOr("SERVIENTA_CONTROL_ADDR", ":8080"),
		FilesHTTPAddr:  envOr("SERVIENTA_FILES_HTTP_ADDR", ":8081"),
		FilesHTTPSAddr: envOr("SERVIENTA_FILES_HTTPS_ADDR", ":8443"),
		FilesFTPAddr:   envOr("SERVIENTA_FILES_FTP_ADDR", ":2121"),
		FilesTFTPAddr:  envOr("SERVIENTA_FILES_TFTP_ADDR", ":6969"),
		FilesSCPAddr:   envOr("SERVIENTA_FILES_SCP_ADDR", ":2222"),
		FilesUser:      envOr("SERVIENTA_FILES_USER", "servienta"),
		FilesPassword:  envOr("SERVIENTA_FILES_PASSWORD", "throwaway-not-a-secret"),
		ReferenceAddr:  envOr("SERVIENTA_REFERENCE_ADDR", ":9000"),
		SyslogUDPAddr:  envOr("SERVIENTA_SYSLOG_UDP_ADDR", ":5514"),
		SyslogTCPAddr:  envOr("SERVIENTA_SYSLOG_TCP_ADDR", ":5515"),
		SyslogRELPAddr: envOr("SERVIENTA_SYSLOG_RELP_ADDR", ":5516"),
		SNMPTrapAddr:   envOr("SERVIENTA_SNMP_ADDR", ":1162"),
		SNMPCommunity:  envOr("SERVIENTA_SNMP_COMMUNITY", "throwaway-public"),
		FixturesDir:    envOr("SERVIENTA_FIXTURES", "/fixtures"),
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	a, err := app.Start(ctx, cfg)
	if err != nil {
		slog.Error("start", "err", err)
		os.Exit(1)
	}
	json.NewEncoder(os.Stdout).Encode(map[string]any{"endpoints": a.Endpoints})
	<-ctx.Done()
}
