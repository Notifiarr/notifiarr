package triggers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/logs/share"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/unmonitor"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"github.com/gorilla/mux"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golift.io/starr"
)

// APIHandler is passed into the webserver so triggers can be executed from the API.
func (a *Actions) APIHandler(req *http.Request) (int, any) {
	return a.handleTrigger(req, website.EventAPI)
}

// Handler handles GUI (non-API) trigger requests.
//
//	@Summary		Trigger action
//	@Description	Executes a specific trigger action.
//	@Tags			System
//	@Produce		text/plain
//	@Param			trigger	path		string	true	"Trigger name to execute"
//	@Param			content	path		string	false	"Optional content for the trigger"
//	@Success		200		{string}	string	"trigger result"
//	@Failure		400		{string}	string	"unknown trigger"
//	@Router			/trigger/{trigger} [get]
//	@Router			/trigger/{trigger}/{content} [get]
func (a *Actions) Handler(response http.ResponseWriter, req *http.Request) {
	code, data := a.handleTrigger(req, website.EventGUI)
	http.Error(response, data, code)
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
	// Triggers and actions internal to the client.
	Triggers []common.TriggerInfo `json:"triggers"`
	// Timers provided by the website to offload scheduling.
	Timers []*timer `json:"timers"`
}

// HandleGetTriggers handles the GET request to get the list of triggers and website timers.
// @Description	Returns a list of triggers and website timers with their intervals, if configured.
// @Summary		Get trigger list
// @Tags			Triggers
// @Produce		json
// @Success		200	{object}	apps.ApiResponse{message=triggers.triggerOutput}	"lists of triggers and website timers"
// @Failure		404	{object}	string												"bad token or api key"
// @Router			/triggers [get]
// @Security		ApiKeyAuth
func (a *Actions) HandleGetTriggers(_ *http.Request) (int, any) {
	triggers, timers, schedules := a.GatherTriggerInfo()
	cronTimers := a.CronTimer.List()
	reply := &triggerOutput{
		Triggers: append(triggers, append(timers, schedules...)...),
		Timers:   make([]*timer, len(cronTimers)),
	}

	for idx, action := range cronTimers {
		reply.Timers[idx] = &timer{
			Name: action.Name,
			Dur:  action.Interval.String(),
			Idx:  idx,
			Path: path.Join(a.Apps.URLBase, fmt.Sprint("api/trigger/custom/", idx)),
		}
	}

	return http.StatusOK, reply
}

// handleTrigger is an abstraction to deal with API or GUI triggers (they have different handlers).
func (a *Actions) handleTrigger(req *http.Request, event website.EventType) (int, string) {
	input := &common.ActionInput{Type: event, ReqID: mnd.GetID(req.Context())}
	trigger := mux.Vars(req)["trigger"]
	content := mux.Vars(req)["content"]

	if content != "" {
		mnd.Log.Printf(input.ReqID, "[%s requested] Incoming Trigger: %s (%s)", event, trigger, content)
	} else {
		mnd.Log.Printf(input.ReqID, "[%s requested] Incoming Trigger: %s", event, trigger)
	}

	_ = req.ParseForm()
	input.Args = req.PostForm["args"]

	return a.runTrigger(req, input, trigger, content)
}

