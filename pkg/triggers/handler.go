package triggers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/logs/share"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/gorilla/mux"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golift.io/starr"
)

// APIHandler is passed into the webserver so triggers can be executed from the API.
func (a *Actions) APIHandler(req *http.Request) (int, interface{}) {
	return a.handleTrigger(req, website.EventAPI)
}

// Handler handles GUI (non-API) trigger requests.
func (a *Actions) Handler(response http.ResponseWriter, req *http.Request) {
	code, data := a.handleTrigger(req, website.EventGUI)
	http.Error(response, data, code)
}

// handleTrigger is an abstraction to deal with API or GUI triggers (they have different handlers).
func (a *Actions) handleTrigger(req *http.Request, event website.EventType) (int, string) {
	input := &common.ActionInput{Type: website.EventAPI}
	trigger := mux.Vars(req)["trigger"]
	content := mux.Vars(req)["content"]

	if content != "" {
		a.Timers.Debugf("[%s requested] Incoming Trigger: %s (%s)", event, trigger, content)
	} else {
		a.Timers.Debugf("[%s requested] Incoming Trigger: %s", event, trigger)
	}

	_ = req.ParseForm()
	input.Args = req.PostForm["args"]

	return a.runTrigger(input, trigger, content)
}

func (a *Actions) runTrigger(input *common.ActionInput, trigger, content string) (int, string) { //nolint:cyclop
	switch trigger {
	case "clientlogs":
		return a.clientLogs(content)
	case "command":
		return a.command(input, content)
	case "cfsync":
		return a.cfsync(input, content)
	case "rpsync":
		return a.rpsync(input, content)
	case "services":
		return a.services(input)
	case "sessions":
		return a.sessions(input)
	case "stuckitems":
		return a.stuckitems(input)
	case "dashboard":
		return a.dashboard(input)
	case "snapshot":
		return a.snapshot(input)
	case "gaps":
		return a.gaps(input)
	case "corrupt":
		return a.corrupt(input, content)
	case "backup":
		return a.backup(input, content)
	case "reload":
		return a.handleConfigReload()
	case "notification":
		return a.notification(content)
	case "emptyplextrash":
		return a.emptyplextrash(input, content)
	default:
		return http.StatusBadRequest, "Unknown trigger provided:'" + trigger + "'"
	}
}

// @Description  Toggle client error log sharing.
// @Description  This allows enabling and disabling of client error logs being shared with the website.
// @Summary      Toggle client error log sharing
// @Tags         Triggers
// @Produce      json
// @Param        enabled  path   bool  true  "Enable or disable client error log sharing."
// @Success      200  {object} apps.Respond.apiResponse{message=string} "success"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trigger/clientlogs/{enabled} [get]
// @Security     ApiKeyAuth
func (a *Actions) clientLogs(content string) (int, string) {
	if content == "true" || content == "on" || content == "enable" {
		share.Setup(a.Timers.Server)
		return http.StatusOK, "Client log notifications enabled."
	}

	share.StopLogs()

	return http.StatusOK, "Client log notifications disabled."
}

// @Description  Execute a pre-programmed command with arguments.
// @Summary      Execute Command w/ args
// @Tags         Triggers
// @Produce      json
// @Param        hash  path   bool  true  "Unique hash for command being executed"
// @Param        args formData []string true "provide args as multiple 'args' paramers in POST body" collectionFormat(multi) example(args=/tmp&args=/var)
// @Accept       application/x-www-form-urlencoded
// @Success      200  {object} apps.Respond.apiResponse{message=string} "success"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "bad or missing hash"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trigger/command/{hash} [post]
// @Security     ApiKeyAuth
//
//nolint:lll
func _() {}

// @Description  Execute a pre-programmed command.
// @Summary      Execute Command
// @Tags         Triggers
// @Produce      json
// @Param        hash  path   bool  true  "Unique hash for command being executed"
// @Success      200  {object} apps.Respond.apiResponse{message=string} "success"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "bad or missing hash"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trigger/command/{hash} [get]
// @Security     ApiKeyAuth
func (a *Actions) command(input *common.ActionInput, content string) (int, string) {
	cmd := a.Commands.GetByHash(content)
	if cmd == nil {
		return http.StatusBadRequest, "No command hash provided."
	}

	cmd.Run(input)

	return http.StatusOK, "Command triggered: " + cmd.Name
}

