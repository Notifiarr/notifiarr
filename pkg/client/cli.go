package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/services"
	"golift.io/rotatorr"
)

/* This file contains procedures for --write, --curl and --ps CLI flags. */

const curlTimeout = 15 * time.Second

// forceWriteWithExit is called only when a user passes --write on the command line.
func (c *Client) forceWriteWithExit(f string) error {
	if f == "-" {
		f = c.Flags.ConfigFile
	} else if f == "example" || f == "---" {
		// Bubilding a default template.
		f = c.Flags.ConfigFile
		c.Config.LogFile = ""
		c.Config.LogConfig.DebugLog = ""
		c.Config.HTTPLog = ""
		c.Config.FileMode = logs.FileMode(rotatorr.FileMode)
		c.Config.Debug = false
		c.Config.Snapshot.Interval.Duration = mnd.HalfHour
		configfile.ForceAllTmpl = true
	}

	f, err := c.Config.Write(f)
	if err != nil { // f purposely shadowed.
		return fmt.Errorf("writing config: %s: %w", f, err)
	}

	c.Print("Wrote Config File:", f)

	return nil
}

// printProcessList is triggered by the --ps command line arg.
//nolint:forbidigo
func printProcessList() error {
	pslist, err := services.GetAllProcesses()
	if err != nil {
		return fmt.Errorf("unable to get processes: %w", err)
	}

	for _, p := range pslist {
		if runtime.GOOS == "freebsd" {
			fmt.Printf("[%-5d] %s\n", p.PID, p.CmdLine)
			continue
		}

		t := "unknown"
		if !p.Created.IsZero() {
			t = time.Since(p.Created).Round(time.Second).String()
		}

		fmt.Printf("[%-5d] %-11s: %s\n", p.PID, t, p.CmdLine)
	}

	return nil
}

// curlURL is called from the --curl CLI arg.
func curlURL(url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), curlTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating http request: %w", err)
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return fmt.Errorf("making http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading body: %w", err)
	}

	printCurlReply(resp, body)

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
