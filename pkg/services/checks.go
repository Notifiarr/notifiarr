package services

import (
	"context"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func (s *Service) validate() error {
	if s.Name == "" {
		return ErrNoName
	} else if s.Value == "" {
		return ErrNoCheck
	}

	switch s.Type {
	case CheckHTTP:
		if s.Expect == "" {
			s.Expect = "200"
		}
	case CheckTCP:
	case CheckPING:
	default:
		return ErrInvalidType
	}

	if s.Timeout.Duration == 0 {
		s.Timeout.Duration = DefaultTimeout
	} else if s.Timeout.Duration < MinimumTimeout {
		s.Timeout.Duration = MinimumTimeout
	}

	return nil
}

// startCheckers runs Parallel checkers.
func (c *Config) startCheckers() {
	for i := uint(0); i < c.Parallel; i++ {
		go func() {
			for check := range c.checks {
				check.check()
				c.done <- struct{}{}
			}

			c.done <- struct{}{}
		}()
	}
}

func (s *Service) check() {
	// check this service.
	switch s.Type {
	case CheckHTTP:
		s.checkHTTP()
	case CheckTCP:
		s.checkTCP()
	case CheckPING:
		s.checkPING()
	}
}

const maxBody = 150

func (s *Service) checkHTTP() {
	s.state = StateUnknown
	s.output = "not working yet"
	s.lastCheck = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout.Duration)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.Value, nil)
	if err != nil {
		s.output = "creating request: " + err.Error()
		return
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		s.output = "making request: " + err.Error()
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.output = "reading body: " + err.Error()
		return
	}

	if strconv.Itoa(resp.StatusCode) == s.Expect {
		s.state = StateOK
		s.output = resp.Status
		return
	}

	if len(body) > maxBody {
		body = body[:maxBody]
	}

	s.state = StateCritical
	s.output = resp.Status + ": " + string(body)
}

func (s *Service) checkTCP() {
	s.state = StateUnknown
	s.output = "not working yet"
	s.lastCheck = time.Now()
}

func (s *Service) checkPING() {
	s.state = StateUnknown
	s.output = "not working yet"
	s.lastCheck = time.Now()
}
