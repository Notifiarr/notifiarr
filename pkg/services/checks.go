package services

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/exp"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

const (
	sslstring   = "SSL" // used for checking HTTPS certs
	expectdelim = ","   // extra (split) delimiter
)

type result struct {
	output string
	state  CheckState
}

// triggerCheck is used to signal the check of one service.
type triggerCheck struct {
	Source  website.EventType
	Service *Service
}

func (s *Service) Validate() error { //nolint:cyclop
	s.svc.Lock()
	defer s.svc.Unlock()

	s.svc.State = StateUnknown

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

		for _, code := range strings.Split(s.Expect, expectdelim) {
			if strings.EqualFold(code, sslstring) {
				s.validSSL = true
			}
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

// CheckOnly runs a service check and returns the result immediately.
// It is not otherwise stored anywhere.
func (s *Service) CheckOnly() *CheckResult {
	res := s.checkNow()

	return &CheckResult{
		Output: res.output,
		State:  res.state,
	}
}

func (s *Service) checkNow() (res *result) {
	switch s.Type {
	case CheckHTTP:
		return s.checkHTTP()
	case CheckTCP:
		return s.checkTCP()
	case CheckPING:
		return s.checkPING()
	case CheckPROC:
		return s.checkProccess()
	default:
		return nil
	}
}

func (s *Service) check() bool {
	return s.update(s.checkNow())
}

// Return true if the service state changed.
func (s *Service) update(res *result) bool {
	if res == nil {
		return false
	}

	exp.ServiceChecks.Add(s.Name+"&&Total", 1)
	exp.ServiceChecks.Add(s.Name+"&&"+res.state.String(), 1)
	//	exp.ServiceChecks.Add("Total Checks Run", 1)

	s.svc.Lock()
	defer s.svc.Unlock()

	if s.svc.LastCheck = time.Now().Round(time.Microsecond); s.svc.Since.IsZero() {
		s.svc.Since = s.svc.LastCheck
	}

	s.svc.Output = res.output

	if s.svc.State == res.state {
		s.svc.log.Printf("Service Checked: %s, state: %s for %v, output: %s",
			s.Name, s.svc.State, time.Since(s.svc.Since).Round(time.Second), s.svc.Output)
		return false
	}

	s.svc.log.Printf("Service Checked: %s, state: %s ~> %s, output: %s", s.Name, s.svc.State, res.state, s.svc.Output)
	s.svc.Since = s.svc.LastCheck
	s.svc.State = res.state

	return true
}

const maxBody = 150

func (s *Service) checkHTTP() *result {
	res := &result{
		state:  StateUnknown,
		output: "unknown",
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout.Duration)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.Value, nil)
	if err != nil {
		res.output = "creating request: " + RemoveSecrets(s.Value, err.Error())
		return res
	}

	resp, err := (&http.Client{Timeout: s.Timeout.Duration, Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !s.validSSL}, //nolint:gosec
	}}).Do(req)
	if err != nil {
		res.output = "making request: " + RemoveSecrets(s.Value, err.Error())
		return res
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		res.output = "reading body: " + RemoveSecrets(s.Value, err.Error())
		return res
	}

	for _, code := range strings.Split(s.Expect, expectdelim) {
		if strconv.Itoa(resp.StatusCode) == strings.TrimSpace(code) {
			res.state = StateOK
			res.output = resp.Status

			return res
		}
	}

	bodyStr := string(body)
	if len(bodyStr) > maxBody {
		bodyStr = bodyStr[:maxBody]
	}

	res.state = StateCritical
	res.output = resp.Status + ": " + strings.TrimSpace(RemoveSecrets(s.Value, bodyStr))

	return res
}

// RemoveSecrets removes secret token values in a message parsed from a url.
func RemoveSecrets(appURL, message string) string {
	url, err := url.Parse(appURL)
	if err != nil {
		return message
	}

	for _, keyName := range []string{"apikey", "token", "pass", "password", "secret"} {
		if secret := url.Query().Get(keyName); secret != "" {
			message = strings.ReplaceAll(message, secret, "********")
		}
	}

	return message
}

func (s *Service) checkTCP() *result {
	res := &result{
		state:  StateUnknown,
		output: "unknown",
	}

	switch conn, err := net.DialTimeout("tcp", s.Value, s.Timeout.Duration); {
	case err != nil:
		res.state = StateCritical
		res.output = "connection error: " + err.Error()
	case conn == nil:
		res.state = StateUnknown
		res.output = "connection failed, no specific error"
	default:
		conn.Close()

		res.state = StateOK
		res.output = "connected to port " + strings.Split(s.Value, ":")[1] + " OK"
	}

	return res
}

func (s *Service) checkPING() *result {
	return &result{
		state:  StateUnknown,
		output: "ping does not work yet",
	}
}
