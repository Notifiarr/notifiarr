package cfsync

import (
	"errors"
	"fmt"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/starr"
	"golift.io/starr/sonarr"
)

const TrigRPSyncSonarr common.TriggerName = "Starting Sonarr Release Profile TRaSH sync."

// SonarrTrashPayload is the payload sent and received
// to/from notifarr.com when updating custom formats for Sonarr.
type SonarrTrashPayload struct {
	Instance           int                         `json:"instance"`
	Name               string                      `json:"name"`
	ReleaseProfiles    []*sonarr.ReleaseProfile    `json:"releaseProfiles,omitempty"`
	QualityProfiles    []*sonarr.QualityProfile    `json:"qualityProfiles,omitempty"`
	CustomFormats      []*sonarr.CustomFormat      `json:"customFormats,omitempty"`
	QualityDefinitions []*sonarr.QualityDefinition `json:"qualityDefinitions,omitempty"`
	Error              string                      `json:"error"`
}

// SyncSonarrRP initializes a release profile sync with sonarr.
func (a *Action) SyncSonarrRP(event website.EventType) {
	a.cmd.Exec(event, TrigRPSyncSonarr)
}

// syncSonarr triggers a custom format sync for Sonarr.
func (c *cmd) syncSonarr(event website.EventType) {
	if c.ClientInfo == nil || len(c.ClientInfo.Actions.Sync.SonarrInstances) < 1 {
		c.Debugf("Cannot sync Sonarr Release Profiles. Website provided 0 instances.")
		return
	} else if len(c.Apps.Sonarr) < 1 {
		c.Debugf("Cannot sync Sonarr Release Profiles. No Sonarr instances configured.")
		return
	}

	for i, app := range c.Apps.Sonarr {
		instance := i + 1
		if app.URL == "" || app.APIKey == "" || app.Timeout.Duration < 0 ||
			!c.ClientInfo.Actions.Sync.SonarrInstances.Has(instance) {
			c.Debugf("CF Sync Skipping Sonarr instance %d. Not in sync list: %v",
				instance, c.ClientInfo.Actions.Sync.SonarrInstances)
			continue
		}

		if err := c.syncSonarrRP(instance, app); err != nil {
			c.Errorf("[%s requested] Sonarr Release Profiles sync for '%d:%s' failed: %v", event, instance, app.URL, err)
			continue
		}

		c.Printf("[%s requested] Synced Sonarr Release Profiles from Notifiarr: %d:%s", event, instance, app.URL)
	}
}

func (c *cmd) syncSonarrRP(instance int, app *apps.SonarrConfig) error {
	var (
		err     error
		payload = SonarrTrashPayload{Instance: instance, Name: app.Name}
		start   = time.Now()
	)

	payload.QualityProfiles, err = app.GetQualityProfiles()
	if err != nil {
		payload.Error += fmt.Sprintf("getting quality profiles: %v ", err)
		return fmt.Errorf("getting quality profiles: %w", err)
	}

	payload.ReleaseProfiles, err = app.GetReleaseProfiles()
	if err != nil {
		return fmt.Errorf("getting release profiles: %w", err)
	}

	payload.QualityDefinitions, err = app.GetQualityDefinitions()
	if err != nil {
		return fmt.Errorf("getting quality definitions: %w", err)
	}

	payload.CustomFormats, err = app.GetCustomFormats()
	if err != nil && !errors.Is(err, starr.ErrInvalidStatusCode) {
		return fmt.Errorf("getting custom formats: %w", err)
	}

	c.SendData(&website.Request{
		Route:      website.CFSyncRoute,
		Params:     []string{"app=sonarr"},
		Payload:    payload,
		LogMsg:     fmt.Sprintf("Sonarr TRaSH Sync (elapsed: %v)", time.Since(start).Round(time.Millisecond)),
		LogPayload: true,
	})

	return nil
}
