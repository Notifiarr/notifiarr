package website

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
)

func (s *server) debughttplog(resp *http.Response, url string, start time.Time, data string, body io.Reader) {
	var headers strings.Builder
	status := "0"

	if resp != nil {
		status = resp.Status

		for k, vs := range resp.Header {
			for _, v := range vs {
				headers.WriteString(k + ": " + v + "\n")
			}
		}
	}

	if s.config.Apps.MaxBody > 0 && len(data) > s.config.Apps.MaxBody {
		data = fmt.Sprintf("%s <data truncated, max: %d>", data[:s.config.Apps.MaxBody], s.config.Apps.MaxBody)
	}

	if data == "" {
		truncatedBody, bodySize := readBodyForLog(body, int64(s.config.Apps.MaxBody))
		mnd.Log.Debugf("Sent GET Request to %s in %s, %s Response (%s):\n%s\n%s",
			url, time.Since(start).Round(time.Microsecond), mnd.FormatBytes(bodySize), status, headers.String(), truncatedBody)
	} else {
		truncatedBody, bodySize := readBodyForLog(body, int64(s.config.Apps.MaxBody))
		mnd.Log.Debugf("Sent %s JSON Payload to %s in %s:\n%s\n%s Response (%s):\n%s\n%s",
			mnd.FormatBytes(len(data)), url, time.Since(start).Round(time.Microsecond),
			data, mnd.FormatBytes(bodySize), status, headers.String(), truncatedBody)
	}
}

// readBodyForLog truncates the response body, or not, for the debug log. errors are ignored.
func readBodyForLog(body io.Reader, max int64) (string, int64) {
	if body == nil {
		return "", 0
	}

	if max > 0 {
		limitReader := io.LimitReader(body, max)
		bodyBytes, _ := io.ReadAll(limitReader)
		remaining, _ := io.Copy(io.Discard, body) // finish reading to the end.
		total := remaining + int64(len(bodyBytes))

		if remaining > 0 {
			return fmt.Sprintf("%s <body truncated, max: %d>", string(bodyBytes), max), total
		}

		return string(bodyBytes), total
	}

	bodyBytes, _ := io.ReadAll(body)

	return string(bodyBytes), int64(len(bodyBytes))
}

func (s *server) debugLogResponseBody(
	start time.Time,
	resp *http.Response,
	url string,
	data []byte,
	log bool,
) io.ReadCloser {
	var buf bytes.Buffer
	tee := io.TeeReader(resp.Body, &buf)

	defer resp.Body.Close() // close this since we return a fake one after logging.

	if log {
		defer s.debughttplog(resp, url, start, string(data), tee)
	} else {
		defer s.debughttplog(resp, url, start, fmt.Sprintf("<data not logged, length:%d>", len(data)), tee)
	}

	return io.NopCloser(&buf)
}
