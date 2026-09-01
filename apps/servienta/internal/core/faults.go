package core

import "sync"

// FileFaultKind is an R2 fault a file transport can be forced into for a
// named fixture. Instance-wide, cleared by Reset (D3, R5).
type FileFaultKind string

const (
	FileAuthReject FileFaultKind = "auth-reject" // authentication rejected
	FileMissing    FileFaultKind = "missing"     // file reported absent
	FileTruncate   FileFaultKind = "truncate"    // transfer cut off midway
	FileCorrupt    FileFaultKind = "corrupt"     // content corrupted
	// "host unreachable" is R9 on the file transport (refuse), not per-file.
)

// ReceiverMode is an R9 failure mode a receiver can be put into. Instance-wide,
// cleared by Reset. A mode is distinguishable from "not running" (R9, N5).
type ReceiverMode string

const (
	ModeNormal  ReceiverMode = ""       // default: serve normally
	ModeRefuse  ReceiverMode = "refuse" // refuse connections
	ModeDrop    ReceiverMode = "drop"   // accept and silently drop
	ModeDelay   ReceiverMode = "delay"  // respond/accept after a delay
	ModeCut     ReceiverMode = "cut"    // cut the connection mid-message
	ModeProtoEr ReceiverMode = "error"  // return a protocol error where allowed
)

// Faults holds instance-wide fault state. Guarded independently of messages so
// a receiver can read its mode on the hot path without contending on records.
type Faults struct {
	mu         sync.RWMutex
	fileFaults map[string]FileFaultKind // fixture path -> fault
	recvModes  map[string]ReceiverMode  // service -> mode
	recvDelay  map[string]int           // service -> delay ms (for ModeDelay)
}

func NewFaults() *Faults {
	return &Faults{
		fileFaults: map[string]FileFaultKind{},
		recvModes:  map[string]ReceiverMode{},
		recvDelay:  map[string]int{},
	}
}

func (f *Faults) SetFileFault(fixture string, kind FileFaultKind) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fileFaults[fixture] = kind
}

func (f *Faults) FileFault(fixture string) (FileFaultKind, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	k, ok := f.fileFaults[fixture]
	return k, ok
}

func (f *Faults) SetReceiverMode(service string, mode ReceiverMode, delayMs int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if mode == ModeNormal {
		delete(f.recvModes, service)
		delete(f.recvDelay, service)
		return
	}
	f.recvModes[service] = mode
	f.recvDelay[service] = delayMs
}

func (f *Faults) ReceiverMode(service string) (ReceiverMode, int) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.recvModes[service], f.recvDelay[service]
}

// Reset clears every fault; part of R5's instance-wide reset.
func (f *Faults) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fileFaults = map[string]FileFaultKind{}
	f.recvModes = map[string]ReceiverMode{}
	f.recvDelay = map[string]int{}
}
