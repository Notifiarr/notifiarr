package plex

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Sessions is the config input data.
type Sessions struct {
	Name     string     `json:"server"`
	HostID   string     `json:"hostId"`
	Sessions []*Session `json:"sessions"`
	Updated  structDur  `json:"updateTime,omitempty"`
}

// ErrBadStatus is returned when plex returns an invalid status code.
var ErrBadStatus = fmt.Errorf("status code not 200")

// GetSessions returns the Plex sessions in JSON format, no timeout.
func (s *Server) GetSessions() (*Sessions, error) {
	return s.GetSessionsWithContext(context.Background())
}

// GetSessionsWithContext returns the Plex sessions in JSON format.
func (s *Server) GetSessionsWithContext(ctx context.Context) (*Sessions, error) {
	if !s.Configured() {
		return nil, ErrNoURLToken
	}

	var (
		v struct {
			MediaContainer struct {
				Sessions []*Session `json:"Metadata"`
			} `json:"MediaContainer"`
		}
		sessions = &Sessions{Name: s.Name}
	)

	ctx, cancel := context.WithTimeout(ctx, s.Timeout.Duration)
	defer cancel()

	body, err := s.getPlexURL(ctx, s.URL+"/status/sessions", nil)
	if err != nil {
		return sessions, fmt.Errorf("%w: %s", err, string(body))
	}

	// fmt.Println("DEBUG PLEX PAYLOAD:", string(body))

	if err = json.Unmarshal(body, &v); err != nil {
		return sessions, fmt.Errorf("parsing plex sessions (TRY UPGRADING PLEX): %w", err)
	}

	sessions.Sessions = v.MediaContainer.Sessions

	return sessions, nil
}

// KillSessionWithContext kills a Plex session.
func (s *Server) KillSessionWithContext(ctx context.Context, sessionID, reason string) ([]byte, error) {
	if !s.Configured() {
		return nil, ErrNoURLToken
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout.Duration)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL+"/status/sessions/terminate", nil)
	if err != nil {
		return nil, fmt.Errorf("creating http request: %w", err)
	}

	req.Header.Set("X-Plex-Token", s.Token)

	q := req.URL.Query()
	q.Add("sessionId", sessionID)
	q.Add("reason", reason)
	req.URL.RawQuery = q.Encode()

	resp, err := s.getClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("making http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return body, fmt.Errorf("reading http response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return body, ErrBadStatus
	}

	return body, nil
}

// KillSession kills a Plex session.
func (s *Server) KillSession(sessionID, reason string) ([]byte, error) {
	return s.KillSessionWithContext(context.Background(), sessionID, reason)
}

func (s *Server) getClient() *http.Client {
	if s.client == nil {
		s.client = &http.Client{}
	}

	return s.client
}
