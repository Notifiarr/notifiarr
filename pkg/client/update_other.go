// +build !windows,!darwin

package client

import (
	"os"
	"os/signal"
	"syscall"
)

func (c *Client) printUpdateMessage() {}

func (c *Client) AutoWatchUpdate() {}

func (c *Client) checkReloadSignal(sigc os.Signal) {
	if sigc == syscall.SIGUSR1 && c.Flags.ConfigFile != "" {
		c.Printf("Writing Config File! Caught Signal: %v", sigc)

		if _, err := c.Config.Write(c.Flags.ConfigFile); err != nil {
			c.Errorf("Writing Config File: %v", err)
		}
	} else {
		c.reloadConfiguration("caught signal: " + sigc.String())
	}
}

func (c *Client) setReloadSignals() {
	signal.Notify(c.sighup, syscall.SIGHUP, syscall.SIGUSR1)
}
