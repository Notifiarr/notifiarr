package plex

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/gorilla/mux"
)

// Plex is used a context ID.
const Plex apps.App = "plex"

// HandleSessions provides a web handlersto the notifiarr client that returns
// the current Plex sessions. The handler satisfies apps.APIHandler, sorry.
func (s *Server) HandleSessions(r *http.Request) (int, interface{}) {
	plexID, _ := r.Context().Value(Plex).(int)

	sessions, err := s.GetSessionsWithContext(r.Context())
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to get sessions (%d): %w", plexID, err)
	}

	return http.StatusOK, &Sessions{
		Name:       s.Name,
		AccountMap: strings.Split(s.AccountMap, "|"),
		Sessions:   sessions,
	}
}

// HandleKillSession provides a web handler to the notifiarr client allows
// notifiarr.com (via Discord request) to end a Plex session.
// The handler satisfies apps.APIHandler, sorry.
func (s *Server) HandleKillSession(r *http.Request) (int, interface{}) {
	var (
		ctx       = r.Context()
		plexID, _ = ctx.Value(Plex).(int)
		sessionID = mux.Vars(r)["sessionId"]
		reason    = mux.Vars(r)["reason"]
	)

	err := s.KillSessionWithContext(ctx, sessionID, reason)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to kill session (%s@%d): %w", sessionID, plexID, err)
	}

	return http.StatusOK, "kilt"
}
