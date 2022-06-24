//nolint:dupl
package cfsync

import (
	"encoding/json"
	"fmt"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/starr/radarr"
)

const TrigCFSyncRadarr common.TriggerName = "Starting Radarr CF TRaSH sync."

// RadarrTrashPayload is the payload sent and received
// to/from notifarr.com when updating custom formats for Radarr.
// This is used in other places, like the trash API handler in the 'client' module.
type RadarrTrashPayload struct {
	Instance        int                      `json:"instance"`
	Name            string                   `json:"name"`
	CustomFormats   []*radarr.CustomFormat   `json:"customFormats,omitempty"`
	QualityProfiles []*radarr.QualityProfile `json:"qualityProfiles,omitempty"`
	Error           string                   `json:"error"`
	NewMaps         *cfMapIDpayload          `json:"newMaps,omitempty"`
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
		if app.URL == "" || app.APIKey == "" || !c.ClientInfo.Actions.Sync.RadarrInstances.Has(instance) {
			c.Debugf("CF Sync Skipping Radarr instance %d. Not in sync list: %v",
				instance, c.ClientInfo.Actions.Sync.RadarrInstances)
			continue
		}

		if err := c.syncRadarrCF(instance, app); err != nil {
			c.Errorf("[%s requested] Radarr Custom Formats sync request for '%d:%s' failed: %v", event, instance, app.URL, err)
			continue
		}

		c.Printf("[%s requested] Synced Custom Formats from Notifiarr for Radarr: %d:%s", event, instance, app.URL)
	}
}

func (c *cmd) syncRadarrCF(instance int, app *apps.RadarrConfig) error {
	var (
		err     error
		payload = RadarrTrashPayload{Instance: instance, Name: app.Name, NewMaps: c.radarrCF[instance]}
	)

	payload.QualityProfiles, err = app.GetQualityProfiles()
	if err != nil {
		return fmt.Errorf("getting quality profiles: %w", err)
	}

	payload.CustomFormats, err = app.GetCustomFormats()
	if err != nil {
		return fmt.Errorf("getting custom formats: %w", err)
	}

	body, err := c.SendData(website.CFSyncRoute.Path("", "app=radarr"), payload, false)
	if err != nil {
		return fmt.Errorf("sending current formats: %w", err)
	}

	delete(c.radarrCF, instance)

	if body.Result != success {
		return fmt.Errorf("%w: %s", website.ErrInvalidResponse, body.Result)
	}

	if err := c.updateRadarrCF(instance, app, body.Details.Response); err != nil {
		return fmt.Errorf("updating application: %w", err)
	}

	return nil
}

func (c *cmd) updateRadarrCF(instance int, app *apps.RadarrConfig, data []byte) error {
	reply := &RadarrTrashPayload{}
	if err := json.Unmarshal(data, &reply); err != nil {
		return fmt.Errorf("bad json response: %w", err)
	}

	c.Debugf("Received %d quality profiles and %d custom formats for Radarr: %d:%s",
		len(reply.QualityProfiles), len(reply.CustomFormats), instance, app.URL)

	maps := &cfMapIDpayload{QP: []idMap{}, CF: []idMap{}, Instance: instance}

	for idx, profile := range reply.CustomFormats {
		newID, existingID := profile.ID, profile.ID

		if _, err := app.UpdateCustomFormat(profile, existingID); err != nil {
			profile.ID = 0

			c.Debugf("Error Updating custom format [%d/%d] (attempting to ADD %d): %v",
				idx+1, len(reply.CustomFormats), existingID, err)

			newAdd, err2 := app.AddCustomFormat(profile)
			if err2 != nil {
				c.Errorf("Ensuring custom format [%d/%d] %d: (update) %v, (add) %v",
					idx+1, len(reply.CustomFormats), existingID, err, err2)
				continue
			}

			newID = newAdd.ID
		}

		maps.CF = append(maps.CF, idMap{profile.Name, int64(existingID), int64(newID)})
	}

	for idx, profile := range reply.QualityProfiles {
		newID, existingID := profile.ID, profile.ID

		if err := app.UpdateQualityProfile(profile); err != nil {
			profile.ID = 0

			c.Debugf("Error Updating quality profile [%d/%d] (attempting to ADD %d): %v",
				idx+1, len(reply.QualityProfiles), existingID, err)

			newAddID, err2 := app.AddQualityProfile(profile)
			if err2 != nil {
				c.Errorf("Ensuring quality profile [%d/%d] %d: (update) %v, (add) %v",
					idx+1, len(reply.QualityProfiles), existingID, err, err2)
				continue
			}

			newID = newAddID
		}

		maps.QP = append(maps.QP, idMap{profile.Name, existingID, newID})
	}

	return c.postbackRadarrCF(instance, maps)
}

// postbackRadarrCF sends the changes back to notifiarr.com.
func (c *cmd) postbackRadarrCF(instance int, maps *cfMapIDpayload) error {
	if len(maps.CF) < 1 && len(maps.QP) < 1 {
		return nil
	}

	_, err := c.SendData(website.CFSyncRoute.Path("", "app=radarr", "updateIDs=true"), &RadarrTrashPayload{
		Instance: instance,
		NewMaps:  maps,
	}, true)
	if err != nil {
		c.radarrCF[instance] = maps
		return fmt.Errorf("updating custom format ID map: %w", err)
	}

	delete(c.radarrCF, instance)

	return nil
}
