package services

import (
	"context"
	"crypto/tls"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

const (
	sslstring   = "SSL" // used for checking HTTPS certs
	expectdelim = ","   // extra (split) delimiter
	maxOutput   = 170   // maximum length of output.
)

type result struct {
	output *Output
	state  CheckState
}

// triggerCheck is used to signal the check of one service.
type triggerCheck struct {
	Source  website.EventType
	Service *Service
}

func (s *ServiceConfig) Validate() error { //nolint:cyclop
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

		for code := range strings.SplitSeq(s.Expect, expectdelim) {
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
	case CheckPING, CheckICMP:
		if err := s.checkPingValues(s.Type == CheckICMP); err != nil {
			return err
		}
	default:
		return ErrInvalidType
	}

	if s.Timeout.Duration == 0 {
		s.Timeout.Duration = DefaultTimeout
	} else if s.Timeout.Duration < MinimumTimeout {
		s.Timeout.Duration = MinimumTimeout
	}

	if s.Interval.Duration != 0 && s.Interval.Duration < MinimumCheckInterval {
		s.Interval.Duration = MinimumCheckInterval
	}

	s.validated = true

	return nil
}

// CheckOnly runs a service check and returns the result immediately.
// It is not otherwise stored anywhere.
func (s *ServiceConfig) CheckOnly(ctx context.Context) *CheckResult {
	if err := s.Validate(); err != nil {
		return &CheckResult{
			Output:   &Output{str: err.Error()},
			State:    StateCritical,
			Metadata: s.Tags,
		}
	}

	service := &Service{ServiceConfig: s, log: mnd.Log}
	res := service.checkNow(ctx)

	return &CheckResult{
		Output:   res.output,
		State:    res.state,
		Metadata: s.Tags,
	}
}

func (s *Service) checkNow(ctx context.Context) *result {
	switch s.Type {
	case CheckHTTP:
		return s.checkHTTP(ctx)
	case CheckTCP:
		return s.checkTCP(ctx)
	case CheckPING, CheckICMP:
		return s.checkPING()
	case CheckPROC:
		return s.checkProccess(ctx)
	default:
		return nil
	}
}

func (s *Service) check(ctx context.Context) bool {
	return s.update(s.checkNow(ctx))
}

// Return true if the service state changed.
func (s *Service) update(res *result) bool {
	if res == nil {
		return false
	}

	mnd.ServiceChecks.Add(s.Name+"&&Total", 1)
	mnd.ServiceChecks.Add(s.Name+"&&"+res.state.String(), 1)
	//	mnd.ServiceChecks.Add("Total Checks Run", 1)

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.LastCheck = time.Now().Round(time.Microsecond); s.Since.IsZero() {
		s.Since = s.LastCheck
	}

	s.Output = res.output

	if s.State == res.state {
		s.log.Printf("Service Checked: %s, state: %s for %v, output: %s",
			s.Name, s.State, time.Since(s.Since).Round(time.Second), s.Output)
		return false
	}

	s.log.Printf("Service Checked: %s, state: %s ~> %s, output: %s", s.Name, s.State, res.state, s.Output)
	s.Since = s.LastCheck
	s.State = res.state

	return true
}

// checkHTTPReq builds the client and request for the http service check.
func (s *Service) checkHTTPReq(ctx context.Context) (*http.Client, *http.Request, error) {
	// Allow adding headers by appending them after a pipe symbol.
	splitVal := strings.Split(s.Value, "|")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, splitVal[0], nil)
	if err != nil {
		return nil, nil, err //nolint:wrapcheck // handled by caller
	}

	for _, val := range splitVal[1:] {
		// s.Value: http://url.com|header=value|another-header=val
		if sv := strings.SplitN(val, ":", 2); len(sv) == 2 { //nolint:mnd
			req.Header.Add(sv[0], sv[1])

			if strings.EqualFold(sv[0], "host") {
				req.Host = sv[1] // https://github.com/golang/go/issues/29865
			}
		}
	}

	return &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: s.Timeout.Duration, Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: !s.validSSL}, //nolint:gosec
		},
	}, req, nil
}

func (s *Service) checkHTTP(ctx context.Context) *result {
	res := &result{
		state:  StateUnknown,
		output: &Output{str: "unknown"},
	}

	if s.Timeout.Duration > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, s.Timeout.Duration)
		defer cancel()
	}

	client, req, err := s.checkHTTPReq(ctx)
	if err != nil {
		res.output = &Output{str: "creating request: " + RemoveSecrets(s.Value, err.Error())}
		return res
	}

	// If there is an error at this point it's a bad request.
	res.state = StateCritical

	resp, err := client.Do(req)
	if err != nil {
		res.output = &Output{str: "making request: " + RemoveSecrets(s.Value, err.Error())}
		return res
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		res.output = &Output{str: "reading body: " + RemoveSecrets(s.Value, err.Error())}
		return res
	}

	for code := range strings.SplitSeq(s.Expect, expectdelim) {
		if strconv.Itoa(resp.StatusCode) == strings.TrimSpace(code) {
			res.state = StateOK
			res.output = &Output{str: resp.Status}

			return res
		}
	}

	// Reduce the size of the string before processing it to speed things up on large body outputs.
	if len(res.output.str) > maxOutput+maxOutput {
		res.output.str = res.output.str[:maxOutput+maxOutput]
	}

	res.state = StateCritical
	res.output = &Output{esc: true, str: resp.Status + ": " + strings.TrimSpace(
		html.EscapeString(strings.Join(strings.Fields(RemoveSecrets(s.Value, string(body))), " ")))}

	// Reduce the string to the final max length.
	// We do it this way so all secrets are properly escaped before string splitting.
	if len(res.output.str) > maxOutput {
		res.output.str = res.output.str[:maxOutput]
	}

	return res
}

// RemoveSecrets removes secret token values in a message parsed from a url.
func RemoveSecrets(appURL, message string) string {
	url, err := url.Parse(strings.SplitN(appURL, "|", 2)[0]) //nolint:mnd
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

func (s *ServiceConfig) checkTCP(ctx context.Context) *result {
	res := &result{
		state:  StateUnknown,
		output: &Output{str: "unknown"},
	}

	dialer := &net.Dialer{Timeout: s.Timeout.Duration}
	switch conn, err := dialer.DialContext(ctx, "tcp", s.Value); {
	case err != nil:
		res.state = StateCritical
		res.output = &Output{str: "connection error: " + err.Error()}
	case conn == nil:
		res.state = StateUnknown
		res.output = &Output{str: "connection failed, no specific error"}
	default:
		conn.Close()

		res.state = StateOK
		res.output = &Output{str: "connected to port " + strings.Split(s.Value, ":")[1] + " OK"}
	}

	return res
}

func (s *Service) Due(now time.Time) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Interval.Duration > 0 && now.Sub(s.LastCheck) > s.Interval.Duration
}
