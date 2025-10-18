package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

const valuePrefix = "serviceCheck-"

var ErrSvcsStopped = errors.New("service check routine stopped")

// RunChecks runs checks from an external package.
func (s *Services) RunChecks(source website.EventType) {
	s.stopLock.Lock()
	defer s.stopLock.Unlock()

	if s.triggerChan == nil || s.stopChan == nil {
		mnd.Log.Errorf("Cannot run service checks. Go routine is not running.")
		return
	}

	s.triggerChan <- source
}

// RunCheck runs a single check from an external package.
func (s *Services) RunCheck(source website.EventType, name string) error {
	s.stopLock.Lock()
	defer s.stopLock.Unlock()

	if s.triggerChan == nil || s.stopChan == nil {
		return fmt.Errorf("cannot check service, %w", ErrSvcsStopped)
	}

	svc, ok := s.services[name]
	if !ok {
		return fmt.Errorf("%w: service '%s' not found", ErrNoName, name)
	}

	s.checkChan <- triggerCheck{Source: source, Service: svc}

	return nil
}

// runCheck runs a service check if it is due. Passing force runs it regardless.
func (s *Services) runCheck(svc *Service, force bool, now time.Time) bool {
	if force || svc.Due(now) {
		s.checks <- svc
		return <-s.done
	}

	return false
}

// runChecks runs checks that are due. Passing true, runs them even if they're not due.
// Returns true if any service state changed.
func (s *Services) runChecks(forceAll bool, now time.Time) bool {
	if s.checks == nil || s.done == nil {
		return false
	}

	count := 0
	changes := false

	for svc := range s.services {
		if forceAll || s.services[svc].Due(now) {
			count++
			s.checks <- s.services[svc]
		}
	}

	for ; count > 0; count-- {
		changes = <-s.done || changes
	}

	return changes
}

// GetResults creates a copy of all the results and returns them.
func (s *Services) GetResults() []*CheckResult {
	svcs := make([]*CheckResult, len(s.services))
	count := 0

	for _, svc := range s.services {
		svcs[count] = svc.copyResults()
		count++
	}

	return svcs
}

func (s *Service) copyResults() *CheckResult {
	s.RLock()
	defer s.RUnlock()

	return &CheckResult{
		Interval:    s.Interval.Duration.Seconds(),
		Name:        s.Name,
		State:       s.State,
		Output:      s.Output,
		Type:        s.Type,
		Time:        s.LastCheck,
		Since:       s.Since,
		Check:       s.Value,
		Expect:      s.Expect,
		IntervalDur: s.Interval.Duration,
		Metadata:    s.Tags,
	}
}

// SendResults sends a set of Results to Notifiarr.
func (s *Services) SendResults(results *Results) {
	results.Interval = s.Interval.Seconds()

	website.SendData(&website.Request{
		Route:      website.SvcRoute,
		Event:      results.What,
		LogPayload: true,
		LogMsg: fmt.Sprintf("%d service updates to Notifiarr, event: %s, buffer: %d/%d",
			len(results.Svcs), results.What, len(s.checks), cap(s.checks)),
		Payload: results,
	})
}

// String turns a check status into a human string.
func (s CheckState) String() string {
	switch s {
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

func (s CheckState) Value() uint {
	return uint(s)
}
