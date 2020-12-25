// +build windows

package dnclient

const (
	// DefaultConfFile is where the app looks for a config file if one is not provided.
	DefaultConfFile = `dnclient.conf`
)

const systrayIcon = "files/favicon.ico"

func (c *Client) startTray() {
	c.startReallyTray()
}
