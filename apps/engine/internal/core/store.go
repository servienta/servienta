// Package core holds the engine's recorded state: messages, run declarations,
// and source claims. Mutation is instance-wide, reading is run-scoped (D3);
// attribution is by claimed source only, never guessed (D4).
package core

import (
	"errors"
	"slices"
	"sync"
	"time"
)

var (
	ErrUnknownService = errors.New("unknown receiver")    // N5: distinguishable from "no messages"
	ErrUnknownRun     = errors.New("run is not declared") // R4: not an empty list
	ErrClaimConflict  = errors.New("source already claimed by another run")
	ErrRunExists      = errors.New("run already declared")
)

// Message is one recorded protocol message (R4).
type Message struct {
	TS      time.Time      `json:"ts"`
	Source  string         `json:"source"`        // source IP as observed
	Run     string         `json:"run,omitempty"` // "" = unattributed (D4)
	Content map[string]any `json:"content"`       // parsed, per receiver
}

type Store struct {
	mu       sync.Mutex
	services map[string][]Message // registered receiver -> arrival-ordered messages
	runs     map[string][]string  // run id -> claimed sources
	claims   map[string]string    // source IP -> run id
	Faults   *Faults              // instance-wide fault state (R2, R9)
}

func NewStore() *Store {
	return &Store{
		services: map[string][]Message{},
		runs:     map[string][]string{},
		claims:   map[string]string{},
		Faults:   NewFaults(),
	}
}

// RegisterService makes a receiver known (N5: existing-but-empty is an empty
// list; unregistered is ErrUnknownService).
func (s *Store) RegisterService(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.services[name]; !ok {
		s.services[name] = []Message{}
	}
}

// Record appends a message, attributing it to the run that claimed the source.
func (s *Store) Record(service, sourceIP string, content map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	msgs, ok := s.services[service]
	if !ok {
		return ErrUnknownService
	}
	s.services[service] = append(msgs, Message{
		TS:      time.Now().UTC(),
		Source:  sourceIP,
		Run:     s.claims[sourceIP],
		Content: content,
	})
	return nil
}

// Received returns messages in arrival order. run == "" returns everything,
// attributed or not; a non-empty undeclared run is an explicit error (R4).
func (s *Store) Received(service, run string) ([]Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	msgs, ok := s.services[service]
	if !ok {
		return nil, ErrUnknownService
	}
	if run == "" {
		return slices.Clone(msgs), nil
	}
	if _, ok := s.runs[run]; !ok {
		return nil, ErrUnknownRun
	}
	out := []Message{}
	for _, m := range msgs {
		if m.Run == run {
			out = append(out, m)
		}
	}
	return out, nil
}

// DeclareRun claims sources for a run. One source, one active run (D4).
func (s *Store) DeclareRun(id string, sources []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.runs[id]; ok {
		return ErrRunExists
	}
	for _, src := range sources {
		if owner, ok := s.claims[src]; ok && owner != id {
			return ErrClaimConflict
		}
	}
	s.runs[id] = slices.Clone(sources)
	for _, src := range sources {
		s.claims[src] = id
	}
	return nil
}

// ReleaseRun drops a declaration and its claims; recorded messages stay.
func (s *Store) ReleaseRun(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sources, ok := s.runs[id]
	if !ok {
		return ErrUnknownRun
	}
	for _, src := range sources {
		delete(s.claims, src)
	}
	delete(s.runs, id)
	return nil
}

// Reset is R5: clears every receiver's messages and releases every run
// declaration. Instance-wide by design (D3). Registered services survive —
// reset returns to "running and empty", never to "not started" (N5).
func (s *Store) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for name := range s.services {
		s.services[name] = []Message{}
	}
	s.runs = map[string][]string{}
	s.claims = map[string]string{}
	s.Faults.Reset() // R5 must lift R2 and R9 faults too
}

// Mode exposes a receiver's R9 failure mode (implements receiver.Recorder).
func (s *Store) Mode(service string) (string, int) {
	m, d := s.Faults.ReceiverMode(service)
	return string(m), d
}