// @Description  Sync TRaSH Radarr data.
// @Summary      Sync TRaSH Radarr data
// @Tags         Triggers,TRaSH
// @Produce      json
// @Param        instance  path   bool  false  "Triggers sync on this instance if provided, otherwise all instances"
// @Success      200  {object} apps.Respond.apiResponse{message=string} "success"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trigger/cfsync/{instance} [get]
// @Security     ApiKeyAuth
func (a *Actions) cfsync(input *common.ActionInput, content string) (int, string) {
	if content == "" {
		a.CFSync.SyncRadarrCF(input.Type)
		return http.StatusOK, "TRaSH Custom Formats Radarr Sync initiated."
	}

	instance, _ := strconv.Atoi(content)
	if err := a.CFSync.SyncRadarrInstanceCF(input.Type, instance); err != nil {
		return http.StatusBadRequest, "TRaSH Custom Formats Radarr Sync failed for instance " + content + ": " + err.Error()
	}

	return http.StatusOK, "TRaSH Custom Formats Radarr Sync initiated for instance " + content
}

// @Description  Sync TRaSH Sonarr data.
// @Summary      Sync TRaSH Sonarr data
// @Tags         Triggers,TRaSH
// @Produce      json
// @Param        instance  path   bool  false  "Triggers sync on this instance if provided, otherwise all instances"
// @Success      200  {object} apps.Respond.apiResponse{message=string} "success"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trigger/rpsync/{instance} [get]
// @Security     ApiKeyAuth
func (a *Actions) rpsync(input *common.ActionInput, content string) (int, string) {
	if content == "" {
		a.CFSync.SyncSonarrRP(input.Type)
		return http.StatusOK, "TRaSH Release Profile Sonarr Sync initiated."
	}

	instance, _ := strconv.Atoi(content)
	if err := a.CFSync.SyncSonarrInstanceRP(input.Type, instance); err != nil {
		return http.StatusBadRequest, "TRaSH Release Profile Sonarr Sync failed for instance " + content + ": " + err.Error()
	}

	return http.StatusOK, "TRaSH Release Profile Sonarr Sync initiated for instance " + content
}

// @Description  Reschedule all service checks to run immediately.
// @Summary      Run all service checks
// @Tags         Triggers
// @Produce      json
// @Success      200  {object} apps.Respond.apiResponse{message=string} "success"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trigger/services [get]
// @Security     ApiKeyAuth
func (a *Actions) services(input *common.ActionInput) (int, string) {
	a.Timers.RunChecks(input.Type)
	return http.StatusOK, "All service checks rescheduled for immediate execution."
}

// @Description  Collect Plex sessions and send a notifciation.
// @Summary      Collect Plex Sessions
// @Tags         Triggers
// @Produce      json
// @Success      200  {object} apps.Respond.apiResponse{message=string} "success"
// @Failure      501  {object} apps.Respond.apiResponse{message=string} "plex is disabled"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trigger/sessions [get]
// @Security     ApiKeyAuth
func (a *Actions) sessions(input *common.ActionInput) (int, string) {
	if !a.Timers.Apps.Plex.Enabled() {
		return http.StatusNotImplemented, "Plex Sessions are not enabled."
	}

	a.PlexCron.Send(input.Type)

	return http.StatusOK, "Plex sessions triggered."
}

// @Description  Sends cached stuck items notification.
// @Summary      Send a stuck items notification
// @Tags         Triggers
// @Produce      json
// @Success      200  {object} apps.Respond.apiResponse{message=string} "success"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trigger/stuckitems [get]
// @Security     ApiKeyAuth
func (a *Actions) stuckitems(input *common.ActionInput) (int, string) {
	a.StarrQueue.StuckItems(input.Type)
	return http.StatusOK, "Stuck Queue Items triggered."
}

// @Description  Collects dashboard data and sends a notification.
// @Summary      Send a dashboard notification
// @Tags         Triggers
// @Produce      json
// @Success      200  {object} apps.Respond.apiResponse{message=string} "success"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trigger/dashboard [get]
// @Security     ApiKeyAuth
func (a *Actions) dashboard(input *common.ActionInput) (int, string) {
	a.Dashboard.Send(input.Type)
	return http.StatusOK, "Dashboard states triggered."
}

// @Description  Collects system snapshot data and sends a notification.
// @Summary      Send a system snapshot notification
// @Tags         Triggers
// @Produce      json
// @Success      200  {object} apps.Respond.apiResponse{message=string} "success"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trigger/snapshot [get]
// @Security     ApiKeyAuth
func (a *Actions) snapshot(input *common.ActionInput) (int, string) {
	a.SnapCron.Send(input.Type)
	return http.StatusOK, "System Snapshot triggered."
}

