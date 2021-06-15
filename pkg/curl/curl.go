package curl

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const timeout = 15 * time.Second

// Get returns an http response and body.
func Get(url string) (*http.Response, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("creating http request: %w", err)
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("making http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, body, fmt.Errorf("reading body: %w", err)
	}

	return resp, body, nil
}

// Print prints an http response.
//nolint:forbidigo
func Print(resp *http.Response, body []byte) {
	fmt.Println(resp.Status)

	for header, value := range resp.Header {
		for _, v := range value {
			fmt.Println(header + ": " + v)
		}
	}

	fmt.Println("")
	fmt.Println(string(body))
}
