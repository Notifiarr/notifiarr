package plex

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type Sessions struct {
	Name       string    `json:"server"`
	AccountMap []string  `json:"account_map"`
	Size       int       `json:"size"`
	Sessions   []Session `json:"sessions"`
	XML        string    `json:"sessions_xml"`
}

type Session interface{}

var ErrBadStatus = fmt.Errorf("status code not 200")

func (s *Server) GetSessions() (*Sessions, error) {
	if s == nil || s.URL == "" || s.Token == "" {
		return nil, ErrNoURLToken
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout.Duration)
	defer cancel()

	xml, err := s.getPlexSessions(ctx, map[string]string{"Accept": "application/xml"})
	if err != nil {
		return nil, err
	}

	var v struct {
		MediaContainer struct {
			Size     int       `json:"size"`
			Sessions []Session `json:"Metadata"`
		} `json:"MediaContainer"`
	}

	if s.ReturnJSON {
		data, err := s.getPlexSessions(ctx, map[string]string{"Accept": "application/json"})
		if err != nil {
			return nil, err
		}

		if err = json.Unmarshal(data, &v); err != nil {
			return nil, fmt.Errorf("parsing plex sessions: %w", err)
		}
	}

	return &Sessions{
		Name:       s.Name,
		Size:       v.MediaContainer.Size,
		AccountMap: strings.Split(s.AccountMap, "|"),
		XML:        string(xml),
		Sessions:   v.MediaContainer.Sessions,
	}, nil
}

func (s *Server) getPlexSessions(ctx context.Context, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL+"/status/sessions", nil)
	if err != nil {
		return nil, fmt.Errorf("creating http request: %w", err)
	}

	req.Header.Set("X-Plex-Token", s.Token)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := s.getClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("making http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading http response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return body, ErrBadStatus
	}

	return body, nil
}

func (s *Server) getClient() *http.Client {
	if s.client == nil {
		s.client = &http.Client{}
	}

	return s.client
}
