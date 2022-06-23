package cfsync

import (
	"time"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

/* CF Sync means Custom Format Sync. This is a premium feature that allows syncing
   TRaSH's custom Radarr formats and Sonarr Release Profiles.
	 The code in this file deals with sending data and getting updates at an interval.
*/

type Action struct {
	*common.Config
	radarrCF map[int]*cfMapIDpayload
	sonarrRP map[int]*cfMapIDpayload
}

// cfMapIDpayload is used to post-back ID changes for profiles and formats.
type cfMapIDpayload struct {
	Instance int     `json:"instance"`
	RP       []idMap `json:"releaseProfiles,omitempty"`
	QP       []idMap `json:"qualityProfiles,omitempty"`
	CF       []idMap `json:"customFormats,omitempty"`
}

// idMap is used a mapping list from old ID to new ID. Part of cfMapIDpayload.
type idMap struct {
	Name  string `json:"name"`
	OldID int64  `json:"oldId"`
	NewID int64  `json:"newId"`
}

// success is a ssuccessful status message from notifiarr.com.
const success = "success"

func (c *Action) Create() {
	c.radarrCF = make(map[int]*cfMapIDpayload)
	c.sonarrRP = make(map[int]*cfMapIDpayload)
	ci := c.ClientInfo

	var (
		radarrTicker *time.Ticker
		sonarrTicker *time.Ticker
	)

	// Check each instance and enable only if needed.
	if ci != nil && ci.Actions.Sync.Interval.Duration > 0 {
		if len(ci.Actions.Sync.RadarrInstances) > 0 {
			radarrTicker = time.NewTicker(ci.Actions.Sync.Interval.Duration)
			c.Printf("==> Keeping %d Radarr Custom Formats synced, interval:%s",
				ci.Actions.Sync.Radarr, ci.Actions.Sync.Interval)
		}

		if len(ci.Actions.Sync.SonarrInstances) > 0 {
			sonarrTicker = time.NewTicker(ci.Actions.Sync.Interval.Duration)
			c.Printf("==> Keeping %d Sonarr Release Profiles synced, interval:%s",
				ci.Actions.Sync.Sonarr, ci.Actions.Sync.Interval)
		}
	}

	c.Add(&common.Action{
		Name: TrigCFSyncRadarr,
		Fn:   c.syncRadarr,
		C:    make(chan website.EventType, 1),
		T:    radarrTicker,
	}, &common.Action{
		Name: TrigCFSyncSonarr,
		Fn:   c.syncSonarr,
		C:    make(chan website.EventType, 1),
		T:    sonarrTicker,
	})
}
