package apps

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"golift.io/starr"
	"golift.io/starr/debuglog"
	"golift.io/starr/prowlarr"
)

// prowlarrHandlers is called once on startup to register the web API paths.
func (a *Apps) prowlarrHandlers() {
	a.HandleAPIpath(starr.Prowlarr, "/notification", prowlarrGetNotifications, "GET")
	a.HandleAPIpath(starr.Prowlarr, "/notification", prowlarrUpdateNotification, "PUT")
	a.HandleAPIpath(starr.Prowlarr, "/notification", prowlarrAddNotification, "POST")
}

func getProwlarr(r *http.Request) Prowlarr {
	return r.Context().Value(starr.Prowlarr).(Prowlarr) //nolint:forcetypeassert
}

type Prowlarr struct {
	StarrApp           `json:"-" toml:"-" xml:"-"`
	*prowlarr.Prowlarr `json:"-" toml:"-" xml:"-"`
}

func (a *AppsConfig) setupProwlarr() ([]Prowlarr, error) {
	output := make([]Prowlarr, len(a.Prowlarr))

	for idx := range a.Prowlarr {
		app := &a.Prowlarr[idx]
		if err := checkUrl(app.URL, starr.Prowlarr.String(), idx); err != nil {
			return nil, err
		}

		if mnd.Log.DebugEnabled() {
			app.Config.Client = starr.ClientWithDebug(app.Timeout.Duration, app.ValidSSL, debuglog.Config{
				MaxBody: a.MaxBody,
				Debugf:  mnd.Log.Debugf,
				Caller:  metricMakerCallback(string(starr.Prowlarr)),
				Redact:  []string{app.APIKey, app.Password, app.HTTPPass},
			})
		} else if PoolingEnabled() {
			app.Config.Client = PooledClient(app.Timeout.Duration, app.ValidSSL)
			app.Config.Client.Transport = NewMetricsRoundTripper(starr.Prowlarr.String(), app.Config.Client.Transport)
		} else {
			app.Config.Client = starr.Client(app.Timeout.Duration, app.ValidSSL)
			app.Config.Client.Transport = NewMetricsRoundTripper(starr.Prowlarr.String(), app.Config.Client.Transport)
		}

		app.URL = strings.TrimRight(app.URL, "/")
		output[idx] = Prowlarr{
			StarrApp: StarrApp{
				StarrConfig: a.Prowlarr[idx],
			},
			Prowlarr: prowlarr.New(&app.Config),
		}
	}

	return output, nil
}

// @Description	Returns Prowlarr Notifications with a name that matches 'notifiar'.
// @Summary		Retrieve Prowlarr Notifications
// @Tags			Prowlarr
// @Produce		json
// @Param			instance	path		int64													true	"instance ID"
// @Success		200			{object}	apps.ApiResponse{message=[]prowlarr.NotificationOutput}	"notifications"
// @Failure		503			{object}	apps.ApiResponse{message=string}						"instance error"
// @Failure		404			{object}	string													"bad token or api key"
// @Router			/prowlarr/{instance}/notifications [get]
// @Security		ApiKeyAuth
func prowlarrGetNotifications(req *http.Request) (int, any) {
	notifs, err := getProwlarr(req).GetNotificationsContext(req.Context())
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "getting notifications", err)
	}

	output := []*prowlarr.NotificationOutput{}

	for _, notif := range notifs {
		if strings.Contains(strings.ToLower(notif.Name), "notifiar") {
			output = append(output, notif)
		}
	}

	return http.StatusOK, output
}

// @Description	Updates a Notification in Prowlarr.
// @Summary		Update Prowlarr Notification
// @Tags			Prowlarr
// @Produce		json
// @Accept			json
// @Param			instance	path		int64								true	"instance ID"
// @Param			PUT			body		prowlarr.NotificationInput			true	"notification content"
// @Success		200			{object}	apps.ApiResponse{message=string}	"ok"
// @Failure		400			{object}	apps.ApiResponse{message=string}	"bad json input"
// @Failure		503			{object}	apps.ApiResponse{message=string}	"instance error"
// @Failure		404			{object}	string								"bad token or api key"
// @Router			/prowlarr/{instance}/notification [put]
// @Security		ApiKeyAuth
func prowlarrUpdateNotification(req *http.Request) (int, any) {
	var notif prowlarr.NotificationInput

	err := json.NewDecoder(req.Body).Decode(&notif)
	if err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	_, err = getProwlarr(req).UpdateNotificationContext(req.Context(), &notif)
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "updating notification", err)
	}

	return http.StatusOK, mnd.Success
}

// @Description	Creates a new Prowlarr Notification.
// @Summary		Add Prowlarr Notification
// @Tags			Prowlarr
// @Produce		json
// @Accept			json
// @Param			instance	path		int64								true	"instance ID"
// @Param			POST		body		prowlarr.NotificationInput			true	"new item content"
// @Success		200			{object}	apps.ApiResponse{message=int64}		"new notification ID"
// @Failure		400			{object}	apps.ApiResponse{message=string}	"json input error"
// @Failure		503			{object}	apps.ApiResponse{message=string}	"instance error"
// @Failure		404			{object}	string								"bad token or api key"
// @Router			/prowlarr/{instance}/notification [post]
// @Security		ApiKeyAuth
func prowlarrAddNotification(req *http.Request) (int, any) {
	var notif prowlarr.NotificationInput

	err := json.NewDecoder(req.Body).Decode(&notif)
	if err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	id, err := getProwlarr(req).AddNotificationContext(req.Context(), &notif)
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "adding notification", err)
	}

	return http.StatusOK, id
}