//nolint:cyclop,funlen,gocyclo
func (a *Actions) runTrigger(req *http.Request, input *common.ActionInput, trigger, content string) (int, string) {
	mnd.Log.Trace(input.ReqID, "start: Actions.runTrigger", input.Type, trigger, content != "")
	defer mnd.Log.Trace(input.ReqID, "end: Actions.runTrigger", input.Type, trigger, content != "")

	switch trigger {
	case "custom", "TrigCustomCronTimer":
		return a.customTimer(input, content)
	case "clientlogs":
		return a.clientLogs(content)
	case "command", "TrigCustomCommand":
		return a.command(input, content)
	case "endpoint", "TrigEndpointURL":
		return a.endpoint(input, content)
	case "cfsync", "TrigCFSyncRadarr":
		return a.cfsync(input, content)
	case "rpsync", "TrigCFSyncSonarr":
		return a.rpsync(input, content)
	case "TrigCFSyncLidarr":
		return http.StatusNotImplemented, "Lidarr sync is not implemented."
	case "services":
		return a.services(input)
	case "sessions", "TrigPlexSessions":
		return a.sessions(input)
	case "stuckitems", "TrigStuckItems":
		return a.stuckitems(input)
	case "dashboard", "TrigDashboard":
		return a.dashboard(input)
	case "snapshot", "TrigSnapshot":
		return a.snapshot(input)
	case "gaps", "TrigCollectionGaps":
		return a.gaps(input)
	case "corrupt":
		return a.corrupt(input, content)
	case "TrigProwlarrCorrupt":
		return a.corrupt(input, starr.Prowlarr.String())
	case "TrigLidarrCorrupt":
		return a.corrupt(input, starr.Lidarr.String())
	case "TrigRadarrCorrupt":
		return a.corrupt(input, starr.Radarr.String())
	case "TrigReadarrCorrupt":
		return a.corrupt(input, starr.Readarr.String())
	case "TrigSonarrCorrupt":
		return a.corrupt(input, starr.Sonarr.String())
	case "backup":
		return a.backup(input, content)
	case "TrigLidarrBackup":
		return a.backup(input, starr.Lidarr.String())
	case "TrigRadarrBackup":
		return a.backup(input, starr.Radarr.String())
	case "TrigReadarrBackup":
		return a.backup(input, starr.Readarr.String())
	case "TrigSonarrBackup":
		return a.backup(input, starr.Sonarr.String())
	case "TrigProwlarrBackup":
		return a.backup(input, starr.Prowlarr.String())
	case "reload", "TrigStop":
		return a.handleConfigReload()
	case "notification":
		return a.notification(req.Context(), content)
	case "emptyplextrash", "TrigPlexEmptyTrash":
		return a.emptyplextrash(input, content)
	case "mdblist", "TrigMDBListSync":
		return a.mdblist(input)
	case "uploadlog", "TrigUploadFile":
		return a.uploadlog(input, content)
	case "reconfig", "TrigReconfig":
		return a.reconfig(req, input)
	case "unmonitor":
		return a.unmonitor(req, input, content)
	case "delete":
		return a.delete(req, input, content)
	default:
		return http.StatusBadRequest, "Unknown trigger provided:'" + trigger + "'"
	}
}

// @Description	Trigger a custom website timer. This sends a GET request to trigger an action on the website.
// @Summary		Trigger custom timer
// @Tags			Triggers
// @Produce		json
// @Param			idx	path		int									true	"ID of the custom website timer to trigger"
// @Success		200	{object}	apps.ApiResponse{message=string}	"success: name of timer"
// @Failure		400	{object}	string								"invalid timer ID"
// @Failure		404	{object}	string								"bad token or api key"
// @Router			/trigger/custom/{idx} [get]
// @Security		ApiKeyAuth
func (a *Actions) customTimer(input *common.ActionInput, content string) (int, string) {
	timerList := a.CronTimer.List()

	customTimerID, err := strconv.Atoi(content)
	if err != nil || customTimerID < 0 || customTimerID >= len(timerList) {
		return http.StatusBadRequest, "invalid timer ID"
	}

	defer timerList[customTimerID].Run(input)

	return http.StatusOK, "Custom Website Timer Triggered: " + timerList[customTimerID].Name
}

// @Description	Toggle client error log sharing.
// @Description	This allows enabling and disabling of client error logs being shared with the website.
// @Summary		Toggle client error log sharing
// @Tags			Triggers
// @Produce		json
// @Param			enabled	path		bool								true	"Enable or disable client error log sharing."
// @Success		200		{object}	apps.ApiResponse{message=string}	"success"
// @Failure		404		{object}	string								"bad token or api key"
// @Router			/trigger/clientlogs/{enabled} [get]
// @Security		ApiKeyAuth
func (a *Actions) clientLogs(content string) (int, string) { //nolint:unparam
	if content == "true" || content == "on" || content == "enable" {
		share.Enable()
		return http.StatusOK, "Client log notifications enabled."
	}

	share.Disable()

	return http.StatusOK, "Client log notifications disabled."
}

// @Description	Execute a pre-programmed command with arguments.
// @Summary		Execute Command w/ args
// @Tags			Triggers
// @Produce		json
// @Param			hash	path		bool		true	"Unique hash for command being executed"
// @Param			args	formData	[]string	true	"provide args as multiple 'args' parameters in POST body"	collectionFormat(multi)	example(args=/tmp&args=/var)
// @Accept			application/x-www-form-urlencoded
// @Success		200	{object}	apps.ApiResponse{message=string}	"success"
// @Failure		400	{object}	apps.ApiResponse{message=string}	"bad or missing hash"
// @Failure		404	{object}	string								"bad token or api key"
// @Router			/trigger/command/{hash} [post]
// @Security		ApiKeyAuth
//
//nolint:lll
func _() {}

