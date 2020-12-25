// +build !windows,!darwin,!freebsd

package dnclient

const (
	// DefaultConfFile is where the app looks for a config file if one is not provided.
	DefaultConfFile = "/etc/discordnotifier-client/dnclient.conf"
)

const systrayIcon = "none"

func (c *Client) startTray() {}
