package client

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

/*
func (c *Client) handleAptHook() error {
	return fmt.Errorf("this feature is not supported on this platform") //nolint:goerr113
}
*/

func (c *Client) handleAptHook() error {
	return fmt.Errorf("this feature is not supported on this platform") //nolint:goerr113
}

func (c *Client) printUpdateMessage() {}

func (c *Client) upgradeWindows(_ interface{}) {}

func (c *Client) AutoWatchUpdate() {}

func (c *Client) checkReloadSignal(sigc os.Signal) error {
	if sigc == syscall.SIGUSR1 && c.Flags.ConfigFile != "" {
		c.Printf("Writing Config File! Caught Signal: %v", sigc)

		if _, err := c.Config.Write(c.Flags.ConfigFile); err != nil {
			c.Errorf("Writing Config File: %v", err)
		}
	} else {
		return c.reloadConfiguration("Caught Signal: " + sigc.String())
	}

	return nil
}

func (c *Client) setSignals() {
	signal.Notify(c.sigkil, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(c.sighup, syscall.SIGHUP, syscall.SIGUSR1)
}
