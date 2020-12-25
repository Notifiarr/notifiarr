// +build freebsd netbsd openbsd

package dnclient

const (
	// DefaultConfFile is where the app looks for a config file if one is not provided.
	DefaultConfFile = "/usr/local/etc/discordnotifier-client/dnclient.conf"
)

const systrayIcon = "none"

func (c *Client) startTray() {
}
