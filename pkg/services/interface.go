package services

import (
	"fmt"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/notifiarr"
)

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

// runChecks runs checks from an external package.
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

// runChecks runs checks that are due. Passing true, runs them even if they're not due.
func (c *Config) runCheck(svc *Service, force bool) {
	if force || svc.lastCheck.Add(svc.Interval.Duration).Before(time.Now()) {
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
		if forceAll || c.services[s].lastCheck.Add(c.services[s].Interval.Duration).Before(time.Now()) {
			count++
			c.checks <- c.services[s]
		}
	}

	for ; count > 0; count-- {
		<-c.done
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
			State:       svc.state,
			Output:      svc.output,
			Type:        svc.Type,
			Time:        svc.lastCheck,
			Since:       svc.since,
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
