// +build !windows,!darwin

package dnclient

func (c *Client) startTray() error {
	return c.Exit()
}
