// Package ntp answers NTP client requests (R3.6) and lets a test control the
// reply (R8): a time offset, an altered stratum, or a refusal to serve. Each
// request is recorded (R4). Default: correct time, stratum 1. RFC 5905 server
// subset — a small, well-specified wire format implemented in-repo (D6).
package ntp

import (
	"context"
	"encoding/binary"
	"net"
	"time"

	"github.com/servienta/servienta/apps/engine/internal/receiver"
)

const service = "ntp"

// ntpEpochOffset is seconds between the NTP epoch (1900) and the Unix epoch (1970).
const ntpEpochOffset = 2208988800

type Receiver struct{}

func (Receiver) Name() string        { return service }
func (Receiver) Endpoints() []string { return []string{"ntp"} }

func (Receiver) Start(ctx context.Context, addrs map[string]string, rec receiver.Recorder) (map[string]net.Addr, error) {
	pc, err := net.ListenPacket("udp", addrs["ntp"])
	if err != nil {
		return nil, err
	}
	go func() { <-ctx.Done(); pc.Close() }()
	go serve(pc, rec)
	return map[string]net.Addr{"ntp": pc.LocalAddr()}, nil
}

func serve(pc net.PacketConn, rec receiver.Recorder) {
	buf := make([]byte, 48)
	for {
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			return
		}
		if n < 48 {
			continue
		}
		if mode, _ := rec.Mode(service); mode == "drop" || mode == "refuse" {
			continue
		}
		src := host(addr.String())
		_ = rec.Record(service, src, map[string]any{"request": true})

		spec := rec.Response(service)
		if str(spec, "outcome") == "refuse" {
			continue // NTP has no "refuse" reply; silence is the refusal
		}

		var req [48]byte
		copy(req[:], buf[:48])
		resp := buildResponse(req, spec)
		pc.WriteTo(resp, addr)
	}
}

func buildResponse(req [48]byte, spec map[string]any) []byte {
	var r [48]byte
	// LI=0, VN=4, Mode=4 (server)
	r[0] = 0<<6 | 4<<3 | 4
	r[1] = byte(stratumOr(spec, 1))
	r[2] = req[2] // echo poll
	r[3] = 0xEC   // precision ~ 2^-20

	offset := time.Duration(intv(spec, "offset_ms")) * time.Millisecond
	now := time.Now().Add(offset)
	ts := toNTP(now)

	// Reference, originate (client's transmit), receive, transmit timestamps.
	binary.BigEndian.PutUint64(r[16:], ts) // reference
	copy(r[24:32], req[40:48])             // originate = client's transmit
	binary.BigEndian.PutUint64(r[32:], ts) // receive
	binary.BigEndian.PutUint64(r[40:], ts) // transmit
	return r[:]
}

func toNTP(t time.Time) uint64 {
	secs := uint64(t.Unix() + ntpEpochOffset)
	frac := uint64((t.Nanosecond()) * (1 << 32) / 1e9)
	return secs<<32 | frac
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

func intv(m map[string]any, k string) int {
	if m == nil {
		return 0
	}
	if f, ok := m[k].(float64); ok {
		return int(f)
	}
	return 0
}

func stratumOr(m map[string]any, def int) int {
	if v := intv(m, "stratum"); v != 0 {
		return v
	}
	return def
}
