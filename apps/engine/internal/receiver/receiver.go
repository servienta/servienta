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
	Response(service string) map[string]any // R8 control, nil = default
}

// Receiver is one protocol service. A service may expose several protocol
// surfaces (e.g. syslog over UDP, TCP, and RELP): Endpoints names them, and
// Start binds each to the resolved address and reports the actual bound
// address (R7: no fixed host ports). All surfaces record under Name(), so a
// test reads them back together at /received/<name> (R4). Recorded state and
// failure modes live in the core, so reset needs nothing from the receiver.
type Receiver interface {
	Name() string
	Endpoints() []string // labels this receiver binds; single-surface receivers return {Name()}
	Start(ctx context.Context, addrs map[string]string, rec Recorder) (map[string]net.Addr, error)
}
