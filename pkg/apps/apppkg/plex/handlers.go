package plex

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"golift.io/starr"
)

// HandleSessions provides a web handler to the notifiarr client that returns
// the current Plex sessions. The handler satisfies apps.APIHandler, sorry.
// @Description  Returns Plex sessions.
// @Summary      Retrieve Plex sessions.
// @Tags         plex
// @Produce      json
// @Success      200  {object} Sessions "contains app info included appStatus"
// @Failure      500  {object} string "Plex error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/plex/sessions [get]
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
// @Description  Kill a Plex session.
// @Summary      Kill a Plex session.
// @Tags         plex
// @Produce      text/plain
// @Param        sessionId  query   string  true  "Plex session ID"
// @Param        reason     query   string  true  "Reason the session is being terminated. Sent to the user."
// @Success      200  {object} string "success"
// @Failure      500  {object} string "Plex error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/plex/kill [get]
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

// @Description  Returns the Plex Library Directory.
// @Summary      Retrieve the Plex Library Directory.
// @Tags         plex
// @Produce      json
// @Success      200  {object} SectionDirectory "Plex Library Directory"
// @Failure      500  {object} string "Plex error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/plex/directory [get]
func (s *Server) HandleDirectory(r *http.Request) (int, interface{}) {
	plexID, _ := r.Context().Value(starr.Plex).(int)

	directory, err := s.GetDirectoryWithContext(r.Context())
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("directory request failed (%d): %w", plexID, err)
	}

	return http.StatusOK, directory
}

// @Description  Returns the Plex Library Directory.
// @Summary      Retrieve the Plex Library Directory.
// @Tags         plex
// @Produce      text/plain
// @Param        libraryKey   path    string true  "Plex Library Section Key"
// @Success      200  {object} string "ok"
// @Failure      500  {object} string "Plex error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/plex/emptytrash/{libraryKey} [get]
func (s *Server) HandleEmptyTrash(r *http.Request) (int, interface{}) {
	plexID, _ := r.Context().Value(starr.Plex).(int)

	body, err := s.EmptyTrashWithContext(r.Context(), mux.Vars(r)["key"])
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("trash empty failed (%d): %w", plexID, err)
	}

	return http.StatusOK, "ok: " + string(body)
}

// @Description  Mark a movie or show as watched.
// @Summary      Marks a Plex item as watched.
// @Tags         plex
// @Produce      text/plain
// @Param        itemKey  path    string true  "Plex Item Key"
// @Success      200  {object} string "ok"
// @Failure      500  {object} string "Plex error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/plex/markwatched/{itemKey} [get]
func (s *Server) HandleMarkWatched(r *http.Request) (int, interface{}) {
	plexID, _ := r.Context().Value(starr.Plex).(int)

	body, err := s.MarkPlayedWithContext(r.Context(), mux.Vars(r)["key"])
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("mark watch failed (%d): %w", plexID, err)
	}

	return http.StatusOK, "ok: " + string(body)
}
