// Package ipfixsend exports an IPFIX flow record to a receiving application's
// collector (R13): a source/destination IPv4 flow.
package ipfixsend

import (
	"context"
	"net"

	"github.com/servienta/servienta/apps/engine/internal/sender"
	"github.com/vmware/go-ipfix/pkg/entities"
	"github.com/vmware/go-ipfix/pkg/exporter"
	"github.com/vmware/go-ipfix/pkg/registry"
)

type Sender struct{}

func (Sender) Name() string { return "ipfix" }

func (Sender) Send(ctx context.Context, target string, payload map[string]any) (map[string]any, error) {
	registry.LoadRegistry()
	host, port := splitHostPort(target, "4739")
	ep, err := exporter.InitExportingProcess(exporter.ExporterInput{
		CollectorAddress:    host + ":" + port,
		CollectorProtocol:   "udp",
		ObservationDomainID: 1,
		TempRefTimeout:      0,
	})
	if err != nil {
		return nil, err
	}
	defer ep.CloseConnToCollector()

	tid := ep.NewTemplateID()
	mk := func(name string, val []byte) (entities.InfoElementWithValue, error) {
		el, err := registry.GetInfoElement(name, registry.IANAEnterpriseID)
		if err != nil {
			return nil, err
		}
		return entities.DecodeAndCreateInfoElementWithValue(el, val)
	}

	src, err := mk("sourceIPv4Address", nil)
	if err != nil {
		return nil, err
	}
	dst, _ := mk("destinationIPv4Address", nil)
	tset := entities.NewSet(false)
	tset.PrepareSet(entities.Template, tid)
	tset.AddRecord([]entities.InfoElementWithValue{src, dst}, tid)
	if _, err := ep.SendSet(tset); err != nil {
		return nil, err
	}

	sv, _ := mk("sourceIPv4Address", net.ParseIP(sender.StrOr(payload, "src", "1.2.3.4")))
	dv, _ := mk("destinationIPv4Address", net.ParseIP(sender.StrOr(payload, "dst", "5.6.7.8")))
	dset := entities.NewSet(false)
	dset.PrepareSet(entities.Data, tid)
	dset.AddRecord([]entities.InfoElementWithValue{sv, dv}, tid)
	if _, err := ep.SendSet(dset); err != nil {
		return nil, err
	}
	return map[string]any{"sent": true}, nil
}

func splitHostPort(target, def string) (string, string) {
	h, p, err := net.SplitHostPort(target)
	if err != nil {
		return target, def
	}
	return h, p
}
