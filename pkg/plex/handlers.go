package plex

import (
	"fmt"
	"net/http"

	"github.com/Notifiarr/notifiarr/pkg/apps"
)

func (s *Server) HandleSessions(r *http.Request) (int, interface{}) {
	id, _ := r.Context().Value(apps.Generic).(int)

	sessions, err := s.GetSessions()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to get sessions (%d): %w", id, err)
	}

	return http.StatusOK, sessions
}

func (s *Server) HandleKillSession(r *http.Request) (int, interface{}) {
	return http.StatusOK, "kill not working yet"
}