// @Description	Execute a pre-programmed command.
// @Summary		Execute Command
// @Tags			Triggers
// @Produce		json
// @Param			hash	path		bool								true	"Unique hash for command being executed"
// @Success		200		{object}	apps.ApiResponse{message=string}	"success"
// @Failure		400		{object}	apps.ApiResponse{message=string}	"bad or missing hash"
// @Failure		404		{object}	string								"bad token or api key"
// @Router			/trigger/command/{hash} [get]
// @Security		ApiKeyAuth
func (a *Actions) command(input *common.ActionInput, content string) (int, string) {
	cmd := a.Commands.GetByHash(content)
	if cmd == nil {
		return http.StatusBadRequest, "No command hash provided."
	}

	cmd.Run(input)

	return http.StatusOK, "Command triggered: " + cmd.Name
}

// @Description	Trigger a pre-programmed endpoint URL passthrough request.
// @Summary		Trigger Endpoint
// @Tags			Triggers
// @Produce		json
// @Param			name	path		bool								true	"Name or URL of endpoint being triggered"
// @Success		200		{object}	apps.ApiResponse{message=string}	"success"
// @Failure		400		{object}	apps.ApiResponse{message=string}	"bad or missing name"
// @Failure		404		{object}	string								"bad token or api key"
// @Router			/trigger/endpoint/{name} [get]
// @Security		ApiKeyAuth
func (a *Actions) endpoint(input *common.ActionInput, content string) (int, string) {
	endpoint := a.Endpoints.List().Get(content)
	if endpoint == nil {
		return http.StatusBadRequest, "Endpoint '" + content + "' not found."
	}

	endpoint.Run(input)

	return http.StatusOK, "Endpoint triggered: " + endpoint.Name
}

// @Description	Sync custom profiles and formats to Radarr.
// @Summary		Sync TRaSH Radarr data
// @Tags			Triggers,TRaSH
// @Produce		json
// @Param			instance	path		bool								false	"Triggers sync on this instance if provided, otherwise all instances"
// @Success		200			{object}	apps.ApiResponse{message=string}	"success"
// @Failure		404			{object}	string								"bad token or api key"
// @Router			/trigger/cfsync/{instance} [get]
// @Security		ApiKeyAuth
func (a *Actions) cfsync(input *common.ActionInput, content string) (int, string) {
	if content == "" {
		a.CFSync.SyncRadarrCF(input)
		return http.StatusOK, "Radarr profile and format sync initiated."
	}

	instance, _ := strconv.Atoi(content)
	if err := a.CFSync.SyncRadarrInstanceCF(input, instance); err != nil {
		return http.StatusBadRequest, "Radarr profile and format sync initiated for instance " + content + ": " + err.Error()
	}

	return http.StatusOK, "Radarr profile and format sync initiated for instance " + content
}

// @Description	Sync custom profiles and formats to Sonarr.
// @Summary		Sync TRaSH Sonarr data
// @Tags			Triggers,TRaSH
// @Produce		json
// @Param			instance	path		bool								false	"Triggers sync on this instance if provided, otherwise all instances"
// @Success		200			{object}	apps.ApiResponse{message=string}	"success"
// @Failure		404			{object}	string								"bad token or api key"
// @Router			/trigger/rpsync/{instance} [get]
// @Security		ApiKeyAuth
func (a *Actions) rpsync(input *common.ActionInput, content string) (int, string) {
	if content == "" {
		a.CFSync.SyncSonarrRP(input)
		return http.StatusOK, "Sonarr profile and format sync initiated."
	}

	instance, _ := strconv.Atoi(content)
	if err := a.CFSync.SyncSonarrInstanceRP(input, instance); err != nil {
		return http.StatusBadRequest, "Sonarr profile and format sync initiated for instance " + content + ": " + err.Error()
	}

	return http.StatusOK, "Sonarr profile and format sync initiated for instance " + content
}

// @Description	Reschedule all service checks to run immediately.
// @Summary		Run all service checks
// @Tags			Triggers
// @Produce		json
// @Success		200	{object}	apps.ApiResponse{message=string}	"success"
// @Failure		404	{object}	string								"bad token or api key"
// @Router			/trigger/services [get]
// @Security		ApiKeyAuth
func (a *Actions) services(input *common.ActionInput) (int, string) {
	a.RunChecks(input)
	return http.StatusOK, "All service checks rescheduled for immediate execution."
}

