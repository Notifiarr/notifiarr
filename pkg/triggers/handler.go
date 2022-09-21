package triggers

import (
	"net/http"
	"strconv"

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
		a.timers.Debugf("[%s requested] Incoming Trigger: %s (%s)", event, trigger, content)
	} else {
		a.timers.Debugf("[%s requested] Incoming Trigger: %s", event, input.Type, trigger)
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
	default:
		return http.StatusBadRequest, "Unknown trigger provided:'" + trigger + "'"
	}
}

func (a *Actions) clientLogs(content string) (int, string) {
	if content == "true" || content == "on" || content == "enable" {
		share.Setup(a.timers.Server)
		return http.StatusBadRequest, "Client log notifications enabled."
	}

	share.StopLogs()

	return http.StatusBadRequest, "Client log notifications disabled."
}

func (a *Actions) command(input *common.ActionInput, content string) (int, string) {
	cmd := a.Commands.GetByHash(content)
	if cmd == nil {
		return http.StatusBadRequest, "No command hash provided."
	}

	cmd.Run(input)

	return http.StatusOK, "Command triggered: " + cmd.Name
}

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

func (a *Actions) services(input *common.ActionInput) (int, string) {
	a.timers.RunChecks(input.Type)
	return http.StatusOK, "All service checks rescheduled for immediate exeution."
}

func (a *Actions) sessions(input *common.ActionInput) (int, string) {
	if !a.timers.Apps.Plex.Enabled() {
		return http.StatusNotImplemented, "Plex Sessions are not enabled."
	}

	a.PlexCron.Send(input.Type)

	return http.StatusOK, "Plex sessions triggered."
}

func (a *Actions) stuckitems(input *common.ActionInput) (int, string) {
	a.StarrQueue.StuckItems(input.Type)
	return http.StatusOK, "Stuck Queue Items triggered."
}

func (a *Actions) dashboard(input *common.ActionInput) (int, string) {
	a.Dashboard.Send(input.Type)
	return http.StatusOK, "Dashboard states triggered."
}

func (a *Actions) snapshot(input *common.ActionInput) (int, string) {
	a.SnapCron.Send(input.Type)
	return http.StatusOK, "System Snapshot triggered."
}

func (a *Actions) gaps(input *common.ActionInput) (int, string) {
	a.Gaps.Send(input.Type)
	return http.StatusOK, "Radarr Collections Gaps initiated."
}

func (a *Actions) corrupt(input *common.ActionInput, content string) (int, string) {
	title := cases.Title(language.AmericanEnglish)

	err := a.Backups.Corruption(input, starr.App(title.String(content)))
	if err != nil {
		return http.StatusBadRequest, "Corruption trigger failed: " + err.Error()
	}

	return http.StatusOK, title.String(content) + " corruption checks initiated."
}

func (a *Actions) backup(input *common.ActionInput, content string) (int, string) {
	title := cases.Title(language.AmericanEnglish)

	err := a.Backups.Backup(input, starr.App(title.String(content)))
	if err != nil {
		return http.StatusBadRequest, "Backup trigger failed: " + err.Error()
	}

	return http.StatusOK, title.String(content) + " backups check initiated."
}

func (a *Actions) handleConfigReload() (int, string) {
	defer a.timers.ReloadApp("HTTP Triggered Reload")
	return http.StatusOK, "Application reload initiated."
}

func (a *Actions) notification(content string) (int, string) {
	if content != "" {
		ui.Notify("Notification: %s", content) //nolint:errcheck
		a.timers.Printf("NOTIFICATION: %s", content)

		return http.StatusOK, "Local Nntification sent."
	}

	return http.StatusBadRequest, "Missing notification content."
}
