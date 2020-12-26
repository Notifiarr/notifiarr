// +build !windows,!darwin

package dnclient

func hasGUI() bool {
	return false
}

func (c *Client) startTray() {}

func openFile(_ string) error { return nil }

/*
func openURL(uri string) error {
	return nil
}
*/
