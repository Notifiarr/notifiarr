package configfile

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
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

// MakeIPs turns a list of CIDR strings (or plain IPs) into a list of net.IPNet.
// This "allowed" list is later used to check incoming IPs from web requests.
func MakeIPs(upstreams []string) (a AllowedIPs) {
	for _, ipAddr := range upstreams {
		if !strings.Contains(ipAddr, "/") {
			if strings.Contains(ipAddr, ":") {
				ipAddr += "/128"
			} else {
				ipAddr += "/32"
			}
		}

		if _, i, err := net.ParseCIDR(ipAddr); err == nil {
			a = append(a, i)
		}
	}

	return a
}

// CheckPort attempts to bind to a port to check if it's in use or not.
// We use this to check the port before starting the webserver.
func CheckPort(addr string) (string, error) {
	// Cleanup user input.
	addr = strings.TrimPrefix(strings.TrimPrefix(strings.TrimRight(addr, "/"), "http://"), "https://")
	if addr == "" {
		addr = mnd.DefaultBindAddr
	} else if !strings.Contains(addr, ":") {
		addr = "0.0.0.0:" + addr
	}

	a, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return addr, fmt.Errorf("provided ip:port combo is invalid: %w", err)
	}

	l, err := net.ListenTCP("tcp", a)
	if err != nil {
		return addr, fmt.Errorf("unable to listen on provided ip:port: %w", err)
	}
	defer l.Close()

	return addr, nil
}

// CopyFile can be used to make a config file backup.
func CopyFile(src, dst string) error {
	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("%w: cannot overwrite file: %s", os.ErrExist, dst)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening existing file: %w", err)
	}
	defer srcFile.Close()

	srcStat, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("stating existing file: %w", err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_RDWR, srcStat.Mode())
	if err != nil {
		return fmt.Errorf("creating new file: %w", err)
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copying data: %w", err)
	}

	return nil
}
