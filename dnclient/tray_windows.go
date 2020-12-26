package dnclient

import (
	"io/ioutil"
	"os/exec"
	"syscall"
)

const systrayIcon = "files/macos.png"

func hasGUI() bool {
	return true
}

func openLog(logFile string) error {
	return openCmd("cmd", "/c", "start", "PowerShell", "Get-Content", "-Tail", "1000", "-Wait", "-Encoding", "utf8", "-Path", logFile)
}

func openURL(uri string) error {
	return openCmd("cmd", "/c", "start", uri)
}

func openFile(filePath string) error {
	return openURL("file://" + filePath)
}

func openCmd(c string, v ...string) error {
	cmd := exec.Command(c, v...)
	cmd.Stdout = ioutil.Discard
	cmd.Stderr = ioutil.Discard
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	return cmd.Run()
}
