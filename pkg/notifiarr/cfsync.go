//nolint:dupl
package notifiarr

import (
	"encoding/json"
	"fmt"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"golift.io/cnfg"
	"golift.io/starr/radarr"
	"golift.io/starr/sonarr"
)

/* CF Sync means Custom Format Sync. This is a premium feature that allows syncing
   TRaSH's custom Radarr formats and Sonarr Release Profiles.
	 The code in this file deals with sending data and getting updates at an interval.
*/

// syncConfig is the configuration returned from the notifiarr website.
type syncConfig struct {
	Interval        cnfg.Duration `json:"interval"`        // how often to fire in minutes.
	Radarr          int64         `json:"radarr"`          // items in sync
	RadarrInstances IntList       `json:"radarrInstances"` // which instance IDs we sync
	Sonarr          int64         `json:"sonarr"`          // items in sync
	SonarrInstances IntList       `json:"sonarrInstances"` // which instance IDs we sync
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

/*//*****/ // Radarr /*//*****///

// RadarrTrashPayload is the payload sent and received
// to/from notifarr.com when updating custom formats for Radarr.
// This is used in other places, like the trash API handler in the 'client' module.
type RadarrTrashPayload struct {
	Instance        int                      `json:"instance"`
	Name            string                   `json:"name"`
	CustomFormats   []*radarr.CustomFormat   `json:"customFormats,omitempty"`
	QualityProfiles []*radarr.QualityProfile `json:"qualityProfiles,omitempty"`
	Error           string                   `json:"error"`
	// Purposely not exported so as to not use it externally.
	NewMaps *cfMapIDpayload `json:"newMaps,omitempty"`
}

func (t *Triggers) SyncCF(event EventType) {
	t.exec(event, (TrigCFSync))
}

func (c *Config) syncCF(event EventType) {
	c.Debugf("Running CF Sync via event: %s", event)
	c.syncRadarr(event)
	c.syncSonarr(event)
}

// syncRadarr triggers a custom format sync for Radarr.
func (c *Config) syncRadarr(event EventType) {
	if c.clientInfo == nil || len(c.clientInfo.Actions.Sync.RadarrInstances) < 1 {
		c.Debugf("Cannot sync Radarr Custom Formats. Website provided 0 instances.")
		return
	} else if len(c.Apps.Radarr) < 1 {
		c.Debugf("Cannot sync Radarr Custom Formats. No Radarr instances configured.")
		return
	}

	for i, app := range c.Apps.Radarr {
		instance := i + 1
		if app.URL == "" || app.APIKey == "" || !c.clientInfo.Actions.Sync.RadarrInstances.Has(instance) {
			c.Debugf("CF Sync Skipping Radarr instance %d. Not in sync list: %v",
				instance, c.clientInfo.Actions.Sync.RadarrInstances)
			continue
		}

		if err := c.syncRadarrCF(instance, app); err != nil {
			c.Errorf("[%s requested] Radarr Custom Formats sync request for '%d:%s' failed: %v", event, instance, app.URL, err)
			continue
		}

		c.Printf("[%s requested] Synced Custom Formats from Notifiarr for Radarr: %d:%s", event, instance, app.URL)
	}
}

func (c *Config) syncRadarrCF(instance int, app *apps.RadarrConfig) error {
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

	body, err := c.SendData(CFSyncRoute.Path("", "app=radarr"), payload, false)
	if err != nil {
		return fmt.Errorf("sending current formats: %w", err)
	}

	delete(c.radarrCF, instance)

	if body.Result != success {
		return fmt.Errorf("%w: %s", ErrInvalidResponse, body.Result)
	}

	if err := c.updateRadarrCF(instance, app, body.Details.Response); err != nil {
		return fmt.Errorf("updating application: %w", err)
	}

	return nil
}

func (c *Config) updateRadarrCF(instance int, app *apps.RadarrConfig, data []byte) error {
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
func (c *Config) postbackRadarrCF(instance int, maps *cfMapIDpayload) error {
	if len(maps.CF) < 1 && len(maps.QP) < 1 {
		return nil
	}

	_, err := c.SendData(CFSyncRoute.Path("", "app=radarr", "updateIDs=true"), &RadarrTrashPayload{
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

/*//*****/ // Sonarr /*//*****///

// SonarrTrashPayload is the payload sent and received
// to/from notifarr.com when updating custom formats for Sonarr.
type SonarrTrashPayload struct {
	Instance        int                      `json:"instance"`
	Name            string                   `json:"name"`
	ReleaseProfiles []*sonarr.ReleaseProfile `json:"releaseProfiles,omitempty"`
	QualityProfiles []*sonarr.QualityProfile `json:"qualityProfiles,omitempty"`
	Error           string                   `json:"error"`
	// Purposely not exported so as to not use it externally.
	NewMaps *cfMapIDpayload `json:"newMaps,omitempty"`
}

// syncSonarr triggers a custom format sync for Sonarr.
func (c *Config) syncSonarr(event EventType) {
	if c.clientInfo == nil || len(c.clientInfo.Actions.Sync.SonarrInstances) < 1 {
		c.Debugf("Cannot sync Sonarr Release Profiles. Website provided 0 instances.")
		return
	} else if len(c.Apps.Sonarr) < 1 {
		c.Debugf("Cannot sync Sonarr Release Profiles. No Sonarr instances configured.")
		return
	}

	for i, app := range c.Apps.Sonarr {
		instance := i + 1
		if app.URL == "" || app.APIKey == "" || !c.clientInfo.Actions.Sync.SonarrInstances.Has(instance) {
			c.Debugf("CF Sync Skipping Sonarr instance %d. Not in sync list: %v",
				instance, c.clientInfo.Actions.Sync.SonarrInstances)
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

	body, err := c.SendData(CFSyncRoute.Path("", "app=sonarr"), payload, false)
	if err != nil {
		return fmt.Errorf("sending current profiles: %w", err)
	}

	delete(c.sonarrRP, instance)

	if body.Result != success {
		return fmt.Errorf("%w: %s", ErrInvalidResponse, body.Result)
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

	_, err := c.SendData(CFSyncRoute.Path("", "app=sonarr", "updateIDs=true"), &SonarrTrashPayload{
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
