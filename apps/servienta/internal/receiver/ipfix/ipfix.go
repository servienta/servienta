// Package ipfix runs an IPFIX collector (R3.8) and records each received data
// record under the "ipfix" service (R4). One-way: the application exports
// flows, the harness records them.
package ipfix

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/servienta/servienta/apps/servienta/internal/receiver"
	"github.com/vmware/go-ipfix/pkg/collector"
	"github.com/vmware/go-ipfix/pkg/entities"
)

const service = "ipfix"

type Receiver struct{}

func (Receiver) Name() string        { return service }
func (Receiver) Endpoints() []string { return []string{"ipfix"} }

func (Receiver) Start(ctx context.Context, addrs map[string]string, rec receiver.Recorder) (map[string]net.Addr, error) {
	cp, err := collector.InitCollectingProcess(collector.CollectorInput{
		Address:       addrs["ipfix"],
		Protocol:      "udp",
		MaxBufferSize: 65535,
		TemplateTTL:   0,
	})
	if err != nil {
		return nil, err
	}
	go cp.Start()
	go func() {
		<-ctx.Done()
		cp.Stop()
	}()
	go func() {
		msgChan := cp.GetMsgChan()
		for msg := range msgChan {
			if mode, _ := rec.Mode(service); mode == "drop" || mode == "refuse" {
				continue
			}
			recordMessage(msg, rec)
		}
	}()

	// Start binds asynchronously; wait for the real bound address (R7).
	var addr net.Addr
	for i := 0; i < 200; i++ {
		if addr = cp.GetAddress(); addr != nil && addr.String() != "" {
			if _, port, _ := net.SplitHostPort(addr.String()); port != "0" && port != "" {
				break
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	if addr == nil {
		cp.Stop()
		return nil, fmt.Errorf("ipfix collector did not report a bound address")
	}
	return map[string]net.Addr{"ipfix": addr}, nil
}

func recordMessage(msg *entities.Message, rec receiver.Recorder) {
	set := msg.GetSet()
	if set == nil || set.GetSetType() != entities.Data {
		return // template sets carry no flow data
	}
	src := msg.GetExportAddress()
	for _, record := range set.GetRecords() {
		fields := map[string]any{}
		for _, ie := range record.GetOrderedElementList() {
			fields[ie.GetName()] = ieValue(ie)
		}
		_ = rec.Record(service, src, map[string]any{
			"obs_domain_id": msg.GetObsDomainID(),
			"fields":        fields,
		})
	}
}

func ieValue(ie entities.InfoElementWithValue) any {
	switch ie.GetInfoElement().DataType {
	case entities.Unsigned8:
		return ie.GetUnsigned8Value()
	case entities.Unsigned16:
		return ie.GetUnsigned16Value()
	case entities.Unsigned32:
		return ie.GetUnsigned32Value()
	case entities.Unsigned64:
		return ie.GetUnsigned64Value()
	case entities.String:
		return ie.GetStringValue()
	case entities.Ipv4Address, entities.Ipv6Address:
		return ie.GetIPAddressValue().String()
	default:
		return fmt.Sprintf("%v", ie.GetInfoElement().DataType)
	}
}
