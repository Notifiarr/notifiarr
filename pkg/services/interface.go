package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

const valuePrefix = "serviceCheck-"

var ErrSvcsStopped = errors.New("service check routine stopped")

// RunChecks runs checks from an external package.
func (s *Services) RunChecks(input *common.ActionInput) {
	s.stopLock.Lock()
	triggerChan := s.triggerChan
	stopped := s.stopped
	halted := s.actionChan == nil || s.stopping
	s.stopLock.Unlock()

	if halted || !sendLive(triggerChan, input, stopped) {
		mnd.Log.Errorf(input.ReqID, "Cannot run service checks. Go routine is not running.")
	}
}

// RunCheck runs a single check from an external package.
func (s *Services) RunCheck(ctx context.Context, source website.EventType, name string) error {
	s.stopLock.Lock()

	checkChan := s.checkChan
	stopped := s.stopped
	svc, found := s.services[name]
	halted := s.actionChan == nil || s.stopping

	s.stopLock.Unlock()

	if checkChan == nil || halted || stopped == nil {
		return fmt.Errorf("cannot check service, %w", ErrSvcsStopped)
	}

	if !found {
		return fmt.Errorf("%w: service '%s' not found", ErrNoName, name)
	}

	event := triggerCheck{ReqID: mnd.GetID(ctx), Source: source, Service: svc}

	select {
	case checkChan <- event:
		return nil
	case <-stopped:
		return fmt.Errorf("cannot check service, %w", ErrSvcsStopped)
	case <-ctx.Done():
		return fmt.Errorf("cannot check service: %w", ctx.Err())
	}
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

	outstanding := 0
	changes := false

	for _, svc := range s.services {
		if svc.Interval.Duration == 0 || (!forceAll && !svc.Due(now)) {
			continue
		}

		outstanding, changes = s.enqueueCheck(svc, outstanding, changes)
	}

	return s.waitChecks(outstanding, changes)
}

// enqueueCheck queues a check without blocking forever if the worker pool is full.
// Completions are drained in the same select so the done channel cannot deadlock.
func (s *Services) enqueueCheck(svc *Service, outstanding int, changes bool) (int, bool) {
	for {
		select {
		case s.checks <- svc:
			return outstanding + 1, changes
		case changed := <-s.done:
			outstanding--
			changes = changes || changed
		}
	}
}

func (s *Services) waitChecks(outstanding int, changes bool) bool {
	for outstanding > 0 {
		changes = <-s.done || changes
		outstanding--
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
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &CheckResult{
		Interval:    s.Interval.Seconds(),
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
func (s *Services) SendResults(results *Results, reqID string) {
	website.SendData(&website.Request{
		ReqID:      reqID,
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
