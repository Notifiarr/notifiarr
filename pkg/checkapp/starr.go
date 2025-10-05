package checkapp

import (
	"context"
	"net/http"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"golift.io/starr/lidarr"
	"golift.io/starr/prowlarr"
	"golift.io/starr/radarr"
	"golift.io/starr/readarr"
	"golift.io/starr/sonarr"
)

func Lidarr(ctx context.Context, config apps.StarrConfig) (string, int) {
	status, err := lidarr.New(&config.Config).GetSystemStatusContext(ctx)
	if err != nil {
		return connecting + err.Error(), http.StatusFailedDependency
	}

	return success + status.Version, http.StatusOK
}

func Prowlarr(ctx context.Context, config apps.StarrConfig) (string, int) {
	status, err := prowlarr.New(&config.Config).GetSystemStatusContext(ctx)
	if err != nil {
		return connecting + err.Error(), http.StatusFailedDependency
	}

	return success + status.Version, http.StatusOK
}

func Radarr(ctx context.Context, config apps.StarrConfig) (string, int) {
	status, err := radarr.New(&config.Config).GetSystemStatusContext(ctx)
	if err != nil {
		return connecting + err.Error(), http.StatusFailedDependency
	}

	return success + status.Version, http.StatusOK
}

func Readarr(ctx context.Context, config apps.StarrConfig) (string, int) {
	status, err := readarr.New(&config.Config).GetSystemStatusContext(ctx)
	if err != nil {
		return connecting + err.Error(), http.StatusFailedDependency
	}

	return success + status.Version, http.StatusOK
}

func Sonarr(ctx context.Context, config apps.StarrConfig) (string, int) {
	status, err := sonarr.New(&config.Config).GetSystemStatusContext(ctx)
	if err != nil {
		return connecting + err.Error(), http.StatusFailedDependency
	}

	return success + status.Version, http.StatusOK
}
