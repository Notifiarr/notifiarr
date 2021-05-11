package plex

import (
	"fmt"
	"net/http"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/gorilla/mux"
)

/* These handlers satisfy apps.APIHandler, sorry. */

const Plex apps.App = "plex"

func (s *Server) HandleSessions(r *http.Request) (int, interface{}) {
	plexID, _ := r.Context().Value(Plex).(int)

	sessions, err := s.GetSessionsWithContext(r.Context())
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to get sessions (%d): %w", plexID, err)
	}

	return http.StatusOK, sessions
}

func (s *Server) HandleKillSession(r *http.Request) (int, interface{}) {
	var (
		ctx       = r.Context()
		plexID, _ = ctx.Value(Plex).(int)
		sessionID = mux.Vars(r)["sessionID"]
		reason    = mux.Vars(r)["reason"]
	)

	err := s.KillSessionWithContext(ctx, sessionID, reason)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to kill session (%s@%d): %w", sessionID, plexID, err)
	}

	return http.StatusOK, "kilt"
}
