package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/notifiarr"
)

const valuePrefix = "serviceCheck-"

var ErrSvcsStopped = fmt.Errorf("service check routine stopped")

// RunChecks runs checks from an external package.
func (c *Config) RunChecks(source notifiarr.EventType) {
	c.stopLock.Lock()
	defer c.stopLock.Unlock()

	if c.triggerChan == nil || c.stopChan == nil {
		c.Errorf("Cannot run service checks. Go routine is not running.")
		return
	}

	c.triggerChan <- source
}

// RunCheck runs a single check from an external package.
func (c *Config) RunCheck(source notifiarr.EventType, name string) error {
	c.stopLock.Lock()
	defer c.stopLock.Unlock()

	if c.triggerChan == nil || c.stopChan == nil {
		return fmt.Errorf("cannot check service, %w", ErrSvcsStopped)
	}

	svc, ok := c.services[name]
	if !ok {
		return fmt.Errorf("%w: service '%s' not found", ErrNoName, name)
	}

	c.checkChan <- triggerCheck{Source: source, Service: svc}

	return nil
}

// runCheck runs a service check if it is due. Passing force runs it regardless.
func (c *Config) runCheck(svc *Service, force bool) {
	if force || svc.svc.LastCheck.Add(svc.Interval.Duration).Before(time.Now()) {
		c.checks <- svc
		<-c.done
	}
}

// runChecks runs checks that are due. Passing true, runs them even if they're not due.
func (c *Config) runChecks(forceAll bool) {
	if c.checks == nil || c.done == nil {
		return
	}

	count := 0

	for s := range c.services {
		if forceAll || c.services[s].svc.LastCheck.Add(c.services[s].Interval.Duration).Before(time.Now()) {
			count++
			c.checks <- c.services[s]
		}
	}

	somethingChanged := false
	for ; count > 0; count-- {
		somethingChanged = <-c.done || somethingChanged
	}

	if somethingChanged {
		c.updateStatesOnSite()
	}
}

func (c *Config) updateStatesOnSite() {
	values := make(map[string][]byte)

	for _, svc := range c.services {
		values[valuePrefix+svc.Name], _ = json.Marshal(svc.svc)
	}

	if len(values) == 0 {
		return
	}

	if err := c.Notifiarr.SetValues(values); err != nil {
		c.Errorf("Setting Service States on Notifiarr.com: %v", err)
	}
}

// GetResults creates a copy of all the results and returns them.
func (c *Config) GetResults() []*CheckResult {
	svcs := make([]*CheckResult, len(c.services))
	count := 0

	for _, svc := range c.services {
		svcs[count] = &CheckResult{
			Interval:    svc.Interval.Duration.Seconds(),
			Name:        svc.Name,
			State:       svc.svc.State,
			Output:      svc.svc.Output,
			Type:        svc.Type,
			Time:        svc.svc.LastCheck,
			Since:       svc.svc.Since,
			Check:       svc.Value,
			Expect:      svc.Expect,
			IntervalDur: svc.Interval.Duration,
		}
		count++
	}

	return svcs
}

// SendResults sends a set of Results to Notifiarr.
func (c *Config) SendResults(results *Results) {
	results.Interval = c.Interval.Seconds()

	resp, err := c.Notifiarr.SendData(notifiarr.SvcRoute.Path(results.What), results, true)
	if err != nil {
		c.Errorf("Sending %d service updates to Notifiarr, event: %s, buffer: %d/%d, error: %v",
			len(results.Svcs), results.What, len(c.checks), cap(c.checks), err)
		return
	}

	c.Printf("Sent %d service check states to Notifiarr, event: %s, buffer: %d/%d. "+
		"Website took %s and replied with: %s, %s",
		len(results.Svcs), results.What, len(c.checks), cap(c.checks),
		resp.Details.Elapsed, resp.Result, resp.Details.Response)
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

func (c CheckState) Value() uint {
	return uint(c)
}
