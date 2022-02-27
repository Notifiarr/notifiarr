package plex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
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
			//nolint:tagliatelle
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

	if err = json.Unmarshal(body, &v); err != nil {
		return sessions, fmt.Errorf("parsing plex sessions (TRY UPGRADING PLEX): %w: %s", err, string(body))
	}

	sessions.Sessions = v.MediaContainer.Sessions

	return sessions, nil
}

// KillSessionWithContext kills a Plex session.
func (s *Server) KillSessionWithContext(ctx context.Context, sessionID, reason string) ([]byte, error) {
	if !s.Configured() {
		return nil, ErrNoURLToken
	}

	params := make(url.Values)
	params.Add("sessionId", sessionID)
	params.Add("reason", reason)

	ctx, cancel := context.WithTimeout(ctx, s.Timeout.Duration)
	defer cancel()

	body, err := s.getPlexURL(ctx, s.URL+"/status/sessions/terminate", params)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, string(body))
	}

	return body, nil
}

// KillSession kills a Plex session.
func (s *Server) KillSession(sessionID, reason string) ([]byte, error) {
	return s.KillSessionWithContext(context.Background(), sessionID, reason)
}
