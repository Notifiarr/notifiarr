package snapshot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func (c *Config) Start(apikey string) {
	if c.Timeout.Duration < minimumTimeout {
		c.Timeout.Duration = minimumTimeout
	}

	if c.Interval.Duration == 0 || c.stopChan != nil {
		return
	}

	t := time.NewTicker(c.Interval.Duration)
	c.stopChan = make(chan struct{})

	go func() {
		defer func() {
			t.Stop()
			close(c.stopChan)
			c.stopChan = nil
		}()

		for {
			select {
			case <-t.C:
				c.sendSnapshot(apikey)
			case <-c.stopChan:
				return
			}
		}
	}()
	c.logStart()
}

func (c *Config) logStart() {
	var ex string

	for k, v := range map[string]bool{
		"raid":    c.Raid,
		"disks":   c.DiskUsage,
		"drives":  c.DriveData,
		"uptime":  c.Uptime,
		"cpumem":  c.CPUMem,
		"cputemp": c.CPUTemp,
		"zfs":     c.ZFSPools != nil,
		"sudo":    c.UseSudo && c.DriveData,
	} {
		if !v {
			continue
		}

		if ex != "" {
			ex += ", "
		}

		ex += k
	}

	c.Printf("==> System Snapshot Collection Started, interval: %v, timeout: %v, enabled: %s", c.Interval, c.Timeout, ex)
}

func (c *Config) Stop() {
	if c != nil && c.stopChan != nil {
		c.stopChan <- struct{}{}
	}
}

func (c *Config) sendSnapshot(apikey string) {
	snapshot, errs := c.GetSnapshot()
	if len(errs) > 0 {
		for _, err := range errs {
			if err != nil {
				c.Errorf("Snapshot: %v", err)
			}
		}
	}

	b, _ := json.Marshal(&snapshot)
	// log.Println(string(b))

	body, err := SendJSON(NotifiarrTestURL, apikey, b)
	if err != nil {
		c.Errorf("Sending snapshot to Notifiarr: %v: %v", err, string(body))
	} else {
		c.Printf("Systems Snapshot sent to Notifiarr, sending again in %s", c.Interval)
	}
}

// SendJSON posts a JSON payload to a URL. Returns the response body or an error.
// The response status code is lost.
func SendJSON(url, apikey string, data []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("creating http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if apikey != "" {
		req.Header.Set("X-API-Key", apikey)
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("making http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return body, fmt.Errorf("reading http response: %w", err)
	}

	return body, nil
}
