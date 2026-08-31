// Package kafkasend produces a message to a receiving application's Kafka
// broker (R13).
package kafkasend

import (
	"context"
	"time"

	"github.com/servienta/servienta/apps/engine/internal/sender"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Sender struct{}

func (Sender) Name() string { return "kafka" }

func (Sender) Send(ctx context.Context, target string, payload map[string]any) (map[string]any, error) {
	cl, err := kgo.NewClient(kgo.SeedBrokers(target))
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res := cl.ProduceSync(cctx, &kgo.Record{
		Topic: sender.StrOr(payload, "topic", "servienta"),
		Key:   []byte(sender.Str(payload, "key")),
		Value: []byte(sender.Str(payload, "value")),
	})
	if err := res.FirstErr(); err != nil {
		return nil, err
	}
	r, _ := res.First()
	return map[string]any{"sent": true, "topic": r.Topic, "partition": r.Partition, "offset": r.Offset}, nil
}
