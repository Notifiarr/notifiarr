// +build !darwin,!windows

package dnclient

func hasGUI() bool {
	return false
}

func (c *Client) startTray() {}
