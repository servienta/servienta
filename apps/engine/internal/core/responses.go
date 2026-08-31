package core

import "sync"

// Responses holds the R8 response controls: for each request-response service
// (RADIUS, TACACS+, DNS, NTP) the test sets, for the run, exactly what the
// engine answers. Instance-wide (D3), cleared by Reset (R5). The default for
// every service is a successful, valid reply — represented by the absence of a
// control here, which each service reads as "answer normally".
type Responses struct {
	mu   sync.RWMutex
	spec map[string]map[string]any // service -> free-form control the service interprets
}

func NewResponses() *Responses {
	return &Responses{spec: map[string]map[string]any{}}
}

// Set replaces a service's response control for the run. A nil or empty spec
// restores the default (successful reply).
func (r *Responses) Set(service string, spec map[string]any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(spec) == 0 {
		delete(r.spec, service)
		return
	}
	r.spec[service] = spec
}

// Get returns the service's control, or nil for the default.
func (r *Responses) Get(service string) map[string]any {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.spec[service]
}

// Reset restores every service to its default reply; part of R5.
func (r *Responses) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.spec = map[string]map[string]any{}
}
