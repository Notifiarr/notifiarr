package plex

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"golift.io/starr"
)

// HandleSessions provides a web handler to the notifiarr client that returns
// the current Plex sessions. The handler satisfies apps.APIHandler, sorry.
func (s *Server) HandleSessions(r *http.Request) (int, interface{}) {
	plexID, _ := r.Context().Value(starr.Plex).(int)

	sessions, err := s.GetSessionsWithContext(r.Context())
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to get sessions (%d): %w", plexID, err)
	}

	return http.StatusOK, sessions
}

// HandleKillSession provides a web handler to the notifiarr client allows
// notifiarr.com (via Discord request) to end a Plex session.
// The handler satisfies apps.APIHandler, sorry.
func (s *Server) HandleKillSession(r *http.Request) (int, interface{}) {
	var (
		ctx       = r.Context()
		plexID, _ = ctx.Value(starr.Plex).(int)
		sessionID = mux.Vars(r)["sessionId"]
		reason    = mux.Vars(r)["reason"]
	)

	_, err := s.KillSessionWithContext(ctx, sessionID, reason)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to kill session (%s@%d): %w", sessionID, plexID, err)
	}

	return http.StatusOK, fmt.Sprintf("kilt session '%s' with reason: %s", sessionID, reason)
}

func (s *Server) HandleDirectory(r *http.Request) (int, interface{}) {
	plexID, _ := r.Context().Value(starr.Plex).(int)

	directory, err := s.GetDirectoryWithContext(r.Context())
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("directory request failed (%d): %w", plexID, err)
	}

	return http.StatusOK, directory
}

func (s *Server) HandleEmptyTrash(r *http.Request) (int, interface{}) {
	plexID, _ := r.Context().Value(starr.Plex).(int)

	body, err := s.EmptyTrashWithContext(r.Context(), mux.Vars(r)["key"])
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("trash empty failed (%d): %w", plexID, err)
	}

	return http.StatusOK, "ok: " + string(body)
}

func (s *Server) HandleMarkWatched(r *http.Request) (int, interface{}) {
	plexID, _ := r.Context().Value(starr.Plex).(int)

	body, err := s.MarkPlayedWithContext(r.Context(), mux.Vars(r)["key"])
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("mark watch failed (%d): %w", plexID, err)
	}

	return http.StatusOK, "ok: " + string(body)
}
