// Package sender defines the contract for active senders (R13): the mirror of a
// receiver. A sender forms a protocol message and sends it to a target the test
// names on each request (D18) — the engine never stores or discovers the
// application's address.
package sender

import "context"

// Sender sends one protocol's message to a caller-supplied target. For
// request-response protocols it returns the application's reply; for one-way
// protocols it returns a small confirmation. Name() is the stand id, so
// sending a protocol requires that stand to be licensed (D15).
type Sender interface {
	Name() string
	Send(ctx context.Context, target string, payload map[string]any) (map[string]any, error)
}
