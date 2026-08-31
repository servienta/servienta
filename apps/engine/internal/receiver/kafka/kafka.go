// Package kafka runs an in-process simulated broker (kfake, D12) and records
// everything produced to it (R3.7) under the "kafka" service (R4). One-way:
// the application produces, the harness records — an internal consumer reads
// the broker and records each message. No broker container ships (D12).
package kafka

import (
	"context"
	"net"

	"github.com/servienta/servienta/apps/engine/internal/receiver"
	"github.com/twmb/franz-go/pkg/kfake"
	"github.com/twmb/franz-go/pkg/kgo"
)

const service = "kafka"

type Receiver struct{}

func (Receiver) Name() string        { return service }
func (Receiver) Endpoints() []string { return []string{"kafka"} }

func (Receiver) Start(ctx context.Context, addrs map[string]string, rec receiver.Recorder) (map[string]net.Addr, error) {
	// Bind an ephemeral TCP port ourselves so we can report it (R7), then hand
	// it to kfake.
	ln, err := net.Listen("tcp", addrs["kafka"])
	if err != nil {
		return nil, err
	}
	addr := ln.Addr().(*net.TCPAddr)
	ln.Close()

	cluster, err := kfake.NewCluster(
		kfake.Ports(addr.Port),
		kfake.SeedTopics(1, "servienta"),
	)
	if err != nil {
		return nil, err
	}
	go func() { <-ctx.Done(); cluster.Close() }()

	// Internal consumer: read the broker and record each produced message.
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(cluster.ListenAddrs()...),
		kgo.ConsumeTopics("servienta"),
	)
	if err != nil {
		cluster.Close()
		return nil, err
	}
	go func() { <-ctx.Done(); cl.Close() }()
	go consume(ctx, cl, rec)

	return map[string]net.Addr{"kafka": addr}, nil
}

func consume(ctx context.Context, cl *kgo.Client, rec receiver.Recorder) {
	for {
		fetches := cl.PollFetches(ctx)
		if fetches.IsClientClosed() || ctx.Err() != nil {
			return
		}
		if mode, _ := rec.Mode(service); mode == "drop" || mode == "refuse" {
			continue // R9: accept but do not record
		}
		fetches.EachRecord(func(r *kgo.Record) {
			_ = rec.Record(service, "kafka", map[string]any{
				"topic": r.Topic,
				"key":   string(r.Key),
				"value": string(r.Value),
			})
		})
	}
}
