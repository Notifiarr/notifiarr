// +build !windows,!darwin

package dnclient

// Run starts the web server and runs Exit to wait for an interrupt signal.
func (c *Client) Run() error {
	c.StartWebServer()

	return c.Exit()
}
