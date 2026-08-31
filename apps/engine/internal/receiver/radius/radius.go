// Package radius answers RADIUS Access-Request (R3.3) and lets a test control
// the outcome (R8): Access-Accept, or Access-Reject with a stated reason. Each
// request is recorded (R4). Default: Access-Accept.
package radius

import (
	"context"
	"net"

	"github.com/servienta/servienta/apps/engine/internal/receiver"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

const service = "radius"

type Receiver struct{ Secret string }

func (Receiver) Name() string        { return service }
func (Receiver) Endpoints() []string { return []string{"radius"} }

func (rv Receiver) Start(ctx context.Context, addrs map[string]string, rec receiver.Recorder) (map[string]net.Addr, error) {
	pc, err := net.ListenPacket("udp", addrs["radius"])
	if err != nil {
		return nil, err
	}
	secret := []byte(rv.Secret)
	srv := radius.PacketServer{
		SecretSource: radius.StaticSecretSource(secret),
		Handler: radius.HandlerFunc(func(w radius.ResponseWriter, r *radius.Request) {
			if mode, _ := rec.Mode(service); mode == "drop" || mode == "refuse" {
				return
			}
			user := rfc2865.UserName_GetString(r.Packet)
			_ = rec.Record(service, host(r.RemoteAddr.String()), map[string]any{"username": user})

			code := radius.CodeAccessAccept
			spec := rec.Response(service)
			if str(spec, "outcome") == "reject" {
				code = radius.CodeAccessReject
			}
			resp := r.Response(code)
			if reason := str(spec, "reason"); reason != "" && code == radius.CodeAccessReject {
				rfc2865.ReplyMessage_SetString(resp, reason)
			}
			w.Write(resp)
		}),
	}
	go func() { <-ctx.Done(); srv.Shutdown(context.Background()) }()
	go srv.Serve(pc)
	return map[string]net.Addr{"radius": pc.LocalAddr()}, nil
}

func host(addr string) string {
	if h, _, err := net.SplitHostPort(addr); err == nil {
		return h
	}
	return addr
}

func str(m map[string]any, k string) string {
	if m == nil {
		return ""
	}
	s, _ := m[k].(string)
	return s
}
