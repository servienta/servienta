// Package receiver defines the contract every protocol receiver implements
// (R10): a new protocol is added by implementing Receiver and registering it —
// with no changes to the engine core.
package receiver

import (
	"context"
	"net"
)

// Recorder is the only write path a receiver has into the engine: every
// message it accepts is recorded with its observed source address, and the
// core attributes it to a run by claimed source (D4). It also exposes the
// receiver's current R9 failure mode so the receiver can honor it.
type Recorder interface {
	Record(service, sourceIP string, content map[string]any) error
	Mode(service string) (mode string, delayMs int)
}

// Receiver is one protocol endpoint. Start binds addr, serves until ctx is
// canceled, and reports the actual bound address (R7: no fixed host ports).
// Recorded state lives in the core, so POST /reset needs nothing from the
// receiver (R5); failure modes (R9) extend this interface in phase 1.
type Receiver interface {
	Name() string
	Start(ctx context.Context, addr string, rec Recorder) (net.Addr, error)
}
