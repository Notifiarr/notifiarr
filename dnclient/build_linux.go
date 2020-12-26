// +build linux

package dnclient

func hasGUI() bool {
	return false
}

func (c *Client) startTray() {}
