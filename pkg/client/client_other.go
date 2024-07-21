//go:build !windows && !darwin && !linux

package client

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Notifiarr/notifiarr/pkg/website"
)

// If you need more fake methods, add them.

//nolint:gochecknoglobals
var menu = make(map[string]*fakeMenu)

type fakeMenu struct{}

func (f *fakeMenu) Uncheck()       {}
func (f *fakeMenu) Check()         {}
func (f *fakeMenu) SetTooltip(any) {}

func (c *Client) setupMenus(any)     {}
func (c *Client) startTray(_, _ any) {}
func (c *Client) handleAptHook(_ context.Context) error {
	return ErrUnsupport
}

func (c *Client) checkReloadSignal(ctx context.Context, sigc os.Signal) error {
	return c.reloadConfiguration(ctx, website.EventSignal, "Caught Signal: "+sigc.String())
}

func (c *Client) setSignals() {
	signal.Notify(c.sigkil, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(c.sighup, syscall.SIGHUP, syscall.SIGUSR1)
}
