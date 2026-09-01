package acceptance

import (
	"net"
	"testing"

	"github.com/vmware/go-ipfix/pkg/entities"
	"github.com/vmware/go-ipfix/pkg/exporter"
	"github.com/vmware/go-ipfix/pkg/registry"
)

// --- R3.8: export an IPFIX flow record, read it back via R4 ---
func TestIPFIXExport(t *testing.T) {
	registry.LoadRegistry()
	e := startEngine(t)
	e.do(t, "PUT", "/api/v1/runs/run-1", map[string]any{"sources": []string{"127.0.0.1"}})

	host, portStr, _ := net.SplitHostPort(e.endpoints["ipfix"])
	ep, err := exporter.InitExportingProcess(exporter.ExporterInput{
		CollectorAddress:    host + ":" + portStr,
		CollectorProtocol:   "udp",
		ObservationDomainID: 1,
		TempRefTimeout:      0,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer ep.CloseConnToCollector()

	templateID := ep.NewTemplateID()
	mkIE := func(name string, val []byte) entities.InfoElementWithValue {
		el, err := registry.GetInfoElement(name, registry.IANAEnterpriseID)
		if err != nil {
			t.Fatalf("element %s: %v", name, err)
		}
		ie, err := entities.DecodeAndCreateInfoElementWithValue(el, val)
		if err != nil {
			t.Fatalf("ie %s: %v", name, err)
		}
		return ie
	}

	// Template set (field definitions, no values).
	tset := entities.NewSet(false)
	tset.PrepareSet(entities.Template, templateID)
	tset.AddRecord([]entities.InfoElementWithValue{
		mkIE("sourceIPv4Address", nil),
		mkIE("destinationIPv4Address", nil),
	}, templateID)
	if _, err := ep.SendSet(tset); err != nil {
		t.Fatalf("send template: %v", err)
	}

	// Data set with one flow record.
	dset := entities.NewSet(false)
	dset.PrepareSet(entities.Data, templateID)
	dset.AddRecord([]entities.InfoElementWithValue{
		mkIE("sourceIPv4Address", net.ParseIP("1.2.3.4")),
		mkIE("destinationIPv4Address", net.ParseIP("5.6.7.8")),
	}, templateID)
	if _, err := ep.SendSet(dset); err != nil {
		t.Fatalf("send data: %v", err)
	}

	msgs := e.receivedN(t, "ipfix", "run-1", 1)
	fields := msgs[0]["content"].(map[string]any)["fields"].(map[string]any)
	if fields["sourceIPv4Address"] != "1.2.3.4" || fields["destinationIPv4Address"] != "5.6.7.8" {
		t.Fatalf("ipfix record not recorded correctly: %v", fields)
	}
}
