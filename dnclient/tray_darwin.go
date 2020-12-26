package dnclient

import (
	"io/ioutil"
	"os"
	"os/exec"
)

const systrayIcon = "files/macos.png"

func hasGUI() bool {
	return os.Getenv("USEGUI") == "true"
}

func openLog(logFile string) error {
	return openCmd("open", "-b", "com.apple.Console", logFile)
}

func openURL(uri string) error {
	return openCmd("open", uri)
}

func openFile(filePath string) error {
	return openCmd("open", filePath)
}

func openCmd(c string, v ...string) error {
	cmd := exec.Command(c, v...)
	cmd.Stdout = ioutil.Discard
	cmd.Stderr = ioutil.Discard

	return cmd.Run()
}
