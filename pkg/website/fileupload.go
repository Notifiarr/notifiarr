package website

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
)

func (s *server) sendFile(ctx context.Context, uri string, file *UploadFile) (*Response, error) {
	mnd.Log.Trace(mnd.GetID(ctx), "start: sendFile")
	defer mnd.Log.Trace(mnd.GetID(ctx), "end: sendFile")

	form, contentType, err := s.createFileUpload(file)
	if err != nil {
		return nil, err
	}

	sent := form.Len()
	url := BaseURL + uri

	// Send the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, form)
	if err != nil {
		return nil, fmt.Errorf("creating http request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Api-Key", s.config.Apps.APIKey)

	start := time.Now()
	msg := fmt.Sprintf("Upload %s, %d bytes", file.FileName, sent)

	resp, err := s.client.Do(req)
	if err != nil {
		s.debughttplog(nil, url, start, msg, nil)
		return nil, fmt.Errorf("making http request: %w", err)
	}
	defer resp.Body.Close()

	reader := resp.Body

	if mnd.Log.DebugEnabled() {
		reader = s.debugLogResponseBody(start, resp, url, []byte(msg), true)
	}

	response, err := unmarshalResponse(url, resp.StatusCode, reader)
	if response != nil {
		response.sent = sent
	}

	return response, err
}

func (s *server) createFileUpload(file *UploadFile) (*bytes.Buffer, string, error) {
	// Create a new multipart writer with the buffer
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	if host := s.hostInfoNoError(); host != nil {
		// Since we can't send the normal hostInfo json payload,
		// we have to shove some things into the form fields.
		if err := writer.WriteField("hostname", host.Hostname); err != nil {
			return nil, "", fmt.Errorf("adding variable to form buffer: %w", err)
		}

		_ = writer.WriteField("hostId", host.HostID)
		_ = writer.WriteField("os", host.OS)
	}

	// Create a new form field
	fw, err := writer.CreateFormFile("file", file.FileName+".gz")
	if err != nil {
		return nil, "", fmt.Errorf("creating form buffer: %w", err)
	}

	compress := gzip.NewWriter(fw)
	compress.Header.Name = file.FileName

	// Copy the contents of the file to the form field with compression.
	if _, err := io.Copy(compress, file); err != nil {
		return nil, "", fmt.Errorf("filling form buffer: %w", err)
	}

	// Close the compressor and multipart writer to finalize the request.
	compress.Close()
	writer.Close()
	file.Close() // Close the file too.

	return &buf, writer.FormDataContentType(), nil
}
