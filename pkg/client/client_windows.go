package client

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Notifiarr/notifiarr/pkg/website"
)

func (c *Client) handleAptHook(_ any) error {
	return ErrUnsupport
}

func (c *Client) checkReloadSignal(ctx context.Context, sigc os.Signal) error {
	return c.reloadConfiguration(ctx, website.EventSignal, "Caught Signal: "+sigc.String())
}

func (c *Client) setSignals() {
	signal.Notify(c.sigkil, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(c.sighup, syscall.SIGHUP)
}
