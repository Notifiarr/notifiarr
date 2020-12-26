package dnclient

import (
	"io/ioutil"
	"os/exec"
)

func openLog(logFile string) error {
	return openCmd("open", "-b", "com.apple.Console", logFile)
}

func openURL(uri string) error {
	return openCmd("open", uri)
}

func openCmd(c string, v ...string) error {
	cmd := exec.Command(c, v...)
	cmd.Stdout = ioutil.Discard
	cmd.Stderr = ioutil.Discard

	return cmd.Run()
}
