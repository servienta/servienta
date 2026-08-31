// Package app wires the engine together: core store, control API, file
// server, and every registered receiver. main() and the acceptance suite both
// start the engine through this package — the suite still talks to it only
// over the network.
package app

import (
	"context"
	"fmt"
	"net"

	"github.com/gosnmp/gosnmp"
	"github.com/servienta/servienta/apps/engine/internal/control"
	"github.com/servienta/servienta/apps/engine/internal/core"
	"github.com/servienta/servienta/apps/engine/internal/fileserver"
	"github.com/servienta/servienta/apps/engine/internal/receiver"
	dnsrecv "github.com/servienta/servienta/apps/engine/internal/receiver/dns"
	"github.com/servienta/servienta/apps/engine/internal/receiver/ipfix"
	"github.com/servienta/servienta/apps/engine/internal/receiver/kafka"
	"github.com/servienta/servienta/apps/engine/internal/receiver/ntp"
	"github.com/servienta/servienta/apps/engine/internal/receiver/radius"
	"github.com/servienta/servienta/apps/engine/internal/receiver/reference"
	"github.com/servienta/servienta/apps/engine/internal/receiver/snmptrap"
	"github.com/servienta/servienta/apps/engine/internal/receiver/syslog"
	"github.com/servienta/servienta/apps/engine/internal/receiver/tacacs"
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
	SyslogUDPAddr  string
	SyslogTCPAddr  string
	SyslogRELPAddr string
	SNMPTrapAddr   string
	SNMPCommunity  string
	RADIUSAddr     string
	RADIUSSecret   string
	TACACSAddr     string
	TACACSSecret   string
	DNSAddr        string
	NTPAddr        string
	KafkaAddr      string
	IPFIXAddr      string
	FixturesDir    string
	LicensePath    string   // mounted license file; absent => Free mode
	LicensePubKey  string   // embedded Ed25519 public key (base64)
	LicensedStands []string // test override; when set, skips file resolution
}

type App struct {
	Endpoints map[string]string // service/endpoint -> host:port (R7)
	License   LicenseStatus
	cancel    context.CancelFunc
}

func receivers(cfg Config) []receiver.Receiver {
	return []receiver.Receiver{
		reference.Receiver{},
		syslog.Receiver{},
		snmptrap.Receiver{Cfg: snmptrap.Config{
			Community: cfg.SNMPCommunity,
			USMUsers:  usmUsers(),
		}},
		radius.Receiver{Secret: cfg.RADIUSSecret},
		tacacs.Receiver{Secret: cfg.TACACSSecret},
		dnsrecv.Receiver{},
		ntp.Receiver{},
		kafka.Receiver{},
		ipfix.Receiver{},
	}
}

// usmUsers seeds the four MD5/SHA × DES/AES-128 USM users (R3.2, throwaway N6).
func usmUsers() []gosnmp.UsmSecurityParameters {
	mk := func(name string, auth gosnmp.SnmpV3AuthProtocol, priv gosnmp.SnmpV3PrivProtocol) gosnmp.UsmSecurityParameters {
		return gosnmp.UsmSecurityParameters{
			UserName:                 name,
			AuthenticationProtocol:   auth,
			AuthenticationPassphrase: "throwaway-auth",
			PrivacyProtocol:          priv,
			PrivacyPassphrase:        "throwaway-priv",
		}
	}
	return []gosnmp.UsmSecurityParameters{
		mk("usm-md5-des", gosnmp.MD5, gosnmp.DES),
		mk("usm-md5-aes", gosnmp.MD5, gosnmp.AES),
		mk("usm-sha-des", gosnmp.SHA, gosnmp.DES),
		mk("usm-sha-aes", gosnmp.SHA, gosnmp.AES),
	}
}

func receiverAddrs(cfg Config, labels []string) map[string]string {
	m := map[string]string{}
	for _, l := range labels {
		switch l {
		case "reference":
			m[l] = cfg.ReferenceAddr
		case "syslog-udp":
			m[l] = cfg.SyslogUDPAddr
		case "syslog-tcp":
			m[l] = cfg.SyslogTCPAddr
		case "syslog-relp":
			m[l] = cfg.SyslogRELPAddr
		case "snmp-traps":
			m[l] = cfg.SNMPTrapAddr
		case "radius":
			m[l] = cfg.RADIUSAddr
		case "tacacs":
			m[l] = cfg.TACACSAddr
		case "dns":
			m[l] = cfg.DNSAddr
		case "ntp":
			m[l] = cfg.NTPAddr
		case "kafka":
			m[l] = cfg.KafkaAddr
		case "ipfix":
			m[l] = cfg.IPFIXAddr
		default:
			m[l] = ":0"
		}
	}
	return m
}

func Start(parent context.Context, cfg Config) (*App, error) {
	ctx, cancel := context.WithCancel(parent)
	store := core.NewStore()
	endpoints := map[string]string{}

	lic := LicenseStatus{Mode: "licensed", Stands: cfg.LicensedStands}
	if cfg.LicensedStands == nil {
		lic = resolveLicense(cfg.LicensePath, cfg.LicensePubKey)
	}
	granted := map[string]bool{}
	for _, s := range lic.Stands {
		granted[s] = true
	}

	files := []struct {
		stand string
		name  string
		start func() (net.Addr, error)
	}{
		{"http", "files-http", func() (net.Addr, error) {
			return fileserver.StartHTTP(ctx, cfg.FilesHTTPAddr, cfg.FixturesDir, store.Faults)
		}},
		{"https", "files-https", func() (net.Addr, error) {
			return fileserver.StartHTTPS(ctx, cfg.FilesHTTPSAddr, cfg.FixturesDir, cfg.FilesUser, cfg.FilesPassword)
		}},
		{"ftp", "files-ftp", func() (net.Addr, error) {
			return fileserver.StartFTP(ctx, cfg.FilesFTPAddr, cfg.FixturesDir, cfg.FilesUser, cfg.FilesPassword)
		}},
		{"tftp", "files-tftp", func() (net.Addr, error) { return fileserver.StartTFTP(ctx, cfg.FilesTFTPAddr, cfg.FixturesDir) }},
		{"scp", "files-scp", func() (net.Addr, error) {
			return fileserver.StartSCP(ctx, cfg.FilesSCPAddr, cfg.FixturesDir, cfg.FilesUser, cfg.FilesPassword)
		}},
	}
	for _, sv := range files {
		if !granted[sv.stand] {
			continue // not licensed (D15): this stand simply does not start
		}
		addr, err := sv.start()
		if err != nil {
			cancel()
			return nil, fmt.Errorf("%s: %w", sv.name, err)
		}
		endpoints[sv.name] = addr.String()
	}

	for _, r := range receivers(cfg) {
		// The reference receiver is the R10 worked example and the public demo
		// receiver — never a billable stand, so it always runs (even in Free mode).
		if r.Name() != "reference" && !granted[r.Name()] {
			continue // unlicensed receiver: not started
		}
		store.RegisterService(r.Name())
		addrs, err := r.Start(ctx, receiverAddrs(cfg, r.Endpoints()), store)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("receiver %s: %w", r.Name(), err)
		}
		for label, addr := range addrs {
			endpoints[label] = addr.String()
		}
	}

	ctl, err := control.New(store, endpoints, lic).Start(ctx, cfg.ControlAddr)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("control: %w", err)
	}
	endpoints["control"] = ctl.String()

	return &App{Endpoints: endpoints, License: lic, cancel: cancel}, nil
}

func (a *App) Close() { a.cancel() }
