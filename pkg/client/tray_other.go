//go:build !windows && !darwin && !linux

package client

// If you need more fake methods, add them.

//nolint:gochecknoglobals
var menu = make(map[string]*fakeMenu)

type fakeMenu struct{}

func (f *fakeMenu) Enable()          {}
func (f *fakeMenu) Uncheck()         {}
func (f *fakeMenu) Check()           {}
func (f *fakeMenu) SetTooltip(any)   {}
func (c *Client) setupMenus(any)     {}
func (c *Client) startTray(_, _ any) {}
