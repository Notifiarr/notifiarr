package checkapp

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/apps"
)

func Plex(ctx context.Context, app apps.PlexConfig) (string, int) {
	server := app.Setup(0)

	info, err := server.GetInfo(ctx)
	if err != nil {
		return "Getting Info: " + err.Error(), http.StatusFailedDependency
	}

	return "Plex OK! Version: " + info.Version, http.StatusOK
}

func Tautulli(ctx context.Context, app apps.TautulliConfig) (string, int) {
	tautulli := app.Setup(0)

	if app.APIKey == "" {
		return "Tautulli API Key is not set", http.StatusFailedDependency
	}

	if !strings.HasPrefix(app.URL, "http") {
		return "Tautulli URL is not set, must begin with http(s)://", http.StatusFailedDependency
	}

	users, err := tautulli.GetUsers(ctx)
	if err != nil {
		return "Getting Users: " + err.Error(), http.StatusFailedDependency
	}

	return fmt.Sprintf("Tautulli OK! Users: %d", len(users.Response.Data)), http.StatusOK
}
