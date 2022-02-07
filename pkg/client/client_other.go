//go:build !windows && !darwin

package client

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Notifiarr/notifiarr/pkg/notifiarr"
)

// handleAptHook takes a payload as stdin from dpkg and relays it to notifiarr.com.
// only useful as an apt integration on Debian-based operating systems.
// NEVER return an error, we don't want to hang up apt.
func (c *Client) handleAptHook() error {
	if !c.Config.EnableApt {
		return nil // apt integration is not enabled, bail.
	}

	var (
		grab   bool
		output struct {
			Data    []string `json:"data"`
			CLI     string   `json:"cli"`
			Install int      `json:"install"`
			Remove  int      `json:"remove"`
		}
	)

	for scanner := bufio.NewScanner(os.Stdin); scanner.Scan(); {
		switch line := scanner.Text(); {
		case strings.HasPrefix(line, "CommandLine"):
			output.CLI = line
		case line == "":
			grab = true // grab everything after the empty line.
		case grab:
			output.Data = append(output.Data, line)

			if strings.HasSuffix(line, ".deb") {
				output.Install++
			} else if strings.HasSuffix(line, "**REMOVE**") {
				output.Remove++
			}

			fallthrough
		default: /* debug /**/
			// fmt.Println("hook line", line)
		}
	}

	resp, err := c.website.SendData(notifiarr.TestRoute.Path("apt"), output, true)
	if err != nil {
		fmt.Printf("ERROR Sending Notification to Notifiarr.com: %v %s\n", err, resp)
	} else {
		fmt.Printf("Sent notification to Notifiarr.com; install: %d, remove: %d. %s\n",
			output.Install, output.Remove, resp)
	}

	return nil
}

// If you need more fake methods, add them.
//nolint:gochecknoglobals
var menu = make(map[string]*fakeMenu)

type fakeMenu struct{}

func (f *fakeMenu) Uncheck()               {}
func (f *fakeMenu) Check()                 {}
func (f *fakeMenu) SetTooltip(interface{}) {}

func (c *Client) printUpdateMessage()     {}
func (c *Client) setupMenus(interface{})  {}
func (c *Client) closeDynamicTimerMenus() {}
func (c *Client) startTray(interface{})   {}

// AutoWatchUpdate is not used on this OS.
func (c *Client) AutoWatchUpdate() {}

func (c *Client) checkReloadSignal(sigc os.Signal) error {
	if sigc == syscall.SIGUSR1 && c.Flags.ConfigFile != "" {
		c.Printf("Writing Config File! Caught Signal: %v", sigc)

		if _, err := c.Config.Write(c.Flags.ConfigFile); err != nil {
			c.Errorf("Writing Config File: %v", err)
		}
	} else {
		return c.reloadConfiguration("Caught Signal: " + sigc.String())
	}

	return nil
}

func (c *Client) setSignals() {
	signal.Notify(c.sigkil, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(c.sighup, syscall.SIGHUP, syscall.SIGUSR1)
}
