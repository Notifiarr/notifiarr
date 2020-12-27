// +build windows darwin

package ui

import "github.com/getlantern/systray"

type menuItem struct {
	*systray.MenuItem
}

func WrapMenu(m *systray.MenuItem) MenuItem {
	return MenuItem(&menuItem{MenuItem: m})
}

func (m *menuItem) Clicked() chan struct{} {
	return m.ClickedCh
}

var _ = MenuItem(&menuItem{MenuItem: nil})
