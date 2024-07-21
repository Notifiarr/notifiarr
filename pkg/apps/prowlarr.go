package apps

import (
	"encoding/json"
	"fmt"
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

// ProwlarrConfig represents the input data for a Prowlarr server.
type ProwlarrConfig struct {
	ExtraConfig
	*starr.Config
	*prowlarr.Prowlarr `json:"-" toml:"-" xml:"-"`
	errorf             func(string, ...interface{}) `json:"-" toml:"-" xml:"-"`
}

func getProwlarr(r *http.Request) *prowlarr.Prowlarr {
	app, _ := r.Context().Value(starr.Prowlarr).(*ProwlarrConfig)
	return app.Prowlarr
}

// Enabled returns true if the Prowlarr instance is enabled and usable.
func (p *ProwlarrConfig) Enabled() bool {
	return p != nil && p.Config != nil && p.URL != "" && p.APIKey != "" && p.Timeout.Duration >= 0
}

func (a *Apps) setupProwlarr() error {
	for idx, app := range a.Prowlarr {
		if app.Config == nil || app.Config.URL == "" {
			return fmt.Errorf("%w: missing url: Prowlarr config %d", ErrInvalidApp, idx+1)
		} else if !strings.HasPrefix(app.Config.URL, "http://") && !strings.HasPrefix(app.Config.URL, "https://") {
			return fmt.Errorf("%w: URL must begin with http:// or https://: Prowlarr config %d", ErrInvalidApp, idx+1)
		}

		if a.Logger.DebugEnabled() {
			app.Config.Client = starr.ClientWithDebug(app.Timeout.Duration, app.ValidSSL, debuglog.Config{
				MaxBody: a.MaxBody,
				Debugf:  a.Debugf,
				Caller:  metricMakerCallback(string(starr.Prowlarr)),
				Redact:  []string{app.APIKey, app.Password, app.HTTPPass},
			})
		} else {
			app.Config.Client = starr.Client(app.Timeout.Duration, app.ValidSSL)
			app.Config.Client.Transport = NewMetricsRoundTripper(starr.Prowlarr.String(), app.Config.Client.Transport)
		}

		app.errorf = a.Errorf
		app.URL = strings.TrimRight(app.URL, "/")
		app.Prowlarr = prowlarr.New(app.Config)
	}

	return nil
}

// @Description  Returns Prowlarr Notifications with a name that matches 'notifiar'.
// @Summary      Retrieve Prowlarr Notifications
// @Tags         Prowlarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Success      200  {object} apps.Respond.apiResponse{message=[]prowlarr.NotificationOutput} "notifications"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/prowlarr/{instance}/notifications [get]
// @Security     ApiKeyAuth
func prowlarrGetNotifications(req *http.Request) (int, interface{}) {
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

// @Description  Updates a Notification in Prowlarr.
// @Summary      Update Prowlarr Notification
// @Tags         Prowlarr
// @Produce      json
// @Accept       json
// @Param        instance  path   int64  true  "instance ID"
// @Param        PUT body prowlarr.NotificationInput  true  "notification content"
// @Success      200  {object} apps.Respond.apiResponse{message=string} "ok"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "bad json input"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/prowlarr/{instance}/notification [put]
// @Security     ApiKeyAuth
func prowlarrUpdateNotification(req *http.Request) (int, interface{}) {
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

// @Description  Creates a new Prowlarr Notification.
// @Summary      Add Prowlarr Notification
// @Tags         Prowlarr
// @Produce      json
// @Accept       json
// @Param        instance  path   int64  true  "instance ID"
// @Param        POST body prowlarr.NotificationInput true "new item content"
// @Success      200  {object} apps.Respond.apiResponse{message=int64} "new notification ID"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "json input error"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/prowlarr/{instance}/notification [post]
// @Security     ApiKeyAuth
func prowlarrAddNotification(req *http.Request) (int, interface{}) {
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
