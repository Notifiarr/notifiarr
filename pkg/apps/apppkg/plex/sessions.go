package plex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// Sessions is the config input data.
type Sessions struct {
	Name     string     `json:"server"`
	HostID   string     `json:"hostId"`
	Sessions []*Session `json:"sessions"`
}

// ErrBadStatus is returned when plex returns an invalid status code.
var ErrBadStatus = errors.New("status code not 200")

// GetSessions returns the Plex sessions in JSON format, no timeout.
func (s *Server) GetSessions() (*Sessions, error) {
	return s.GetSessionsWithContext(context.Background())
}

// GetSessionsWithContext returns the Plex sessions in JSON format.
func (s *Server) GetSessionsWithContext(ctx context.Context) (*Sessions, error) {
	var (
		output struct {
			//nolint:tagliatelle
			MediaContainer struct {
				Sessions []*Session `json:"Metadata"`
			} `json:"MediaContainer"`
		}
		sessions = &Sessions{Name: s.name}
	)

	body, err := s.getPlexURL(ctx, s.config.URL+"/status/sessions", nil)
	if err != nil {
		return sessions, fmt.Errorf("%w: %s", err, string(body))
	}

	if err = json.Unmarshal(body, &output); err != nil {
		return sessions, fmt.Errorf("parsing plex sessions (TRY UPGRADING PLEX): %w: %s", err, string(body))
	}

	sessions.Sessions = output.MediaContainer.Sessions

	return sessions, nil
}

// KillSessionWithContext kills a Plex session.
func (s *Server) KillSessionWithContext(ctx context.Context, sessionID, reason string) ([]byte, error) {
	params := make(url.Values)
	params.Add("sessionId", sessionID)
	params.Add("reason", reason)

	body, err := s.getPlexURL(ctx, s.config.URL+"/status/sessions/terminate", params)
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
func (s *Server) MarkPlayedWithContext(ctx context.Context, key string) ([]byte, error) {
	params := make(url.Values)
	params.Add("identifier", "com.plexapp.plugins.library")
	params.Add("key", key)

	body, err := s.getPlexURL(ctx, s.config.URL+"/:/scrobble", params)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, string(body))
	}

	return body, nil
}

// MarkPlayed marks a video as played.
func (s *Server) MarkPlayed(key string) ([]byte, error) {
	return s.MarkPlayedWithContext(context.Background(), key)
}

// EmptyTrashWithContext deletes (a section's) trash.
func (s *Server) EmptyTrashWithContext(ctx context.Context, libraryKey string) ([]byte, error) {
	// Requires PUT with no data.
	body, err := s.putPlexURL(ctx, s.config.URL+"/library/sections/"+libraryKey+"/emptyTrash", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, string(body))
	}

	return body, nil
}

// EmptyTrash deletes (a section's) trash.
func (s *Server) EmptyTrash(sectionKey string) ([]byte, error) {
	return s.EmptyTrashWithContext(context.Background(), sectionKey)
}

// EmptyAllTrashWithContext deletes the trash in all sections.
func (s *Server) EmptyAllTrashWithContext(ctx context.Context) error {
	directory, err := s.GetDirectoryWithContext(ctx)
	if err != nil {
		return err
	}

	for _, library := range directory.Directory {
		if _, err := s.EmptyTrashWithContext(ctx, library.Key); err != nil {
			return fmt.Errorf("emptying section '%s' trash: %w", library.Title, err)
		}
	}

	return nil
}

// GetMediaTranscode returns the transcode info in a format for an html template to consume.
func GetMediaTranscode(mediaList []*Media) []string {
	if len(mediaList) == 0 || len(mediaList[0].Part) == 0 {
		return []string{"", ""}
	}

	var (
		media    = mediaList[0]
		videoMsg string
		audioMsg string
	)

	for _, stream := range media.Part[0].Stream {
		if stream.StreamType == 1 {
			videoMsg = stream.DisplayTitle
			if stream.Decision == "transcode" {
				videoMsg += fmt.Sprintf(" → %s (%s)", media.VideoResolution, strings.ToUpper(stream.Codec))
			}
		} else if stream.StreamType == 2 { //nolint:mnd
			audioMsg = stream.DisplayTitle
			if stream.Decision == "transcode" {
				audioMsg += " → " + strings.ToUpper(stream.Codec)
			}
		}
	}

	return []string{videoMsg, audioMsg}
}
