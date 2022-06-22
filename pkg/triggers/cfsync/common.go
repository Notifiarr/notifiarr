package cfsync

import "github.com/Notifiarr/notifiarr/pkg/triggers/common"

/* CF Sync means Custom Format Sync. This is a premium feature that allows syncing
   TRaSH's custom Radarr formats and Sonarr Release Profiles.
	 The code in this file deals with sending data and getting updates at an interval.
*/

type Config struct {
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
