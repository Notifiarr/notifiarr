// Package epconfig contains the config input for the endpoints package.
// This is in its own package to avoid an import cycle with clientinfo.
package epconfig

import (
	"net/http"
	"net/url"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common/scheduler"
)

// Endpoint contains the cronjob definition and url query parameters.
// This is the input data to poll a url on a frequency.
type Endpoint struct {
	Query  url.Values  `json:"query"  toml:"query"  xml:"query"  yaml:"query"`
	Header http.Header `json:"header" toml:"header" xml:"header" yaml:"header"`
	Name   string      `json:"name"   toml:"name"   xml:"name"   yaml:"name"`
	URL    string      `json:"url"    toml:"url"    xml:"url"    yaml:"url"`
	Method string      `json:"method" toml:"method" xml:"method" yaml:"method"`
	Body   string      `json:"body"   toml:"body"   xml:"body"   yaml:"body"`
	Follow bool        `json:"follow" toml:"follow" xml:"follow" yaml:"follow"` // redirects
	url    string      // url + query
	scheduler.CronJob
}

// CheckRedirect returns a function to facilitate the follow redirect setting.
func (e *Endpoint) CheckRedirect() func(_ *http.Request, _ []*http.Request) error {
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
