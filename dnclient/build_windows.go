// +build windows

package dnclient

const systrayIcon = "files/favicon.ico"

func (c *Client) startTray() {
	c.startReallyTray()
}
