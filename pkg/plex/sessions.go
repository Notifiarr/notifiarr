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

// MarkPlayedWithContext marks a video as played.
func (s *Server) MarkPlayedWithContext(ctx context.Context, key string) error {
	if !s.Configured() {
		return ErrNoURLToken
	}

	params := make(url.Values)
	params.Add("identifier", "com.plexapp.plugins.library")
	params.Add("key", key)

	ctx, cancel := context.WithTimeout(ctx, s.Timeout.Duration)
	defer cancel()

	body, err := s.getPlexURL(ctx, s.URL+"/:/scrobble", params)
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(body))
	}

	return nil
}

// MarkPlayed marks a video as played.
func (s *Server) MarkPlayed(key string) error {
	return s.MarkPlayedWithContext(context.Background(), key)
}

// EmptyTrashWithContext deletes (a section's) trash.
func (s *Server) EmptyTrashWithContext(ctx context.Context, sectionKey string) error {
	if !s.Configured() {
		return ErrNoURLToken
	}

	params := make(url.Values)
	params.Add("key", sectionKey)

	ctx, cancel := context.WithTimeout(ctx, s.Timeout.Duration)
	defer cancel()

	body, err := s.getPlexURL(ctx, s.URL+"/library/sections/"+sectionKey+"/emptyTrash", params)
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(body))
	}

	return nil
}

// EmptyTrash deletes (a section's) trash.
func (s *Server) EmptyTrash(sectionKey string) error {
	return s.EmptyTrashWithContext(context.Background(), sectionKey)
}

// EmptyTrash deletes (a section's) trash.
func (s *Server) EmptyAllTrashWithContext(ctx context.Context) error {
	if !s.Configured() {
		return ErrNoURLToken
	}

	directory, err := s.GetDirectoryWithContext(ctx)
	if err != nil {
		return err
	}

	for _, library := range directory.Directory {
		if err := s.EmptyTrashWithContext(ctx, library.Key); err != nil {
			return fmt.Errorf("emptying library '%s' trash: %w", library.Title, err)
		}
	}

	return nil
}
