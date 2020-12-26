// +build !windows,!darwin

package dnclient

func hasGUI() bool {
	return false
}

func (c *Client) startTray() {}

func openURL(uri string) error {
	return nil
}

func openFile(filePath string) error {
	return nil
}
