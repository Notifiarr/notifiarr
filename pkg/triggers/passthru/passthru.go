package passthru

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

// New configures the passthru module.
func New(config *common.Config, endpoints []*Endpoint) *Action {
	return &Action{
		conf: config,
		urls: endpoints,
	}
}

// Run polls an endpoint url and relays the data to the website.
func (t *Schedule) Run(input *common.ActionInput) {
	if t.ch == nil {
		return
	}

	t.ch <- input // fires run() below.
}

// run responds to the channel that Run() fired into.
func (t *Schedule) run(ctx context.Context, input *common.ActionInput) {
	req, err := http.NewRequestWithContext(ctx, t.Method, t.url, bytes.NewBufferString(t.Body))
	if err != nil {
		t.conf.Errorf("Creating request for passthrough URL '%s': %v", t.URL, err)
		return
	}

	req.Header = t.Header

	resp, err := t.client.Do(req)
	if err != nil {
		t.conf.Errorf("Requesting passthrough URL '%s': %v", t.URL, err)
		return
	}
	defer resp.Body.Close()

	var body bytes.Buffer
	if _, err := io.Copy(gzip.NewWriter(&body), resp.Body); err != nil {
		t.conf.Errorf("Reading passthrough URL '%s' response body: %v", t.URL, err)
		return
	}

	t.conf.SendData(&website.Request{
		Route:      website.TestRoute,
		Event:      website.EventSched,
		LogPayload: true,
		Payload: map[string]any{
			"gzb64body": base64.StdEncoding.EncodeToString(body.Bytes()),
			"headers":   resp.Header,
			"status":    resp.StatusCode,
		},
	})
}

// List returns a list of scheduled endpoint url pollers that can be executed ad-hoc.
func (a *Action) List() []*Schedule {
	return a.list
}

// Create initializes the endpoint url poller.
func (a *Action) Create() {
	a.create()
}

// Verify the interface is satisfied.
var _ = common.Create(&Action{})

func (a *Action) create() {
	httpClient := &http.Client{
		Timeout:   globalTimeout,
		Transport: apps.NewMetricsRoundTripper("passthrough", nil),
	}

	for _, endpoint := range a.urls {
		if endpoint.url = endpoint.URL; len(endpoint.Query) > 0 {
			endpoint.url += "?" + endpoint.Query.Encode()
		}

		if endpoint.Method == "" {
			endpoint.Method = http.MethodGet
		}

		schedule := &Schedule{
			Endpoint: endpoint,
			ch:       make(chan *common.ActionInput, 1),
			client:   httpClient,
			conf:     a.conf,
		}
		a.list = append(a.list, schedule)

		// Schedule this cron job.
		a.conf.Add(&common.Action{
			Name: common.TriggerName(fmt.Sprintf("Polling passthrough URL '%s'", endpoint.URL)),
			Fn:   schedule.run,
			C:    schedule.ch,
			J:    &endpoint.CronJob,
		})
	}

	a.conf.Printf("==> Passthrough Poller Enabled: %d URLs provided", len(a.urls))
}
