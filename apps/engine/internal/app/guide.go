package app

import (
	"fmt"
	"sort"
	"strings"
)

// gettingStarted builds the first-run guide shown as the startup banner and
// served at GET / — so someone who just `docker run`s the image, or curls the
// control port, immediately sees what is running and what to do next.
func gettingStarted(lic LicenseStatus, endpoints map[string]string) string {
	var b strings.Builder
	line := func(s string) { b.WriteString(s + "\n") }

	line("")
	line("  ▪ ▪ ▪   Servienta engine · " + lic.Mode + " mode")
	line("  ▪ □ ▪   network integrations, verifiable")
	line("  ▪ ▪ ▪   stands: " + standsSummary(lic))
	line("")
	line("  Start here (host ports are whatever you mapped with -p):")
	line("    curl localhost:8080/                       this guide, any time")
	line("    curl localhost:8080/api/v1/endpoints       what's running, and where")
	line("    curl localhost:8080/api/v1/try             see the record→read loop work, one call")
	line("    curl localhost:8081/<file>                 serve a mounted fixture")
	line("")
	line("  Record traffic and read it back, scoped to a test run:")
	line("    curl -X PUT localhost:8080/api/v1/runs/run-1 -d '{\"sources\":[\"172.17.0.1\"]}'")
	line("    echo 'hello from my app' | nc localhost 9000")
	line("    curl 'localhost:8080/api/v1/received/reference?run=run-1'")
	line("    curl -X POST localhost:8080/api/v1/reset")
	line("")
	if lic.Mode == "free" {
		line("  Free mode runs the file server and a demo receiver. A license unlocks")
		line("  syslog, SNMP, RADIUS, TACACS+, DNS, NTP, Kafka, IPFIX — mount it and restart.")
		line("")
	}
	line("  Running now:")
	for _, name := range sortedKeys(endpoints) {
		line(fmt.Sprintf("    %-14s %s", name, endpoints[name]))
	}
	line("")
	line("  Docs: https://servienta.com/docs")
	line("")
	return b.String()
}

func standsSummary(lic LicenseStatus) string {
	if len(lic.Stands) == 0 {
		return "(none)"
	}
	return strings.Join(lic.Stands, ", ")
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
