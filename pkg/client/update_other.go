// +build !windows,!darwin

package client

func (c *Client) printUpdateMessage() {}

func (c *Client) AutoWatchUpdate() {}
