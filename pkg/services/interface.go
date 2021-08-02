package services

import (
	"encoding/json"
	"time"
)

func (c *Config) RunAllChecksSendResult(source string) {
	if !c.Disabled {
		c.triggerChan <- source
	}
}

// RunChecks runs checks that are due. Passing true, runs them even if they're not due.
// Returns true if a service state changed.
func (c *Config) RunChecks(forceAll bool) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	stateChange := false

	for s := range c.services {
		if forceAll || c.services[s].lastCheck.Add(c.services[s].Interval.Duration).Before(time.Now()) {
			count++
			c.checks <- c.services[s]
		}
	}

	for ; count > 0; count-- {
		if sc := <-c.done; sc {
			stateChange = true
		}
	}

	return stateChange
}

// GetResults creates a copy of all the results and returns them.
func (c *Config) GetResults() []*CheckResult {
	c.mu.Lock()
	defer c.mu.Unlock()

	svcs := make([]*CheckResult, len(c.services))
	count := 0

	for _, s := range c.services {
		svcs[count] = &CheckResult{
			Interval: s.Interval.Duration.Seconds(),
			Name:     s.Name,
			State:    s.state,
			Output:   s.output,
			Type:     s.Type,
			Time:     s.lastCheck,
			Since:    s.since,
		}
		count++
	}

	return svcs
}

// SendResults sends a set of Results to Notifiarr.
func (c *Config) SendResults(url string, results *Results) {
	results.Type = NotifiarrEventType
	results.Interval = c.Interval.Seconds()

	data, _ := json.MarshalIndent(results, "", " ")
	if _, _, err := c.Notifiarr.SendJSON(url, data); err != nil {
		c.Errorf("Sending service check update to %s: %v", url, err)
	} else {
		c.Printf("Sent %d service check states to %s", len(results.Svcs), url)
	}
}

// String turns a check status into a human string.
func (c CheckState) String() string {
	switch c {
	default:
		fallthrough
	case StateUnknown:
		return "Unknown"
	case StateCritical:
		return "Critical"
	case StateWarning:
		return "Warning"
	case StateOK:
		return "OK"
	}
}
