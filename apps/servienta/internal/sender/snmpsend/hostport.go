package snmpsend

import (
	"net"
	"strconv"
)

func splitHostPort(target string, def uint16) (string, uint16) {
	h, p, err := net.SplitHostPort(target)
	if err != nil {
		return target, def
	}
	n, err := strconv.Atoi(p)
	if err != nil {
		return h, def
	}
	return h, uint16(n)
}
