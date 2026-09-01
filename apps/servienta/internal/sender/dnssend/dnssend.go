// Package dnssend queries a receiving application's DNS server (R13) and
// returns the answer.
package dnssend

import (
	"context"
	"time"

	"github.com/miekg/dns"
	"github.com/servienta/servienta/apps/servienta/internal/sender"
)

type Sender struct{}

func (Sender) Name() string { return "dns" }

func (Sender) Send(ctx context.Context, target string, payload map[string]any) (map[string]any, error) {
	qname := dns.Fqdn(sender.StrOr(payload, "qname", "example.test"))
	m := new(dns.Msg)
	m.SetQuestion(qname, dns.TypeA)
	c := &dns.Client{Timeout: 3 * time.Second}
	r, _, err := c.Exchange(m, target)
	if err != nil {
		return nil, err
	}
	answers := []string{}
	for _, a := range r.Answer {
		if rr, ok := a.(*dns.A); ok {
			answers = append(answers, rr.A.String())
		}
	}
	return map[string]any{"sent": true, "rcode": dns.RcodeToString[r.Rcode], "answers": answers}, nil
}
