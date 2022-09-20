package triggers

import (
	"net/http"
	"strconv"

	"github.com/Notifiarr/notifiarr/pkg/logs/share"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/gorilla/mux"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golift.io/starr"
)

// Handler is passed into the webserver so triggers can be executed from the API.
func (a *Actions) Handler(r *http.Request) (int, interface{}) {
	return a.runTrigger(website.EventAPI, mux.Vars(r)["trigger"], mux.Vars(r)["content"])
}

// RunTrigger is fired by GUI web requests and possibly other places.
func (a *Actions) RunTrigger(source website.EventType, trigger, content string) (int, string) {
	return a.runTrigger(source, trigger, content)
}

func (a *Actions) runTrigger(source website.EventType, trigger, content string) (int, string) { //nolint:cyclop
	if content != "" {
		a.timers.Debugf("Incoming API Trigger: %s (%s)", trigger, content)
	} else {
		a.timers.Debugf("Incoming API Trigger: %s", trigger)
	}

	switch trigger {
	case "clientlogs":
		return a.clientLogs(content)
	case "command":
		return a.command(source, content)
	case "cfsync":
		return a.cfsync(source, content)
	case "rpsync":
		return a.rpsync(source, content)
	case "services":
		return a.services(source)
	case "sessions":
		return a.sessions(source)
	case "stuckitems":
		return a.stuckitems(source)
	case "dashboard":
		return a.dashboard(source)
	case "snapshot":
		return a.snapshot(source)
	case "gaps":
		return a.gaps(source)
	case "corrupt":
		return a.corrupt(source, content)
	case "backup":
		return a.backup(source, content)
	case "reload":
		// reload is handled in another package and never gets triggered here.
		return http.StatusExpectationFailed, "Impossible Code Path Reached"
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

func (a *Actions) command(source website.EventType, content string) (int, string) {
	cmd := a.Commands.GetByHash(content)
	if cmd == nil {
		return http.StatusBadRequest, "No command hash provided."
	}

	cmd.Run(source)

	return http.StatusOK, "Command triggered: " + cmd.Name
}

func (a *Actions) cfsync(source website.EventType, content string) (int, string) {
	if content == "" {
		a.CFSync.SyncRadarrCF(source)
		return http.StatusOK, "TRaSH Custom Formats Radarr Sync initiated."
	}

	instance, _ := strconv.Atoi(content)
	if err := a.CFSync.SyncRadarrInstanceCF(source, instance); err != nil {
		return http.StatusBadRequest, "TRaSH Custom Formats Radarr Sync failed for instance " + content + ": " + err.Error()
	}

	return http.StatusOK, "TRaSH Custom Formats Radarr Sync initiated for instance " + content
}

func (a *Actions) rpsync(source website.EventType, content string) (int, string) {
	if content == "" {
		a.CFSync.SyncSonarrRP(source)
		return http.StatusOK, "TRaSH Release Profile Sonarr Sync initiated."
	}

	instance, _ := strconv.Atoi(content)
	if err := a.CFSync.SyncSonarrInstanceRP(source, instance); err != nil {
		return http.StatusBadRequest, "TRaSH Release Profile Sonarr Sync failed for instance " + content + ": " + err.Error()
	}

	return http.StatusOK, "TRaSH Release Profile Sonarr Sync initiated for instance " + content
}

func (a *Actions) services(source website.EventType) (int, string) {
	a.timers.RunChecks(source)
	return http.StatusOK, "All service checks rescheduled for immediate exeution."
}

func (a *Actions) sessions(source website.EventType) (int, string) {
	if !a.timers.Apps.Plex.Enabled() {
		return http.StatusNotImplemented, "Plex Sessions are not enabled."
	}

	a.PlexCron.Send(source)

	return http.StatusOK, "Plex sessions triggered."
}

func (a *Actions) stuckitems(source website.EventType) (int, string) {
	a.StarrQueue.StuckItems(source)
	return http.StatusOK, "Stuck Queue Items triggered."
}

func (a *Actions) dashboard(source website.EventType) (int, string) {
	a.Dashboard.Send(source)
	return http.StatusOK, "Dashboard states triggered."
}

func (a *Actions) snapshot(source website.EventType) (int, string) {
	a.SnapCron.Send(source)
	return http.StatusOK, "System Snapshot triggered."
}

func (a *Actions) gaps(source website.EventType) (int, string) {
	a.Gaps.Send(source)
	return http.StatusOK, "Radarr Collections Gaps initiated."
}

func (a *Actions) corrupt(source website.EventType, content string) (int, string) {
	title := cases.Title(language.AmericanEnglish)

	err := a.Backups.Corruption(source, starr.App(title.String(content)))
	if err != nil {
		return http.StatusBadRequest, "Corruption trigger failed: " + err.Error()
	}

	return http.StatusOK, title.String(content) + " corruption checks initiated."
}

func (a *Actions) backup(source website.EventType, content string) (int, string) {
	title := cases.Title(language.AmericanEnglish)

	err := a.Backups.Backup(source, starr.App(title.String(content)))
	if err != nil {
		return http.StatusBadRequest, "Backup trigger failed: " + err.Error()
	}

	return http.StatusOK, title.String(content) + " backups check initiated."
}

func (a *Actions) notification(content string) (int, string) {
	if content != "" {
		ui.Notify("Notification: %s", content) //nolint:errcheck
		a.timers.Printf("NOTIFICATION: %s", content)

		return http.StatusOK, "Local Nntification sent."
	}

	return http.StatusBadRequest, "Missing notification content."
}
