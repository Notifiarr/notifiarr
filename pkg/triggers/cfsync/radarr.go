//nolint:dupl
package cfsync

import (
	"fmt"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/starr/radarr"
)

const TrigCFSyncRadarr common.TriggerName = "Starting Radarr Custom Format TRaSH sync."

// RadarrTrashPayload is the payload sent and received
// to/from notifarr.com when updating custom formats for Radarr.
// This is used in other places, like the trash API handler in the 'client' module.
type RadarrTrashPayload struct {
	Instance           int                         `json:"instance"`
	Name               string                      `json:"name"`
	CustomFormats      []*radarr.CustomFormat      `json:"customFormats,omitempty"`
	QualityProfiles    []*radarr.QualityProfile    `json:"qualityProfiles,omitempty"`
	QualityDefinitions []*radarr.QualityDefinition `json:"qualityDefinitions,omitempty"`
	Error              string                      `json:"error"`
}

// SyncRadarrCF initializes a custom format sync with radarr.
func (a *Action) SyncRadarrCF(event website.EventType) {
	a.cmd.Exec(event, TrigCFSyncRadarr)
}

// syncRadarr triggers a custom format sync for Radarr.
func (c *cmd) syncRadarr(event website.EventType) {
	if c.ClientInfo == nil || len(c.ClientInfo.Actions.Sync.RadarrInstances) < 1 {
		c.Debugf("Cannot sync Radarr Custom Formats. Website provided 0 instances.")
		return
	} else if len(c.Apps.Radarr) < 1 {
		c.Debugf("Cannot sync Radarr Custom Formats. No Radarr instances configured.")
		return
	}

	for i, app := range c.Apps.Radarr {
		instance := i + 1
		if app.URL == "" || app.APIKey == "" || app.Timeout.Duration < 0 ||
			!c.ClientInfo.Actions.Sync.RadarrInstances.Has(instance) {
			c.Debugf("CF Sync Skipping Radarr instance %d. Not in sync list: %v",
				instance, c.ClientInfo.Actions.Sync.RadarrInstances)
			continue
		}

		if err := c.syncRadarrCF(event, instance, app); err != nil {
			c.Errorf("[%s requested] Radarr Custom Formats sync request for '%d:%s' failed: %v", event, instance, app.URL, err)
			continue
		}

		c.Printf("[%s requested] Synced Custom Formats from Notifiarr for Radarr: %d:%s", event, instance, app.URL)
	}
}

func (c *cmd) syncRadarrCF(event website.EventType, instance int, app *apps.RadarrConfig) error {
	var (
		err     error
		payload = RadarrTrashPayload{Instance: instance, Name: app.Name}
		start   = time.Now()
	)

	payload.QualityProfiles, err = app.GetQualityProfiles()
	if err != nil {
		return fmt.Errorf("getting quality profiles: %w", err)
	}

	payload.CustomFormats, err = app.GetCustomFormats()
	if err != nil {
		return fmt.Errorf("getting custom formats: %w", err)
	}

	payload.QualityDefinitions, err = app.GetQualityDefinitions()
	if err != nil {
		return fmt.Errorf("getting quality definitions: %w", err)
	}

	c.SendData(&website.Request{
		Route:      website.CFSyncRoute,
		Event:      event,
		Params:     []string{"app=radarr"},
		Payload:    payload,
		LogMsg:     fmt.Sprintf("Radarr TRaSH Sync (elapsed: %v)", time.Since(start).Round(time.Millisecond)),
		LogPayload: true,
		ErrorsOnly: false,
	})

	return nil
}
