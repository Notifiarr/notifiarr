package configfile

import (
	"fmt"
	"net"
)

/* This is a helper method to check if an IP is in a list/cidr. */

// AllowedIPs determines who can set x-forwarded-for.
type AllowedIPs []*net.IPNet

var _ = fmt.Stringer(AllowedIPs(nil))

// String turns a list of allowedIPs into a printable masterpiece.
func (n AllowedIPs) String() (s string) {
	if len(n) < 1 {
		return "(none)"
	}

	for i := range n {
		if s != "" {
			s += ", "
		}

		s += n[i].String()
	}

	return s
}

// Contains returns true if an IP is allowed.
func (n AllowedIPs) Contains(ip string) bool {
	for i := range n {
		if n[i].Contains(net.ParseIP(ip)) {
			return true
		}
	}

	return false
}