// @Description  Send Radarr Library Collection Gaps notification.
// @Summary      Send Collections Gaps Notification
// @Tags         Triggers
// @Produce      json
// @Success      200  {object} apps.Respond.apiResponse{message=string} "success"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trigger/gaps [get]
// @Security     ApiKeyAuth
func (a *Actions) gaps(input *common.ActionInput) (int, string) {
	a.Gaps.Send(input.Type)
	return http.StatusOK, "Radarr Collections Gaps initiated."
}

// @Description  Start corruption check on all application backups of a specific type.
// @Summary      Start app-specific corruption check
// @Tags         Triggers
// @Produce      json
// @Param        app  path   string  true  "app type to check" Enum(lidarr, prowlarr, radarr, readarr, sonarr)
// @Success      200  {object} apps.Respond.apiResponse{message=string} "success"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "missing app"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trigger/corrupt/{app} [get]
// @Security     ApiKeyAuth
func (a *Actions) corrupt(input *common.ActionInput, content string) (int, string) {
	title := cases.Title(language.AmericanEnglish)

	err := a.Backups.Corruption(input, starr.App(title.String(content)))
	if err != nil {
		return http.StatusBadRequest, "Corruption trigger failed: " + err.Error()
	}

	return http.StatusOK, title.String(content) + " corruption checks initiated."
}

// @Description  Start backup file check on all applications of a specific type.
// @Summary      Start app-specific backup check
// @Tags         Triggers
// @Produce      json
// @Param        app  path   string  true  "app type to check" Enum(lidarr, prowlarr, radarr, readarr, sonarr)
// @Success      200  {object} apps.Respond.apiResponse{message=string} "success"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "missing app"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trigger/backup/{app} [get]
// @Security     ApiKeyAuth
func (a *Actions) backup(input *common.ActionInput, content string) (int, string) {
	title := cases.Title(language.AmericanEnglish)

	err := a.Backups.Backup(input, starr.App(title.String(content)))
	if err != nil {
		return http.StatusBadRequest, "Backup trigger failed: " + err.Error()
	}

	return http.StatusOK, title.String(content) + " backups check initiated."
}

// @Description  Reload this application's configuration immediately. Reload shuts down everything re-reads the config file and starts back up.
// @Summary      Reload Application
// @Tags         Triggers
// @Produce      json
// @Success      200  {object} apps.Respond.apiResponse{message=string} "success"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trigger/reload [get]
// @Security     ApiKeyAuth
//
//nolint:lll
func (a *Actions) handleConfigReload() (int, string) {
	defer a.Timers.ReloadApp("HTTP Triggered Reload")
	return http.StatusOK, "Application reload initiated."
}

// @Description  Write log entry, and send GUI notification if client has GUI enabled (mac/windows only).
// @Summary      Send Client User Notification
// @Tags         Triggers
// @Produce      json
// @Param        content  path   string  true  "Data for the notification."
// @Success      200  {object} apps.Respond.apiResponse{message=string} "success"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "no content"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trigger/notification/{content} [get]
// @Security     ApiKeyAuth
func (a *Actions) notification(content string) (int, string) {
	if content != "" {
		ui.Notify("Notification: %s", content) //nolint:errcheck
		a.Timers.Printf("NOTIFICATION: %s", content)

		return http.StatusOK, "Local Nntification sent."
	}

	return http.StatusBadRequest, "Missing notification content."
}

// @Description  Empties one or more Plex library trash cans.
// @Summary      Empty Plex Trashes
// @Tags         Triggers,Plex
// @Produce      json
// @Param        libraryKeys  path   []string  true  "List of library keys, comma separated."
// @Success      200  {object} apps.Respond.apiResponse{message=string} "started"
// @Failure      501  {object} apps.Respond.apiResponse{message=string} "plex not enabled"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trigger/emptyplextrash/{libraryKeys} [get]
// @Security     ApiKeyAuth
func (a *Actions) emptyplextrash(input *common.ActionInput, content string) (int, string) {
	if !a.Timers.Apps.Plex.Enabled() {
		return http.StatusNotImplemented, "Plex is not enabled."
	}

	a.EmptyTrash.Plex(input.Type, strings.Split(content, ","))

	return http.StatusOK, "Emptying Plex Trash for library " + content
}
