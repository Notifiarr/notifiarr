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

func (s *server) debughttplog(
	reqID string, resp *http.Response, url string, start time.Time, sentSize int, data []byte, body io.Reader,
) {
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

	size := len(data)
	if s.config.Apps.MaxBody > 0 && size > s.config.Apps.MaxBody {
		data = fmt.Appendf(data[:s.config.Apps.MaxBody],
			"\n<%d bytes truncated, max: %d>", size-s.config.Apps.MaxBody, s.config.Apps.MaxBody)
	}

	if len(data) == 0 {
		truncatedBody, bodySize := readBodyForLog(body, int64(s.config.Apps.MaxBody))
		mnd.Log.Debugf(reqID, "Sent GET Request to %s in %s, %s Response (%s):\n%s\n%s",
			url, time.Since(start).Round(time.Microsecond), mnd.FormatBytes(bodySize), status, headers.String(), truncatedBody)
	} else {
		truncatedBody, bodySize := readBodyForLog(body, int64(s.config.Apps.MaxBody))
		mnd.Log.Debugf(reqID, "Sent %s (%s decompressed) JSON Payload to %s in %s:\n%s\n%s Response (%s):\n%s\n%s",
			mnd.FormatBytes(sentSize), mnd.FormatBytes(size), url, time.Since(start).Round(time.Microsecond),
			string(data), mnd.FormatBytes(bodySize), status, headers.String(), truncatedBody)
	}
}

// readBodyForLog truncates the response body, or not, for the debug log. errors are ignored.
func readBodyForLog(body io.Reader, max int64) (string, int64) {
	if body == nil {
		return "", 0
	}

	if max > 0 {
		bodyBytes, _ := io.ReadAll(io.LimitReader(body, max))
		remaining, _ := io.Copy(io.Discard, body) // finish reading to the end.
		total := remaining + int64(len(bodyBytes))

		if remaining > 0 {
			return fmt.Sprintf("%s\n<%d response bytes truncated, max: %d>", string(bodyBytes), remaining, max), total
		}

		return string(bodyBytes), total
	}

	bodyBytes, _ := io.ReadAll(body)

	return string(bodyBytes), int64(len(bodyBytes))
}

func (s *server) debugLogResponseBody(
	reqID string,
	start time.Time,
	resp *http.Response,
	url string,
	sentSize int, // different from size, because it's the compressed size.
	data []byte, // uncompressed data.
	log bool,
) io.ReadCloser {
	var buf bytes.Buffer
	tee := io.TeeReader(resp.Body, &buf)

	defer resp.Body.Close() // close this since we return a fake one after logging.

	if log {
		defer s.debughttplog(reqID, resp, url, start, sentSize, data, tee)
	} else {
		defer s.debughttplog(reqID, resp, url, start, sentSize,
			fmt.Appendf(nil, "<data not logged, length:%d>", len(data)), tee)
	}

	return io.NopCloser(&buf)
}
