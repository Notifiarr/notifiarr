package website

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// SetValue sets a value stored in the website database.
func (s *Server) SetValue(key string, value []byte) error {
	return s.SetValues(map[string][]byte{key: value})
}

// SetValueContext sets a value stored in the website database.
func (s *Server) SetValueContext(ctx context.Context, key string, value []byte) error {
	return s.SetValuesContext(ctx, map[string][]byte{key: value})
}

// SetValues sets values stored in the website database.
func (s *Server) SetValues(values map[string][]byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Timeout.Duration)
	defer cancel()

	return s.SetValuesContext(ctx, values)
}

// SetValuesContext sets values stored in the website database.
func (s *Server) SetValuesContext(ctx context.Context, values map[string][]byte) error {
	for key, val := range values {
		if val != nil { // ignore nil byte slices.
			values[key] = []byte(base64.StdEncoding.EncodeToString(val))
		}
	}

	data, err := json.Marshal(map[string]interface{}{"fields": values})
	if err != nil {
		return fmt.Errorf("converting values to json: %w", err)
	}

	code, body, err := s.sendJSON(ctx, s.config.BaseURL+ClientRoute.Path("setStates"), data, true)
	if err != nil {
		return fmt.Errorf("invalid response (%d): %w", code, err)
	}

	_, err = unmarshalResponse(s.config.BaseURL+ClientRoute.Path("getStates"), code, body)

	return err
}

// DelValue deletes a value stored in the website database.
func (s *Server) DelValue(keys ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Timeout.Duration)
	defer cancel()

	return s.DelValueContext(ctx, keys...)
}

// DelValueContext deletes a value stored in the website database.
func (s *Server) DelValueContext(ctx context.Context, keys ...string) error {
	values := make(map[string]interface{})
	for _, key := range keys {
		values[key] = nil
	}

	data, err := json.Marshal(map[string]interface{}{"fields": values})
	if err != nil {
		return fmt.Errorf("converting values to json: %w", err)
	}

	code, body, err := s.sendJSON(ctx, s.config.BaseURL+ClientRoute.Path("setStates"), data, true)
	if err != nil {
		return fmt.Errorf("invalid response (%d): %w", code, err)
	}

	_, err = unmarshalResponse(s.config.BaseURL+ClientRoute.Path("setStates"), code, body)
	if err != nil {
		return err
	}

	return nil
}

// GetValue gets a value stored in the website database.
func (s *Server) GetValue(keys ...string) (map[string][]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Timeout.Duration)
	defer cancel()

	return s.GetValueContext(ctx, keys...)
}

// GetValueContext gets a value stored in the website database.
func (s *Server) GetValueContext(ctx context.Context, keys ...string) (map[string][]byte, error) {
	data, err := json.Marshal(map[string][]string{"fields": keys})
	if err != nil {
		return nil, fmt.Errorf("converting keys to json: %w", err)
	}

	code, body, err := s.sendJSON(ctx, s.config.BaseURL+ClientRoute.Path("getStates"), data, true)
	if err != nil {
		return nil, fmt.Errorf("invalid response (%d): %w", code, err)
	}

	resp, err := unmarshalResponse(s.config.BaseURL+ClientRoute.Path("getStates"), code, body)
	if err != nil {
		return nil, err
	}

	var output struct {
		LastUpdated time.Time         `json:"lastUpdated"`
		Fields      map[string][]byte `json:"fields"`
	}

	if err := json.Unmarshal(resp.Details.Response, &output); err != nil {
		return nil, fmt.Errorf("converting response values to json: %w", err)
	}

	for key, val := range output.Fields {
		b, err := base64.StdEncoding.DecodeString(string(val))
		if err != nil {
			return nil, fmt.Errorf("invalid base64 encoded data: %w", err)
		}

		output.Fields[key] = b
	}

	return output.Fields, nil
}
