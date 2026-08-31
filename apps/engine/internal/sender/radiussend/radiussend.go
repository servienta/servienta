// Package radiussend sends a RADIUS Access-Request to a receiving application's
// server (R13) and returns the response code.
package radiussend

import (
	"context"
	"time"

	"github.com/servienta/servienta/apps/engine/internal/sender"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

type Sender struct{}

func (Sender) Name() string { return "radius" }

func (Sender) Send(ctx context.Context, target string, payload map[string]any) (map[string]any, error) {
	secret := []byte(sender.StrOr(payload, "secret", "throwaway-radius"))
	p := radius.New(radius.CodeAccessRequest, secret)
	rfc2865.UserName_SetString(p, sender.StrOr(payload, "username", "user"))
	rfc2865.UserPassword_SetString(p, sender.StrOr(payload, "password", "password"))
	cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	resp, err := radius.Exchange(cctx, p, target)
	if err != nil {
		return nil, err
	}
	return map[string]any{"sent": true, "code": resp.Code.String()}, nil
}

var _ = time.Second
