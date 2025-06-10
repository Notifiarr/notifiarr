// Package endpoints allows the user to configure a list of urls, along with cron schedules.
// The app then polls the urls according to their schedule. The URL response is gzipped,
// base64 encoded then sent to the website for parsing and notification handling.
// "Endpoint URL Passthrough" is the name.
package endpoints

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/endpoints/epconfig"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

// globalTimeout is the max duration an endpoint url http request may elapse.
const globalTimeout = 5 * time.Minute

// Action contains the exported methods for this package.
type Action struct {
	conf *common.Config
	list []*Schedule
	urls []*epconfig.Endpoint
}

// Schedule is used to schedule endpoint url queries.
type Schedule struct {
	*epconfig.Endpoint
	ch     chan *common.ActionInput
	client *http.Client
	conf   *common.Config
}

// Schedules is a slice of Schedule.
type Schedules []*Schedule

// New configures the endpoints module.
func New(config *common.Config, endpoints []*epconfig.Endpoint) *Action {
	return &Action{
		conf: config,
		urls: endpoints,
	}
}

// List returns a list of scheduled endpoint url pollers that can be executed ad-hoc.
func (a *Action) List() Schedules {
	return a.list
}

// Create initializes the endpoint url poller.
func (a *Action) Create() {
	for _, endpoint := range a.urls {
		if endpoint.Method == "" {
			endpoint.Method = http.MethodGet
		}

		if endpoint.Name == "" {
			endpoint.Name = endpoint.URL
		}

		if endpoint.Header == nil {
			endpoint.Header = make(http.Header)
		}

		if endpoint.Query == nil {
			endpoint.Query = make(url.Values)
		}

		schedule := NewSchedule(endpoint, a.conf)
		a.list = append(a.list, schedule)

		// Schedule this cron job.
		a.conf.Add(&common.Action{
			Name: common.TriggerName(fmt.Sprintf("Polling endpoint URL '%s'", endpoint.Name)),
			Fn:   schedule.run,
			C:    schedule.ch,
			J:    &endpoint.CronJob,
		})
	}

	if len(a.urls) > 0 {
		mnd.Log.Printf("==> Endpoint URL Passthrough Enabled: %d URL(s)", len(a.urls))
	}
}

func NewSchedule(endpoint *epconfig.Endpoint, conf *common.Config) *Schedule {
	return &Schedule{
		conf:     conf,
		Endpoint: endpoint,
		ch:       make(chan *common.ActionInput, 1),
		client: &http.Client{
			Timeout: globalTimeout,
			Transport: apps.NewMetricsRoundTripper("endpoints", &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: !endpoint.ValidSSL,
				},
			}),
			CheckRedirect: endpoint.CheckRedirect(),
		},
	}
}

// Get a schedule by name or URL.
func (s Schedules) Get(nameOrURL string) *Schedule {
	for _, schedule := range s {
		if schedule.URL == nameOrURL || schedule.Name == nameOrURL {
			return schedule
		}
	}

	return nil
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
		mnd.Log.Errorf("Endpoint URL '%s' failed: %v", s.Name, err)
		return
	}

	if s.Template == mnd.False {
		return // template is false, do not send anything to website.
	}

	website.Site.SendData(&website.Request{
		Route:      website.EndpointRoute,
		Event:      input.Type,
		LogPayload: true,
		Payload: map[string]any{
			"name":     s.Name,
			"url":      s.URL,
			"template": s.Template,
			"gzb64":    body,
			"header":   header,
			"status":   code,
		},
	})
}

func (s *Schedule) getURLBody(ctx context.Context) (http.Header, int, string, error) {
	req, err := http.NewRequestWithContext(ctx, s.Method, s.GetURL(), bytes.NewBufferString(s.Body))
	if err != nil {
		return nil, 0, "", fmt.Errorf("creating request: %w", err)
	}

	for header, value := range s.Header {
		req.Header[header] = value
	}

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

// Verify the interface is satisfied.
var _ = common.Create(&Action{})
