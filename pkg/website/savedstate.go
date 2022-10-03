package website

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// SetValueContext sets a value stored in the website database.
func (s *Server) SetValueContext(ctx context.Context, key string, value []byte) error {
	return s.SetValuesContext(ctx, map[string][]byte{key: value})
}

// SetValuesContext sets values stored in the website database.
func (s *Server) SetValuesContext(ctx context.Context, values map[string][]byte) error {
	for key, val := range values {
		if val != nil { // ignore nil byte slices.
			values[key] = []byte(base64.StdEncoding.EncodeToString(val))
		}
	}

	resp, err := s.GetData(&Request{
		Route:      ClientRoute,
		Event:      "setStates",
		Payload:    map[string]interface{}{"fields": values},
		LogPayload: true,
	})
	if err != nil {
		return fmt.Errorf("invalid response: %w: %s", err, resp)
	}

	return nil
}

// DelValueContext deletes a value stored in the website database.
func (s *Server) DelValueContext(ctx context.Context, keys ...string) error {
	values := make(map[string]interface{})
	for _, key := range keys {
		values[key] = nil
	}

	resp, err := s.GetData(&Request{
		Route:      ClientRoute,
		Event:      "setStates",
		Payload:    map[string]interface{}{"fields": values},
		LogPayload: true,
	})
	if err != nil {
		return fmt.Errorf("invalid response: %w: %s", err, resp)
	}

	return nil
}

// GetValueContext gets a value stored in the website database.
func (s *Server) GetValueContext(ctx context.Context, keys ...string) (map[string][]byte, error) {
	resp, err := s.GetData(&Request{
		Route:      ClientRoute,
		Event:      "getStates",
		Payload:    map[string][]string{"fields": keys},
		LogPayload: true,
	})
	if err != nil {
		return nil, fmt.Errorf("invalid response: %w: %s", err, resp)
	}

	var output struct {
		LastUpdated time.Time         `json:"lastUpdated"`
		Fields      map[string][]byte `json:"fields"`
	}

	if err := json.Unmarshal(resp.Details.Response, &output); err != nil {
		return nil, fmt.Errorf("converting response values to json: %w", err)
	}

	for key, val := range output.Fields {
		data, err := base64.StdEncoding.DecodeString(string(val))
		if err != nil {
			return nil, fmt.Errorf("invalid base64 encoded data: %w", err)
		}

		output.Fields[key] = data
	}

	return output.Fields, nil
}

/**/