// @Description	Collect Plex sessions and send a notifciation.
// @Summary		Collect Plex Sessions
// @Tags			Triggers
// @Produce		json
// @Success		200	{object}	apps.ApiResponse{message=string}	"success"
// @Failure		501	{object}	apps.ApiResponse{message=string}	"plex is disabled"
// @Failure		404	{object}	string								"bad token or api key"
// @Router			/trigger/sessions [get]
// @Security		ApiKeyAuth
func (a *Actions) sessions(input *common.ActionInput) (int, string) {
	if !a.Apps.Plex.Enabled() {
		return http.StatusNotImplemented, "Plex Sessions are not enabled."
	}

	a.PlexCron.Send(input)

	return http.StatusOK, "Plex sessions triggered."
}

// @Description	Sends cached stuck items notification.
// @Summary		Send a stuck items notification
// @Tags			Triggers
// @Produce		json
// @Success		200	{object}	apps.ApiResponse{message=string}	"success"
// @Failure		404	{object}	string								"bad token or api key"
// @Router			/trigger/stuckitems [get]
// @Security		ApiKeyAuth
func (a *Actions) stuckitems(input *common.ActionInput) (int, string) {
	a.StarrQueue.StuckItems(input)
	return http.StatusOK, "Stuck Queue Items triggered."
}

// @Description	Collects dashboard data and sends a notification.
// @Summary		Send a dashboard notification
// @Tags			Triggers
// @Produce		json
// @Success		200	{object}	apps.ApiResponse{message=string}	"success"
// @Failure		404	{object}	string								"bad token or api key"
// @Router			/trigger/dashboard [get]
// @Security		ApiKeyAuth
func (a *Actions) dashboard(input *common.ActionInput) (int, string) {
	a.Dashboard.Send(input)
	return http.StatusOK, "Dashboard states triggered."
}

// @Description	Collects system snapshot data and sends a notification.
// @Summary		Send a system snapshot notification
// @Tags			Triggers
// @Produce		json
// @Success		200	{object}	apps.ApiResponse{message=string}	"success"
// @Failure		404	{object}	string								"bad token or api key"
// @Router			/trigger/snapshot [get]
// @Security		ApiKeyAuth
func (a *Actions) snapshot(input *common.ActionInput) (int, string) {
	a.SnapCron.Send(input)
	return http.StatusOK, "System Snapshot triggered."
}

// @Description	Send Radarr Library Collection Gaps notification.
// @Summary		Send Collections Gaps Notification
// @Tags			Triggers
// @Produce		json
// @Success		200	{object}	apps.ApiResponse{message=string}	"success"
// @Failure		404	{object}	string								"bad token or api key"
// @Router			/trigger/gaps [get]
// @Security		ApiKeyAuth
func (a *Actions) gaps(input *common.ActionInput) (int, string) {
	a.Gaps.Send(input)
	return http.StatusOK, "Radarr Collections Gaps initiated."
}

// @Description	Start corruption check on all application backups of a specific type.
// @Summary		Start app-specific corruption check
// @Tags			Triggers
// @Produce		json
// @Param			app	path		string								true	"app type to check"	Enum(lidarr, prowlarr, radarr, readarr, sonarr)
// @Success		200	{object}	apps.ApiResponse{message=string}	"success"
// @Failure		400	{object}	apps.ApiResponse{message=string}	"missing app"
// @Failure		404	{object}	string								"bad token or api key"
// @Router			/trigger/corrupt/{app} [get]
// @Security		ApiKeyAuth
func (a *Actions) corrupt(input *common.ActionInput, content string) (int, string) {
	title := cases.Title(language.AmericanEnglish)

	err := a.Backups.Corruption(input, starr.App(title.String(content)))
	if err != nil {
		return http.StatusBadRequest, "Corruption trigger failed: " + err.Error()
	}

	return http.StatusOK, title.String(content) + " corruption checks initiated."
}

// @Description	Start backup file check on all applications of a specific type.
// @Summary		Start app-specific backup check
// @Tags			Triggers
// @Produce		json
// @Param			app	path		string								true	"app type to check"	Enum(lidarr, prowlarr, radarr, readarr, sonarr)
// @Success		200	{object}	apps.ApiResponse{message=string}	"success"
// @Failure		400	{object}	apps.ApiResponse{message=string}	"missing app"
// @Failure		404	{object}	string								"bad token or api key"
// @Router			/trigger/backup/{app} [get]
// @Security		ApiKeyAuth
func (a *Actions) backup(input *common.ActionInput, content string) (int, string) {
	title := cases.Title(language.AmericanEnglish)

	err := a.Backups.Backup(input, starr.App(title.String(content)))
	if err != nil {
		return http.StatusBadRequest, "Backup trigger failed: " + err.Error()
	}

	return http.StatusOK, title.String(content) + " backups check initiated."
}

