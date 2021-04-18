package services

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type result struct {
	output string
	state  CheckState
}

func (s *Service) validate() error { //nolint:cyclop
	s.state = StateUnknown

	if s.Name == "" {
		return fmt.Errorf("%s: %w", s.Value, ErrNoName)
	} else if s.Value == "" {
		return fmt.Errorf("%s: %w", s.Name, ErrNoCheck)
	}

	switch s.Type {
	case CheckHTTP:
		if s.Expect == "" {
			s.Expect = "200"
		}
	case CheckTCP:
		if !strings.Contains(s.Value, ":") {
			return ErrBadTCP
		}
	case CheckPROC:
		if err := s.checkProcValues(); err != nil {
			return err
		}
	case CheckPING:
	default:
		return ErrInvalidType
	}

	if s.Timeout.Duration == 0 {
		s.Timeout.Duration = DefaultTimeout
	} else if s.Timeout.Duration < MinimumTimeout {
		s.Timeout.Duration = MinimumTimeout
	}

	if s.Interval.Duration == 0 {
		s.Interval.Duration = DefaultCheckInterval
	} else if s.Interval.Duration < MinimumCheckInterval {
		s.Interval.Duration = MinimumCheckInterval
	}

	return nil
}

func (s *Service) check() bool {
	// check this service.
	switch s.Type {
	case CheckHTTP:
		return s.update(s.checkHTTP())
	case CheckTCP:
		return s.update(s.checkTCP())
	case CheckPING:
		return s.update(s.checkPING())
	case CheckPROC:
		return s.update(s.checkProccess())
	default:
		return false
	}
}

// Return true if the service state changed.
func (s *Service) update(r *result) bool {
	if s.lastCheck = time.Now().Round(time.Microsecond); s.since.IsZero() {
		s.since = s.lastCheck
	}

	s.output = r.output

	if s.state == r.state {
		s.log.Printf("Service Checked: %s, state: %s for %v, output: %s",
			s.Name, s.state, time.Since(s.since).Round(time.Second), s.output)
		return false
	}

	s.log.Printf("Service Checked: %s, state: %s => %s, output: %s", s.Name, s.state, r.state, s.output)
	s.since = s.lastCheck
	s.state = r.state

	return true
}

const maxBody = 150

func (s *Service) checkHTTP() *result {
	r := &result{
		state:  StateUnknown,
		output: "unknown",
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout.Duration)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.Value, nil)
	if err != nil {
		r.output = "creating request: " + err.Error()
		return r
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		r.output = "making request: " + err.Error()
		return r
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		r.output = "reading body: " + err.Error()
		return r
	}

	if strconv.Itoa(resp.StatusCode) == s.Expect {
		r.state = StateOK
		r.output = resp.Status

		return r
	}

	if len(body) > maxBody {
		body = body[:maxBody]
	}

	r.state = StateCritical
	r.output = resp.Status + ": " + strings.TrimSpace(string(body))

	return r
}

func (s *Service) checkTCP() *result {
	r := &result{
		state:  StateUnknown,
		output: "unknown",
	}

	switch conn, err := net.DialTimeout("tcp", s.Value, s.Timeout.Duration); {
	case err != nil:
		r.state = StateCritical
		r.output = "connection error: " + err.Error()
	case conn == nil:
		r.state = StateUnknown
		r.output = "connection failed, no specific error"
	default:
		conn.Close()

		r.state = StateOK
		r.output = "connected to port " + strings.Split(s.Value, ":")[1] + " OK"
	}

	return r
}

func (s *Service) checkPING() *result {
	return &result{
		state:  StateUnknown,
		output: "ping does not work yet",
	}
}
