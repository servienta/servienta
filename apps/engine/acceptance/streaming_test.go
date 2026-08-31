package acceptance

import (
	"testing"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

// --- R3.7: produce to Kafka, read it back via R4 ---
func TestKafkaProduce(t *testing.T) {
	e := startEngine(t)
	e.do(t, "PUT", "/api/v1/runs/run-1", map[string]any{"sources": []string{"kafka"}})

	cl, err := kgo.NewClient(kgo.SeedBrokers(e.endpoints["kafka"]))
	if err != nil {
		t.Fatal(err)
	}
	defer cl.Close()

	ctx, cancel := timeoutCtx(5 * time.Second)
	defer cancel()
	res := cl.ProduceSync(ctx, &kgo.Record{
		Topic: "servienta",
		Key:   []byte("k1"),
		Value: []byte("flow-event"),
	})
	if err := res.FirstErr(); err != nil {
		t.Fatalf("produce: %v", err)
	}

	msgs := e.receivedN(t, "kafka", "run-1", 1)
	c := msgs[0]["content"].(map[string]any)
	if c["topic"] != "servienta" || c["value"] != "flow-event" {
		t.Fatalf("kafka message not recorded correctly: %v", c)
	}
}