// @Description	Reload this application's configuration immediately. Reload shuts down everything re-reads the config file and starts back up.
// @Summary		Reload Application
// @Tags			Triggers
// @Produce		json
// @Success		200	{object}	apps.ApiResponse{message=string}	"success"
// @Failure		404	{object}	string								"bad token or api key"
// @Router			/trigger/reload [get]
// @Security		ApiKeyAuth
//
//nolint:lll
func (a *Actions) handleConfigReload() (int, string) {
	go func() {
		// Until we have a way to reliably finish the tunnel requests, this is the best I got.
		time.Sleep(200 * time.Millisecond) //nolint:mnd
		a.ReloadApp("HTTP Triggered Reload")
	}()

	return http.StatusOK, "Application reload initiated."
}

// @Description	Write log entry, and send GUI notification if client has GUI enabled (mac/windows only).
// @Summary		Send Client User Notification
// @Tags			Triggers
// @Produce		json
// @Param			content	path		string								true	"Data for the notification."
// @Success		200		{object}	apps.ApiResponse{message=string}	"success"
// @Failure		400		{object}	apps.ApiResponse{message=string}	"no content"
// @Failure		404		{object}	string								"bad token or api key"
// @Router			/trigger/notification/{content} [get]
// @Security		ApiKeyAuth
func (a *Actions) notification(ctx context.Context, content string) (int, string) {
	if content != "" {
		ui.Toast(ctx, "Notification: %s", content) //nolint:errcheck
		mnd.Log.Printf(mnd.GetID(ctx), "NOTIFICATION: %s", content)

		return http.StatusOK, "Local Nntification sent."
	}

	return http.StatusBadRequest, "Missing notification content."
}

// @Description	Empties one or more Plex library trash cans.
// @Summary		Empty Plex Trashes
// @Tags			Triggers,Plex
// @Produce		json
// @Param			libraryKeys	path		[]string							true	"List of library keys, comma separated."
// @Success		200			{object}	apps.ApiResponse{message=string}	"started"
// @Failure		501			{object}	apps.ApiResponse{message=string}	"plex not enabled"
// @Failure		404			{object}	string								"bad token or api key"
// @Router			/trigger/emptyplextrash/{libraryKeys} [get]
// @Security		ApiKeyAuth
func (a *Actions) emptyplextrash(input *common.ActionInput, content string) (int, string) {
	if !a.Apps.Plex.Enabled() {
		return http.StatusNotImplemented, "Plex is not enabled."
	}

	a.EmptyTrash.Plex(input, strings.Split(content, ","))

	return http.StatusOK, "Emptying Plex Trash for library " + content
}

// @Description	Sends Radarr and Sonarr Libraries for MDBList Syncing.
// @Summary		Send Libraries for MDBList
// @Tags			Triggers
// @Produce		json
// @Success		200	{object}	apps.ApiResponse{message=string}	"success"
// @Failure		404	{object}	string								"bad token or api key"
// @Router			/trigger/mdblist [get]
// @Security		ApiKeyAuth
func (a *Actions) mdblist(input *common.ActionInput) (int, string) {
	a.MDbList.Send(input)
	return http.StatusOK, "MDBList library update started."
}

// @Description	Uploads a log file to Notifiarr.com.
// @Summary		Upload log file to Notifiarr.com
// @Tags			Triggers
// @Produce		json
// @Param			file	path		string								true	"File to upload. Must be one of app, http, debug"
// @Success		200		{object}	apps.ApiResponse{message=string}	"success"
// @Failure		400		{object}	apps.ApiResponse{message=string}	"bad or missing file"
// @Failure		424		{object}	apps.ApiResponse{message=string}	"log uploads disabled"
// @Failure		404		{object}	string								"bad token or api key"
// @Router			/trigger/uploadlog/{file} [get]
// @Security		ApiKeyAuth
func (a *Actions) uploadlog(input *common.ActionInput, file string) (int, string) {
	if logs.Log.NoUploads() {
		return http.StatusFailedDependency, "Uploads Administratively Disabled"
	}

	err := a.FileUpload.Log(input, file)
	if err != nil {
		return http.StatusBadRequest, fmt.Sprintf("Uploading %s log file: %v", file, err)
	}

	return http.StatusOK, fmt.Sprintf("Uploading %s log file.", file)
}

