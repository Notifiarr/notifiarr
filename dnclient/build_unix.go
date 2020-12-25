// +build !windows,!darwin

package dnclient

const systrayIcon = "none"

func (c *Client) startTray() {}
