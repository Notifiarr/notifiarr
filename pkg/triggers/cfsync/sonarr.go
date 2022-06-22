package cfsync

import (
	"encoding/json"
	"fmt"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/starr/sonarr"
)

const TrigCFSyncSonarr common.TriggerName = "Starting Sonarr QP TRaSH sync."

// SonarrTrashPayload is the payload sent and received
// to/from notifarr.com when updating custom formats for Sonarr.
type SonarrTrashPayload struct {
	Instance        int                      `json:"instance"`
	Name            string                   `json:"name"`
	ReleaseProfiles []*sonarr.ReleaseProfile `json:"releaseProfiles,omitempty"`
	QualityProfiles []*sonarr.QualityProfile `json:"qualityProfiles,omitempty"`
	Error           string                   `json:"error"`
	NewMaps         *cfMapIDpayload          `json:"newMaps,omitempty"`
}

// syncSonarr triggers a custom format sync for Sonarr.
func (c *Config) syncSonarr(event website.EventType) {
	if c.ClientInfo == nil || len(c.ClientInfo.Actions.Sync.SonarrInstances) < 1 {
		c.Debugf("Cannot sync Sonarr Release Profiles. Website provided 0 instances.")
		return
	} else if len(c.Apps.Sonarr) < 1 {
		c.Debugf("Cannot sync Sonarr Release Profiles. No Sonarr instances configured.")
		return
	}

	for i, app := range c.Apps.Sonarr {
		instance := i + 1
		if app.URL == "" || app.APIKey == "" || !c.ClientInfo.Actions.Sync.SonarrInstances.Has(instance) {
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

func (c *Config) syncSonarrRP(instance int, app *apps.SonarrConfig) error {
	var (
		err     error
		payload = SonarrTrashPayload{Instance: instance, Name: app.Name, NewMaps: c.sonarrRP[instance]}
	)

	payload.QualityProfiles, err = app.GetQualityProfiles()
	if err != nil {
		return fmt.Errorf("getting quality profiles: %w", err)
	}

	payload.ReleaseProfiles, err = app.GetReleaseProfiles()
	if err != nil {
		return fmt.Errorf("getting release profiles: %w", err)
	}

	body, err := c.SendData(website.CFSyncRoute.Path("", "app=sonarr"), payload, false)
	if err != nil {
		return fmt.Errorf("sending current profiles: %w", err)
	}

	delete(c.sonarrRP, instance)

	if body.Result != success {
		return fmt.Errorf("%w: %s", website.ErrInvalidResponse, body.Result)
	}

	if err := c.updateSonarrRP(instance, app, body.Details.Response); err != nil {
		return fmt.Errorf("updating application: %w", err)
	}

	return nil
}

func (c *Config) updateSonarrRP(instance int, app *apps.SonarrConfig, data []byte) error {
	reply := &SonarrTrashPayload{}
	if err := json.Unmarshal(data, &reply); err != nil {
		return fmt.Errorf("bad json response: %w", err)
	}

	c.Debugf("Received %d quality profiles and %d release profiles for Sonarr: %d:%s",
		len(reply.QualityProfiles), len(reply.ReleaseProfiles), instance, app.URL)

	maps := &cfMapIDpayload{RP: []idMap{}, QP: []idMap{}, Instance: instance}

	for idx, profile := range reply.ReleaseProfiles {
		newID, existingID := profile.ID, profile.ID

		if _, err := app.UpdateReleaseProfile(profile); err != nil {
			profile.ID = 0

			c.Debugf("Error Updating release profile [%d/%d] (attempting to ADD %d): %v",
				idx+1, len(reply.ReleaseProfiles), existingID, err)

			newProfile, err2 := app.AddReleaseProfile(profile)
			if err2 != nil {
				c.Errorf("Ensuring release profile [%d/%d] %d: (update) %v, (add) %v",
					idx+1, len(reply.ReleaseProfiles), existingID, err, err2)
				continue
			}

			newID = newProfile.ID
		}

		maps.RP = append(maps.RP, idMap{profile.Name, existingID, newID})
	}

	for idx, profile := range reply.QualityProfiles {
		newID, existingID := profile.ID, profile.ID

		if _, err := app.UpdateQualityProfile(profile); err != nil {
			profile.ID = 0

			c.Debugf("Error Updating quality format [%d/%d] (attempting to ADD %d): %v",
				idx+1, len(reply.QualityProfiles), existingID, err)

			newProfile, err2 := app.AddQualityProfile(profile)
			if err2 != nil {
				c.Errorf("Ensuring quality format [%d/%d] %d: (update) %v, (add) %v",
					idx+1, len(reply.QualityProfiles), existingID, err, err2)
				continue
			}

			newID = newProfile.ID
		}

		maps.QP = append(maps.QP, idMap{profile.Name, existingID, newID})
	}

	return c.postbackSonarrRP(instance, maps)
}

// postbackSonarrRP sends the changes back to notifiarr.com.
func (c *Config) postbackSonarrRP(instance int, maps *cfMapIDpayload) error {
	if len(maps.QP) < 1 && len(maps.RP) < 1 {
		return nil
	}

	_, err := c.SendData(website.CFSyncRoute.Path("", "app=sonarr", "updateIDs=true"), &SonarrTrashPayload{
		Instance: instance,
		NewMaps:  maps,
	}, true)
	if err != nil {
		c.sonarrRP[instance] = maps
		return fmt.Errorf("updating quality release ID map: %w", err)
	}

	delete(c.sonarrRP, instance)

	return nil
}