// @Description	Reconfigures actions based on website settings.
// @Summary		Reconfigure Actions
// @Tags			Triggers
// @Produce		json
// @Success		200		{object}	apps.ApiResponse{message=string}	"success"
// @Failure		404		{object}	string								"bad token or api key"
// @Router			/trigger/reconfig [get]
// @Security		ApiKeyAuth
func (a *Actions) reconfig(req *http.Request, input *common.ActionInput) (int, string) {
	var actions *clientinfo.Actions

	if req.Method == http.MethodPost {
		actions = new(clientinfo.Actions)
		if err := json.NewDecoder(req.Body).Decode(actions); err != nil {
			return http.StatusBadRequest, "decoding client info actions: " + err.Error()
		}
	}

	mnd.Log.Printf(mnd.GetID(req.Context()), "[%s requested] Reconfiguring Actions from website settings.", input.Type)
	a.inCh <- inChData{EventType: input.Type, Actions: actions}

	return http.StatusOK, <-a.outCh
}

type unmonitorData struct {
	// Instance number in the app. 1,2,3
	Instances []int `json:"instances"`
	// TMDb ID for the movie, when interacting Radarr.
	TmdbID int64 `json:"tmdbid"`
	// TVDB ID for the series when interacting Sonarr.
	TvdbID int64 `json:"tvdbid"`
	// Season number for the episode when interacting Sonarr.
	Season int `json:"season"`
	// Episode number for the episode when interacting Sonarr.
	Episode int `json:"episode"`
}

// @Description	Unmonitors content in Sonarr or Radarr.
// @Summary		Unmonitor content in Sonarr or Radarr
// @Tags			Triggers
// @Produce		json
// @Param			app	path		string								true	"app type to unmonitor"	Enum(sonarr, radarr)
// @Param			data	body		unmonitorData						true	"Data for the unmonitor request"
// @Success		200		{object}	apps.ApiResponse{message=string}	"success"
// @Failure		400		{object}	apps.ApiResponse{message=string}	"bad json input"
// @Failure		404		{object}	string								"bad token or api key"
// @Router			/trigger/unmonitor/{app} [post]
// @Security		ApiKeyAuth
func (a *Actions) unmonitor(req *http.Request, input *common.ActionInput, app string) (int, string) {
	var data unmonitorData
	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		return http.StatusBadRequest, "decoding unmonitor data: " + err.Error()
	}

	a.Unmonitor.Now(input, &unmonitor.UnmonitorData{
		Action:    "unmonitor",
		App:       app,
		Instances: data.Instances,
		TvdbID:    data.TvdbID,
		TmdbID:    data.TmdbID,
		Season:    data.Season,
		Episode:   data.Episode,
	})

	return http.StatusOK, fmt.Sprintf("Unmonitoring from %d %s instances. Request ID: %s.",
		len(data.Instances), app, input.ReqID)
}

// @Description	Deletes content from Sonarr or Radarr.
// @Summary		Delete content from Sonarr or Radarr
// @Tags			Triggers
// @Produce		json
// @Param			app	path		string								true	"app type to delete"	Enum(sonarr, radarr)
// @Param			data	body		unmonitorData						true	"Data for the delete request"
// @Success		200		{object}	apps.ApiResponse{message=string}	"success"
// @Failure		400		{object}	apps.ApiResponse{message=string}	"bad json input"
// @Failure		404		{object}	string								"bad token or api key"
// @Router			/trigger/delete/{app} [post]
// @Security		ApiKeyAuth
func (a *Actions) delete(req *http.Request, input *common.ActionInput, app string) (int, string) {
	var data unmonitorData
	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		return http.StatusBadRequest, "decoding unmonitor data: " + err.Error()
	}

	a.Unmonitor.Now(input, &unmonitor.UnmonitorData{
		Action:    "delete",
		App:       app,
		Instances: data.Instances,
		TvdbID:    data.TvdbID,
		TmdbID:    data.TmdbID,
		Season:    data.Season,
		Episode:   data.Episode,
	})

	return http.StatusOK, fmt.Sprintf("Deleting from %d %s instances. Request ID: %s.",
		len(data.Instances), app, input.ReqID)
}
