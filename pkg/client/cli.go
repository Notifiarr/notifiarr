package client

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/bindata"
	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/services"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/rotatorr"
)

/* This file contains procedures for --write, --curl and --ps CLI flags. */

const curlTimeout = 15 * time.Second

// Errors.
var (
	ErrInvalidHeader = errors.New("invalid header provided; must contain a colon")
)

// forceWriteWithExit is called only when a user passes --write or --reset on the command line.
func (c *Client) forceWriteWithExit(ctx context.Context, fileName string) error {
	if fileName == "example" || fileName == "---" {
		// Bubilding a default template.
		fileName = c.Flags.ConfigFile + ".new"
		c.Config.LogFile = ""
		c.Config.LogConfig.DebugLog = ""
		c.Config.HTTPLog = ""
		c.Config.FileMode = logs.FileMode(rotatorr.FileMode)
		c.Config.Debug = false
		configfile.ForceAllTmpl = true
	}

	f, err := c.Config.Write(ctx, fileName, false)
	if err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	c.Print("Wrote Config File:", f)

	return nil
}

func (c *Client) resetAdminPassword(ctx context.Context) error {
	c.Config.SSLCrtFile = ""
	c.Config.SSLKeyFile = ""

	password := configfile.DefaultUsername + ":" + configfile.GeneratePassword()
	if err := c.Config.UIPassword.Set(password); err != nil {
		return fmt.Errorf("setting password failed: %w", err)
	}

	c.Printf("New '%s' user password: %s", configfile.DefaultUsername, password)
	c.Printf("Writing Config File: %s", c.Flags.ConfigFile)

	return c.saveNewConfig(ctx, c.Config)
}

// printProcessList is triggered by the --ps command line arg.
func printProcessList(ctx context.Context) error {
	ps, err := getProcessList(ctx)
	if err != nil {
		return err
	}

	fmt.Println(ps.String()) //nolint:forbidigo

	return nil
}

func getProcessList(ctx context.Context) (*bytes.Buffer, error) {
	pslist, err := services.GetAllProcesses(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get processes: %w", err)
	}

	var buf bytes.Buffer

	for _, proc := range pslist {
		if mnd.IsFreeBSD {
			fmt.Fprintf(&buf, "[%-5d] %s\n", proc.PID, proc.CmdLine)
			continue
		}

		t := "unknown"
		if !proc.Created.IsZero() {
			t = time.Since(proc.Created).Round(time.Second).String()
		}

		fmt.Fprintf(&buf, "[%-5d] %-11s: %s\n", proc.PID, t, proc.CmdLine)
	}

	return &buf, nil
}

// curlURL is called from the --curl CLI arg.
func curlURL(url string, headers []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), curlTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating http request: %w", err)
	}

	if err := addHeaders(req, headers); err != nil {
		return err
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return fmt.Errorf("making http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading body: %w", err)
	}

	printCurlReply(resp, body)

	return nil
}

func addHeaders(req *http.Request, headers []string) error {
	const headerSize = 2

	for i, h := range headers {
		if !strings.Contains(h, ":") {
			return fmt.Errorf("%w: %d: %s", ErrInvalidHeader, i+1, h)
		}

		header := strings.SplitN(h, ":", headerSize)
		req.Header.Add(strings.TrimSpace(header[0]), strings.TrimSpace(header[1]))
	}

	return nil
}

//nolint:forbidigo
func printCurlReply(resp *http.Response, body []byte) {
	fmt.Println(resp.Status)

	for header, value := range resp.Header {
		for _, v := range value {
			fmt.Println(header + ": " + v)
		}
	}

	fmt.Println("")
	fmt.Println(string(body))
}

// Fortune returns a fortune.
func Fortune() string {
	fortunes := strings.Split(bindata.Fortunes, "\n%\n")
	return fortunes[rand.Intn(len(fortunes))] //nolint:gosec
}

// handleAptHook takes a payload as stdin from dpkg and relays it to notifiarr.com.
// only useful as an apt integration on Debian-based operating systems.
// NEVER return an error, we don't want to hang up apt.
func (c *Client) handleAptHook(ctx context.Context) error { //nolint:cyclop
	if !mnd.IsLinux {
		return ErrUnsupport
	} else if !c.Config.EnableApt {
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
		} //nolint:wsl
	}

	resp, _, err := c.Config.RawGetData(ctx, &website.Request{
		Route:   website.PkgRoute,
		Event:   "apt",
		Payload: output,
	})
	//nolint:forbidigo
	if err != nil {
		fmt.Printf("ERROR Sending Notification to notifiarr.com: %v%s\n", err, resp)
	} else {
		fmt.Printf("Sent notification to notifiarr.com; install: %d, remove: %d%s\n",
			output.Install, output.Remove, resp)
	}

	return nil
}
