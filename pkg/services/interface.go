package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/website"
)

const valuePrefix = "serviceCheck-"

var ErrSvcsStopped = fmt.Errorf("service check routine stopped")

// RunChecks runs checks from an external package.
func (c *Config) RunChecks(source website.EventType) {
	c.stopLock.Lock()
	defer c.stopLock.Unlock()

	if c.triggerChan == nil || c.stopChan == nil {
		c.Errorf("Cannot run service checks. Go routine is not running.")
		return
	}

	c.triggerChan <- source
}

// RunCheck runs a single check from an external package.
func (c *Config) RunCheck(source website.EventType, name string) error {
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
func (c *Config) runCheck(svc *Service, force bool) bool {
	if force || svc.Due() {
		c.checks <- svc
		return <-c.done
	}

	return false
}

// runChecks runs checks that are due. Passing true, runs them even if they're not due.
func (c *Config) runChecks(forceAll bool) {
	if c.checks == nil || c.done == nil {
		return
	}

	count := 0

	for s := range c.services {
		func() {
			c.services[s].svc.RLock()
			defer c.services[s].svc.RUnlock()

			if forceAll || c.services[s].svc.LastCheck.Add(c.services[s].Interval.Duration).Before(time.Now()) {
				count++
				c.checks <- c.services[s]
			}
		}()
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
	var (
		err    error
		values = make(map[string][]byte)
	)

	for _, svc := range c.services {
		func() {
			svc.svc.RLock()
			defer svc.svc.RUnlock()

			if values[valuePrefix+svc.Name], err = json.Marshal(&(svc.svc)); err != nil {
				c.Errorf("Marshaling Service %s: %v", svc.Name, err)
			}
		}()
	}

	if len(values) == 0 {
		return
	}

	if err := c.Website.SetValues(values); err != nil {
		c.Errorf("Setting Service States on Notifiarr.com: %v", err)
	}
}

// GetResults creates a copy of all the results and returns them.
func (c *Config) GetResults() []*CheckResult {
	svcs := make([]*CheckResult, len(c.services))
	count := 0

	for _, svc := range c.services {
		svcs[count] = svc.copyResults()
		count++
	}

	return svcs
}

func (s *Service) copyResults() *CheckResult {
	s.svc.RLock()
	defer s.svc.RUnlock()

	return &CheckResult{
		Interval:    s.Interval.Duration.Seconds(),
		Name:        s.Name,
		State:       s.svc.State,
		Output:      s.svc.Output,
		Type:        s.Type,
		Time:        s.svc.LastCheck,
		Since:       s.svc.Since,
		Check:       s.Value,
		Expect:      s.Expect,
		IntervalDur: s.Interval.Duration,
	}
}

// SendResults sends a set of Results to Notifiarr.
func (c *Config) SendResults(results *Results) {
	results.Interval = c.Interval.Seconds()

	c.Website.SendData(&website.Request{
		Route:      website.SvcRoute,
		Event:      results.What,
		LogPayload: true,
		LogMsg: fmt.Sprintf("%d service updates to Notifiarr, event: %s, buffer: %d/%d",
			len(results.Svcs), results.What, len(c.checks), cap(c.checks)),
		Payload: results,
	})
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
