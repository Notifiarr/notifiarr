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

func (c *Config) Start() {
	if c.Timeout.Duration < minimumInterval {
		c.Timeout.Duration = minimumTimeout
	}

	if c.Interval.Duration == 0 || c.stopChan != nil {
		return
	}

	t := time.NewTicker(c.Interval.Duration)
	c.stopChan = make(chan struct{})

	go func() {
		for {
			select {
			case <-t.C:
				c.sendSnapshot()
			case <-c.stopChan:
				t.Stop()
				return
			}
		}
	}()

	c.Printf("==> System Snapshot Collection Started, interval: %v", c.Interval)
}

func (c *Config) Stop() {
	if c == nil || c.stopChan == nil {
		return
	}

	c.stopChan <- struct{}{}
	close(c.stopChan)
	c.stopChan = nil
}

func (c *Config) sendSnapshot() {
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

	body, err := SendJSON(NotifiarrTestURL, b)
	if err != nil {
		c.Errorf("Sending snapshot to Notifiarr: %v: %v", err, string(body))
	} else {
		c.Printf("Systems Snapshot sent to Notifiarr, sending again in %s", c.Interval)
	}
}

func SendJSON(url string, data []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("creating http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("making http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading http response: %w", err)
	}

	return body, nil
}
