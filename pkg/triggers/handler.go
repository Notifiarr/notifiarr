package triggers

import (
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

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

type trigger struct {
	Name string `json:"name"`
	Dur  string `json:"interval,omitempty"`
	Path string `json:"apiPath,omitempty"`
}

type timer struct {
	Name string `json:"name"`
	Dur  string `json:"interval"`
	// Use this ID to trigger this timer with the trigger/custom endpoint.
	Idx int `json:"id"`
	// The client API path to trigger this custom timer.
	Path string `json:"apiPath"`
}

type triggerOutput struct {
	Triggers []*trigger `json:"triggers"`
	Timers   []*timer   `json:"timers"`
}

// @Description  Returns a list of triggers and website timers with their intervals, if configured.
// @Summary      Get trigger list
// @Tags         Triggers
// @Produce      json
// @Success      200  {object} apps.Respond.apiResponse{message=triggers.triggerOutput} "lists of triggers and timers"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/triggers [get]
// @Security     ApiKeyAuth
func (a *Actions) HandleGetTriggers(_ *http.Request) (int, interface{}) {
	triggers, timers := a.Timers.GatherTriggerInfo()
	temp := make(map[string]*trigger) // used to dedup.

	for name, dur := range triggers {
		if dur.Duration == 0 {
			temp[name] = &trigger{Name: name}
		} else {
			temp[name] = &trigger{Name: name, Dur: dur.String()}
		}
	}

	for name, dur := range timers {
		if _, ok := temp[name]; !ok {
			temp[name] = &trigger{Name: name, Dur: dur.String()}
		}
	}

	cronTimers := a.CronTimer.List()
	reply := &triggerOutput{
		Triggers: make([]*trigger, len(temp)),
		Timers:   make([]*timer, len(cronTimers)),
	}

	idx := 0

	for _, t := range temp {
		reply.Triggers[idx] = t
		idx++
	}

	for idx, action := range cronTimers {
		reply.Timers[idx] = &timer{
			Name: action.Name,
			Dur:  action.Interval.String(),
			Idx:  idx,
			Path: path.Join(a.Timers.Apps.URLBase, fmt.Sprint("api/trigger/custom/", idx)),
		}
	}

	return http.StatusOK, reply
}

// handleTrigger is an abstraction to deal with API or GUI triggers (they have different handlers).
func (a *Actions) handleTrigger(req *http.Request, event website.EventType) (int, string) {
	input := &common.ActionInput{Type: event}
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
	case "custom":
		return a.customTimer(input, content)
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
	case "mdblist":
		return a.mdblist(input)
	case "uploadlog":
		return a.uploadlog(input, content)
	default:
		return http.StatusBadRequest, "Unknown trigger provided:'" + trigger + "'"
	}
}

// @Description  Trigger a custom website timer. This sends a GET request to trigger an action on the website.
// @Summary      Trigger custom timer
// @Tags         Triggers
// @Produce      json
// @Param        idx  path   int  true  "ID of the custom website timer to trigger"
// @Success      200  {object} apps.Respond.apiResponse{message=string} "success: name of timer"
// @Failure      400  {object} string "invalid timer ID"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trigger/custom/{idx} [get]
// @Security     ApiKeyAuth
func (a *Actions) customTimer(input *common.ActionInput, content string) (int, string) {
	timerList := a.CronTimer.List()

	customTimerID, err := strconv.Atoi(content)
	if err != nil || customTimerID < 0 || customTimerID >= len(timerList) {
		return http.StatusBadRequest, "invalid timer ID"
	}

	defer timerList[customTimerID].Run(input)

	return http.StatusOK, "Custom Website Timer Triggered: " + timerList[customTimerID].Name
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
	go func() {
		// Until we have a way to reliably finish the tunnel requests, this is the best I got.
		time.Sleep(200 * time.Millisecond) //nolint:gomnd
		a.Timers.ReloadApp("HTTP Triggered Reload")
	}()

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

// @Description  Sends Radarr and Sonarr Libraries for MDBList Syncing.
// @Summary      Send Libraries for MDBList
// @Tags         Triggers
// @Produce      json
// @Success      200  {object} apps.Respond.apiResponse{message=string} "success"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trigger/mdblist [get]
// @Security     ApiKeyAuth
func (a *Actions) mdblist(input *common.ActionInput) (int, string) {
	a.MDbList.Send(input.Type)
	return http.StatusOK, "MDBList library update started."
}

// @Description  Uploads a log file to Notifiarr.com.
// @Summary      Upload log file to Notifiarr.com
// @Tags         Triggers
// @Produce      json
// @Param        file  path   string  true  "File to upload. Must be one of app, http, debug"
// @Success      200  {object} apps.Respond.apiResponse{message=string} "success"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "bad or missing file"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trigger/uploadlog/{file} [get]
// @Security     ApiKeyAuth
func (a *Actions) uploadlog(input *common.ActionInput, file string) (int, string) {
	err := a.FileUpload.Log(input.Type, file)
	if err != nil {
		return http.StatusBadRequest, fmt.Sprintf("Uploading %s log file: %v", file, err)
	}

	return http.StatusOK, fmt.Sprintf("Uploading %s log file.", file)
}
