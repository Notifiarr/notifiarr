// Package endpoints allows the user to configure a list of urls, along with cron schedules.
// The app then polls the urls according to their schedule. The URL response is sent to the
// website for parsing and notification handling. "Endpoint URL Passthrough" is the name.
package endpoints

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

// globalTimeout is the max duration an endpoint url http request may elapse.
const globalTimeout = 5 * time.Minute

// Action contains the exported methods for this package.
type Action struct {
	conf *common.Config
	list []*Schedule
	urls []*Endpoint
}

// Endpoint contains the cronjob definition and url query parameters.
// This is the input data to poll a url on a frequency.
type Endpoint struct {
	Query  url.Values  `json:"query"  toml:"query"  xml:"query"  yaml:"query"`
	Header http.Header `json:"header" toml:"header" xml:"header" yaml:"header"`
	URL    string      `json:"url"    toml:"url"    xml:"url"    yaml:"url"`
	Method string      `json:"method" toml:"method" xml:"method" yaml:"method"`
	Body   string      `json:"body"   toml:"body"   xml:"body"   yaml:"body"`
	Follow bool        `json:"follow" toml:"follow" xml:"follow" yaml:"follow"` // redirects
	url    string      // url + query
	common.CronJob
}

// Schedule is used to schedule endpoint url queries.
type Schedule struct {
	*Endpoint
	ch     chan *common.ActionInput
	client *http.Client
	conf   *common.Config
}

// New configures the endpoints module.
func New(config *common.Config, endpoints []*Endpoint) *Action {
	return &Action{
		conf: config,
		urls: endpoints,
	}
}

// List returns a list of scheduled endpoint url pollers that can be executed ad-hoc.
func (a *Action) List() []*Schedule {
	return a.list
}

// Create initializes the endpoint url poller.
func (a *Action) Create() {
	for _, endpoint := range a.urls {
		if endpoint.url = endpoint.URL; len(endpoint.Query) > 0 {
			endpoint.url += "?" + endpoint.Query.Encode()
		}

		if endpoint.Method == "" {
			endpoint.Method = http.MethodGet
		}

		schedule := endpoint.schedule(a.conf)
		a.list = append(a.list, schedule)

		// Schedule this cron job.
		a.conf.Add(&common.Action{
			Name: common.TriggerName(fmt.Sprintf("Polling endpoint URL '%s'", endpoint.URL)),
			Fn:   schedule.run,
			C:    schedule.ch,
			J:    &endpoint.CronJob,
		})
	}

	if len(a.urls) > 0 {
		a.conf.Printf("==> Endpoint URL Passthrough Enabled: %d URL(s)", len(a.urls))
	}
}

// Run polls an endpoint url and relays the data to the website.
func (s *Schedule) Run(input *common.ActionInput) {
	if s.ch == nil {
		return
	}

	s.ch <- input // fires run() below.
}

// run responds to the channel that Run() fired into.
func (s *Schedule) run(ctx context.Context, input *common.ActionInput) {
	header, code, body, err := s.getURLBody(ctx)
	if err != nil {
		s.conf.Errorf("Endpoint URL '%s' failed: %v", s.URL, err)
		return
	}

	s.conf.SendData(&website.Request{
		Route:      website.TestRoute,
		Event:      input.Type,
		LogPayload: true,
		Payload:    map[string]any{"gzb64": body, "header": header, "status": code},
	})
}

func (s *Schedule) getURLBody(ctx context.Context) (http.Header, int, string, error) {
	req, err := http.NewRequestWithContext(ctx, s.Method, s.url, bytes.NewBufferString(s.Body))
	if err != nil {
		return nil, 0, "", fmt.Errorf("creating request: %w", err)
	}

	req.Header = s.Header

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, 0, "", fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	var body bytes.Buffer
	gzwriter := gzip.NewWriter(&body)

	if _, err = io.Copy(gzwriter, resp.Body); err != nil {
		return nil, 0, "", fmt.Errorf("reading response body: %w", err)
	}

	_ = gzwriter.Close()

	return resp.Header, resp.StatusCode, base64.StdEncoding.EncodeToString(body.Bytes()), nil
}

func (e *Endpoint) schedule(conf *common.Config) *Schedule {
	return &Schedule{
		conf:     conf,
		Endpoint: e,
		ch:       make(chan *common.ActionInput, 1),
		client: &http.Client{
			Timeout:       globalTimeout,
			Transport:     apps.NewMetricsRoundTripper("endpoints", nil),
			CheckRedirect: e.checkRedirect(),
		},
	}
}

func (e *Endpoint) checkRedirect() func(_ *http.Request, _ []*http.Request) error {
	if e.Follow {
		return nil
	}

	return func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
}

// Verify the interface is satisfied.
var _ = common.Create(&Action{})
