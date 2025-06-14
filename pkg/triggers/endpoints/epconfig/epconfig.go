// Package epconfig contains the config input for the endpoints package.
// This is in its own package to avoid an import cycle with clientinfo.
package epconfig

import (
	"crypto/tls"
	"net/http"
	"net/url"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common/scheduler"
	"golift.io/cnfg"
)

// Endpoint contains the cronjob definition and url query parameters.
// This is the input data to poll a url on a frequency.
type Endpoint struct {
	Query    url.Values    `json:"query"    toml:"query"     xml:"query"     yaml:"query"`
	Header   http.Header   `json:"header"   toml:"header"    xml:"header"    yaml:"header"`
	Template string        `json:"template" toml:"template"  xml:"template"  yaml:"template"`
	Name     string        `json:"name"     toml:"name"      xml:"name"      yaml:"name"`
	URL      string        `json:"url"      toml:"url"       xml:"url"       yaml:"url"`
	Method   string        `json:"method"   toml:"method"    xml:"method"    yaml:"method"`
	Body     string        `json:"body"     toml:"body"      xml:"body"      yaml:"body"`
	Follow   bool          `json:"follow"   toml:"follow"    xml:"follow"    yaml:"follow"`   // redirects
	ValidSSL bool          `json:"validSsl" toml:"valid_ssl" xml:"valid_ssl" yaml:"validSsl"` // https only
	Timeout  cnfg.Duration `json:"timeout"  toml:"timeout"   xml:"timeout"   yaml:"timeout"`
	url      string        // url + query
	scheduler.CronJob
}

func (e *Endpoint) GetClient() *http.Client {
	return &http.Client{
		Timeout: e.Timeout.Duration,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: !e.ValidSSL,
			},
		},
		CheckRedirect: e.checkRedirect(),
	}
}

// CheckRedirect returns a function to facilitate the follow redirect setting.
func (e *Endpoint) checkRedirect() func(_ *http.Request, _ []*http.Request) error {
	if e.Follow {
		return nil
	}

	return func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
}

// GetURL returns the url with a provided query string appended.
func (e *Endpoint) GetURL() string {
	if e.url != "" {
		return e.url
	}

	if e.url = e.URL; len(e.Query) > 0 {
		e.url += "?" + e.Query.Encode()
	}

	return e.url
}

func (e *Endpoint) SetHeaders(req *http.Request) {
	for header, v := range e.Header {
		for _, val := range v {
			req.Header.Set(header, val)
		}
	}
}
