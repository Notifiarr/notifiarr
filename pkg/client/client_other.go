//go:build !windows && !darwin && !linux

package client

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Notifiarr/notifiarr/pkg/website"
)

// If you need more fake methods, add them.
//
//nolint:gochecknoglobals
var menu = make(map[string]*fakeMenu)

type fakeMenu struct{}

func (f *fakeMenu) Uncheck()               {}
func (f *fakeMenu) Check()                 {}
func (f *fakeMenu) SetTooltip(interface{}) {}

func (c *Client) printUpdateMessage()        {}
func (c *Client) setupMenus(interface{})     {}
func (c *Client) startTray(_, _ interface{}) {}
func (c *Client) handleAptHook(_ context.Context) error {
	return fmt.Errorf("this feature is not supported on this platform") //nolint:goerr113
}

// AutoWatchUpdate is not used on this OS.
func (c *Client) AutoWatchUpdate(_ interface{}) {}

func (c *Client) checkReloadSignal(ctx context.Context, sigc os.Signal) error {
	return c.reloadConfiguration(ctx, website.EventSignal, "Caught Signal: "+sigc.String())
}

func (c *Client) setSignals() {
	signal.Notify(c.sigkil, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(c.sighup, syscall.SIGHUP, syscall.SIGUSR1)
}
