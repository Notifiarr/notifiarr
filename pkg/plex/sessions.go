package plex

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Sessions is the config input data.
type Sessions struct {
	Name       string     `json:"server"`
	AccountMap []string   `json:"account_map"`
	Sessions   []*Session `json:"sessions"`
	XML        string     `json:"sessions_xml,omitempty"`
}

// ErrBadStatus is returned when plex returns an invalid status code.
var ErrBadStatus = fmt.Errorf("status code not 200")

// GetXMLSessions returns the Plex sessions in XML format.
func (s *Server) GetXMLSessions() (*Sessions, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout.Duration)
	defer cancel()

	body, err := s.getPlexURL(ctx, s.URL+"/status/sessions", map[string]string{"Accept": "application/xml"})
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, string(body))
	}

	var v struct {
		MediaContainer struct {
			Sessions []*Session `json:"Metadata"`
		} `json:"MediaContainer"`
	}

	if s.ReturnJSON {
		body, err := s.getPlexURL(ctx, s.URL+"/status/sessions", nil)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", err, string(body))
		}

		// log.Print("DEBUG PLEX PAYLOAD:\n", string(data))
		if err = json.Unmarshal(body, &v); err != nil {
			return nil, fmt.Errorf("parsing plex sessions: %w", err)
		}
	}

	return &Sessions{
		Name:       s.Name,
		AccountMap: strings.Split(s.AccountMap, "|"),
		XML:        string(body),
		Sessions:   v.MediaContainer.Sessions,
	}, nil
}

// GetSessions returns the Plex sessions in JSON format, no timeout.
func (s *Server) GetSessions() ([]*Session, error) {
	return s.GetSessionsWithContext(context.Background())
}

// GetSessionsWithContext returns the Plex sessions in JSON format.
func (s *Server) GetSessionsWithContext(ctx context.Context) ([]*Session, error) {
	var v struct {
		MediaContainer struct {
			Sessions []*Session `json:"Metadata"`
		} `json:"MediaContainer"`
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout.Duration)
	defer cancel()

	body, err := s.getPlexURL(ctx, s.URL+"/status/sessions", nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, string(body))
	}

	// log.Print("DEBUG PLEX PAYLOAD:\n", string(data))
	if err = json.Unmarshal(body, &v); err != nil {
		return nil, fmt.Errorf("parsing plex sessions: %w", err)
	}

	return v.MediaContainer.Sessions, nil
}

// KillSessionWithContext kills a Plex session.
func (s *Server) KillSessionWithContext(ctx context.Context, sessionID, reason string) error {
	if s == nil || s.URL == "" || s.Token == "" {
		return ErrNoURLToken
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout.Duration)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL+"/status/sessions/terminate", nil)
	if err != nil {
		return fmt.Errorf("creating http request: %w", err)
	}

	req.Header.Set("X-Plex-Token", s.Token)

	q := req.URL.Query()
	q.Add("sessionId", sessionID)
	q.Add("reason", reason)
	req.URL.RawQuery = q.Encode()

	resp, err := s.getClient().Do(req)
	if err != nil {
		return fmt.Errorf("making http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading http response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s", ErrBadStatus, string(body))
	}

	return nil
}

// KillSession kills a Plex session.
func (s *Server) KillSession(sessionID, reason string) error {
	return s.KillSessionWithContext(context.Background(), sessionID, reason)
}

func (s *Server) getClient() *http.Client {
	if s.client == nil {
		s.client = &http.Client{}
	}

	return s.client
}
