package notifiarr

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"golift.io/cnfg"
)

// EventType identifies the type of event that sent a paylaod to notifiarr.
type EventType string

// These are all our known event types.
const (
	EventCron    EventType = "cron"
	EventUser    EventType = "user"
	EventAPI     EventType = "api"
	EventHook    EventType = "webhook"
	EventStart   EventType = "start"
	EventMovie   EventType = "movie"
	EventEpisode EventType = "episode"
	EventPoll    EventType = "poll"
	EventReload  EventType = "reload"
)

// Payload is the outbound payload structure that is sent to Notifiarr for Plex and system snapshot data.
type Payload struct {
	Plex *plex.Sessions       `json:"plex,omitempty"`
	Snap *snapshot.Snapshot   `json:"snapshot,omitempty"`
	Load *plexIncomingWebhook `json:"payload,omitempty"`
}

// Route is used to give us methods on our route paths.
type Route string

// Notifiarr URLs. Data sent to these URLs:
/*
api/v1/notification/plex?event=...
  api (was plexcron)
  user (was plexcron)
  cron (was plexcron)
  webhook (was plexhook)
  movie
  episode

api/v1/notification/services?event=...
  api
  user
  cron
  start (only fires on startup)

api/v1/notification/snapshot?event=...
  api
  user
  cron

api/v1/notification/dashboard?event=... (requires interval from website/client endpoint)
  api
  user
  cron

api/v1/notification/stuck?event=...
  api
  user
  cron

api/v1/user/gaps?app=radarr&event=...
  api
  user
  cron

api/v2/user/client?event=start
  see description https://github.com/Notifiarr/notifiarr/pull/115

api/v1/user/trash?app=...
  radarr
  sonarr
*/
const (
	BaseURL            = "https://notifiarr.com"
	DevBaseURL         = "https://dev.notifiarr.com"
	userRoute1   Route = "/api/v1/user"
	userRoute2   Route = "/api/v2/user"
	ClientRoute  Route = userRoute2 + "/client"
	CFSyncRoute  Route = userRoute1 + "/trash"
	GapsRoute    Route = userRoute1 + "/gaps"
	notifiRoute  Route = "/api/v1/notification"
	DashRoute    Route = notifiRoute + "/dashboard"
	StuckRoute   Route = notifiRoute + "/stuck"
	PlexRoute    Route = notifiRoute + "/plex"
	SnapRoute    Route = notifiRoute + "/snapshot"
	SvcRoute     Route = notifiRoute + "/services"
	CorruptRoute Route = notifiRoute + "/corruption"
	BackupRoute  Route = notifiRoute + "/backup"
	TestRoute    Route = notifiRoute + "/test"
)

// Path adds parameters to a route path and turns it into a string.
func (r Route) Path(event EventType, params ...string) string {
	switch {
	case len(params) == 0 && event == "":
		return string(r)
	case len(params) == 0:
		return string(r) + "?event=" + string(event)
	case event == "":
		return string(r) + "?" + strings.Join(params, "&")
	default:
		return string(r) + "?" + strings.Join(append(params, "event="+string(event)), "&")
	}
}

// Response is what notifiarr replies to our requests with.
/* try this
{
    "response": "success",
    "message": {
        "response": {
            "instance": 1,
            "debug": null
        },
        "started": "23:57:03",
        "finished": "23:57:03",
        "elapsed": "0s"
    }
}

{
    "response": "success",
    "message": {
        "response": "Service status cron processed.",
        "started": "00:04:15",
        "finished": "00:04:15",
        "elapsed": "0s"
    }
}

{
    "response": "success",
    "message": {
        "response": "Channel stats cron processed.",
        "started": "00:04:31",
        "finished": "00:04:36",
        "elapsed": "5s"
    }
}

{
    "response": "success",
    "message": {
        "response": "Dashboard payload processed.",
        "started": "00:02:04",
        "finished": "00:02:11",
        "elapsed": "7s"
    }
}
*/
// nitsua: all responses should be that way.. but response might not always be an object.
type Response struct {
	Result  string `json:"result"`
	Details struct {
		Response json.RawMessage `json:"response"` // can be anything. type it out later.
		Help     string          `json:"help"`
		Started  time.Time       `json:"started"`
		Finished time.Time       `json:"finished"`
		Elapsed  cnfg.Duration   `json:"elapsed"`
	} `json:"details"`
}

// String turns the response into a log entry.
func (r *Response) String() string {
	if r == nil {
		return ""
	}

	return fmt.Sprintf("Website took %s and replied with: %s, %s %s",
		r.Details.Elapsed, r.Result, r.Details.Response, r.Details.Help)
}
