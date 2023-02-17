package client

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Notifiarr/notifiarr/pkg/website"
)

/*
func (c *Client) handleAptHook() error {
	return fmt.Errorf("this feature is not supported on this platform") //nolint:goerr113
}
*/

func (c *Client) handleAptHook(_ context.Context) error {
	return fmt.Errorf("this feature is not supported on this platform") //nolint:goerr113
}

func (c *Client) printUpdateMessage() {}

func (c *Client) upgradeWindows(_, _ interface{}) {}

func (c *Client) AutoWatchUpdate(_ interface{}) {}

func (c *Client) checkReloadSignal(ctx context.Context, sigc os.Signal) error {
	return c.reloadConfiguration(ctx, website.EventSignal, "Caught Signal: "+sigc.String())
}

func (c *Client) setSignals() {
	signal.Notify(c.sigkil, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(c.sighup, syscall.SIGHUP, syscall.SIGUSR1)
}
