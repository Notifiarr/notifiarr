package notifiarr

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"golift.io/starr/radarr"
	"golift.io/starr/sonarr"
)

// syncConfig is the configuration returned from the notifiarr website.
type syncConfig struct {
	Instances intList `json:"instances"` // which instance IDs we sync
	Minutes   int     `json:"timer"`     // how often to fire in minutes.
	URI       string  `json:"endpoint"`  // "api/v1/user/sync"
	Radarr    int64   `json:"radarr"`    // items in sync
	Sonarr    int64   `json:"sonarr"`    // items in sync
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
	OldID int64 `json:"oldId"`
	NewID int64 `json:"newId"`
}

/*//*****/ // Radarr /*//*****///

// RadarrCustomFormatPayload is the payload sent and received
// to/from notifarr.com when updating custom formats for Radarr.
type RadarrCustomFormatPayload struct {
	Instance        int                      `json:"instance"`
	Name            string                   `json:"name"`
	CustomFormats   []*radarr.CustomFormat   `json:"customFormats,omitempty"`
	QualityProfiles []*radarr.QualityProfile `json:"qualityProfiles,omitempty"`
	NewMaps         *cfMapIDpayload          `json:"newMaps,omitempty"`
}

func (t *Triggers) SyncCF(wait bool) {
	if t.stop == nil {
		return
	}

	if !wait {
		t.syncCF <- nil
		return
	}

	reply := make(chan struct{})
	t.syncCF <- reply
	<-reply
}

func (c *Config) syncCF(reply chan struct{}) {
	c.syncRadarr()
	c.syncSonarr()

	if reply != nil {
		reply <- struct{}{}
	}
}

// syncRadarr triggers a custom format sync for Radarr.
func (c *Config) syncRadarr() {
	ci, err := c.GetClientInfo()
	if err != nil {
		c.Errorf("Cannot sync Radarr Custom Formats. Error: %v", err)
		return
	} else if ci.Actions.Sync.Radarr < 1 {
		c.Debugf("Cannot sync Radarr Custom Formats. Website provided 0 instances.")
		return
	}

	for i, app := range c.Apps.Radarr {
		instance := i + 1
		if app.URL == "" || app.APIKey == "" || !ci.Actions.Sync.Instances.Has(instance) {
			continue
		}

		switch synced, err := c.syncRadarrCF(instance, app); {
		case err != nil:
			c.Errorf("Radarr Custom Formats sync request for '%d:%s' failed: %v", instance, app.URL, err)
		case synced:
			c.Printf("Synced Custom Formats from Notifiarr for Radarr: %d:%s", instance, app.URL)
		default:
			c.Printf("Sent Custom Formats sync request to Notifiarr for Radarr: %d:%s", instance, app.URL)
		}
	}
}

func (c *Config) syncRadarrCF(instance int, app *apps.RadarrConfig) (bool, error) {
	var (
		err     error
		payload = RadarrCustomFormatPayload{Instance: instance, Name: app.Name, NewMaps: c.radarrCF[instance]}
	)

	payload.QualityProfiles, err = app.GetQualityProfiles()
	if err != nil {
		return false, fmt.Errorf("getting quality profiles: %w", err)
	}

	payload.CustomFormats, err = app.GetCustomFormats()
	if err != nil {
		return false, fmt.Errorf("getting custom formats: %w", err)
	}

	//nolint:bodyclose // already closed
	resp, body, err := c.SendData(c.BaseURL+CFSyncRoute+"?app=radarr", payload, false)
	if err != nil {
		return false, fmt.Errorf("sending current formats: %w", err)
	} else if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("%w: %s", ErrNon200, resp.Status)
	}

	delete(c.radarrCF, instance)

	if len(body) < 1 {
		return false, nil
	} else if err := c.updateRadarrCF(instance, app, body); err != nil {
		return false, fmt.Errorf("updating application: %w", err)
	}

	return true, nil
}

func (c *Config) updateRadarrCF(instance int, app *apps.RadarrConfig, data []byte) error {
	payload := struct {
		Response string `json:"response"`
		Message  struct {
			RadarrCustomFormatPayload
		} `json:"message"`
	}{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("bad json response: %w", err)
	}

	reply := payload.Message
	c.Debugf("Received %d quality profiles and %d custom formats for Radarr: %d:%s",
		len(reply.QualityProfiles), len(reply.CustomFormats), instance, app.URL)

	if payload.Response != success {
		return fmt.Errorf("%w: %s", ErrInvalidResponse, payload.Response)
	}

	maps := &cfMapIDpayload{QP: []idMap{}, CF: []idMap{}, Instance: instance}

	for i, profile := range reply.CustomFormats {
		id := profile.ID
		if _, err := app.UpdateCustomFormat(profile, id); err != nil {
			profile.ID = 0

			c.Debugf("Error Updating custom format [%d/%d] (attempting to ADD %d): %v",
				i, len(reply.CustomFormats), id, err)

			newID, err2 := app.AddCustomFormat(profile)
			if err2 != nil {
				return fmt.Errorf("[%d/%d] updating custom format: %d: (update) %v, (add) %w",
					i, len(reply.CustomFormats), id, err, err2)
			}

			maps.CF = append(maps.CF, idMap{int64(id), int64(newID.ID)})
		}
	}

	for i, profile := range reply.QualityProfiles {
		if err := app.UpdateQualityProfile(profile); err != nil {
			id := profile.ID
			profile.ID = 0

			c.Debugf("Error Updating quality profile [%d/%d] (attempting to ADD %d): %v",
				i, len(reply.QualityProfiles), id, err)

			newID, err2 := app.AddQualityProfile(profile)
			if err2 != nil {
				return fmt.Errorf("[%d/%d] updating quality profile: %d: (update) %v, (add) %w",
					i, len(reply.QualityProfiles), id, err, err2)
			}

			maps.QP = append(maps.QP, idMap{id, newID})
		}
	}

	return c.postbackRadarrCF(instance, maps)
}

// postbackRadarrCF sends the changes back to notifiarr.com.
func (c *Config) postbackRadarrCF(instance int, maps *cfMapIDpayload) error {
	if len(maps.CF) < 1 && len(maps.QP) < 1 {
		return nil
	}

	//nolint:bodyclose // already closed.
	resp, _, err := c.SendData(c.BaseURL+CFSyncRoute+"?app=radarr&updateIDs=true", &RadarrCustomFormatPayload{
		Instance: instance,
		NewMaps:  maps,
	}, false)
	if err != nil {
		c.radarrCF[instance] = maps
		return fmt.Errorf("updating custom format ID map: %w", err)
	} else if resp.StatusCode != http.StatusOK {
		c.radarrCF[instance] = maps
		return fmt.Errorf("updating custom format ID map: %w: %s", ErrNon200, resp.Status)
	}

	delete(c.radarrCF, instance)

	return nil
}

/*//*****/ // Sonarr /*//*****///

// SonarrCustomFormatPayload is the payload sent and received
// to/from notifarr.com when updating custom formats for Sonarr.
type SonarrCustomFormatPayload struct {
	Instance        int                      `json:"instance"`
	Name            string                   `json:"name"`
	ReleaseProfiles []*sonarr.ReleaseProfile `json:"releaseProfiles,omitempty"`
	QualityProfiles []*sonarr.QualityProfile `json:"qualityProfiles,omitempty"`
	NewMaps         *cfMapIDpayload          `json:"newMaps,omitempty"`
}

// syncSonarr triggers a custom format sync for Sonarr.
func (c *Config) syncSonarr() {
	ci, err := c.GetClientInfo()
	if err != nil {
		c.Debugf("Cannot sync Sonarr Release Profiles. Error: %v", err)
		return
	} else if ci.Actions.Sync.Sonarr < 1 {
		c.Debugf("Cannot sync Sonarr Release Profiles. Website provided 0 instances.")
		return
	}

	for i, app := range c.Apps.Sonarr {
		instance := i + 1
		if app.URL == "" || app.APIKey == "" || !ci.Actions.Sync.Instances.Has(instance) {
			continue
		}

		switch synced, err := c.syncSonarrRP(instance, app); {
		case err != nil:
			c.Errorf("Sonarr Release Profiles sync for '%d:%s' failed: %v", instance, app.URL, err)
		case synced:
			c.Printf("Synced Release Profiles from Notifiarr for Sonarr: %d:%s", instance, app.URL)
		default:
			c.Printf("Sent Release Profiles sync request to Notifiarr for Sonarr: %d:%s", instance, app.URL)
		}
	}
}

func (c *Config) syncSonarrRP(instance int, app *apps.SonarrConfig) (bool, error) {
	var (
		err     error
		payload = SonarrCustomFormatPayload{Instance: instance, Name: app.Name, NewMaps: c.sonarrRP[instance]}
	)

	payload.QualityProfiles, err = app.GetQualityProfiles()
	if err != nil {
		return false, fmt.Errorf("getting quality profiles: %w", err)
	}

	payload.ReleaseProfiles, err = app.GetReleaseProfiles()
	if err != nil {
		return false, fmt.Errorf("getting release profiles: %w", err)
	}

	//nolint:bodyclose // already closed
	resp, body, err := c.SendData(c.BaseURL+CFSyncRoute+"?app=sonarr", payload, false)
	if err != nil {
		return false, fmt.Errorf("sending current profiles: %w", err)
	} else if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("%w: %s", ErrNon200, resp.Status)
	}

	delete(c.sonarrRP, instance)

	if len(body) < 1 {
		return false, nil
	} else if err := c.updateSonarrRP(instance, app, body); err != nil {
		return false, fmt.Errorf("updating application: %w", err)
	}

	return true, nil
}

func (c *Config) updateSonarrRP(instance int, app *apps.SonarrConfig, data []byte) error {
	payload := struct {
		Response string `json:"response"`
		Message  struct {
			SonarrCustomFormatPayload
		} `json:"message"`
	}{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("bad json response: %w", err)
	}

	reply := payload.Message
	c.Debugf("Received %d quality profiles and %d release profiles for Sonarr: %d:%s",
		len(reply.QualityProfiles), len(reply.ReleaseProfiles), instance, app.URL)

	if payload.Response != success {
		return fmt.Errorf("%w: %s", ErrInvalidResponse, payload.Response)
	}

	maps := &cfMapIDpayload{RP: []idMap{}, QP: []idMap{}, Instance: instance}

	for i, profile := range reply.ReleaseProfiles {
		if err := app.UpdateReleaseProfile(profile); err != nil {
			id := profile.ID
			profile.ID = 0

			c.Debugf("Error Updating release profile [%d/%d] (attempting to ADD %d): %v",
				i, len(reply.ReleaseProfiles), id, err)

			newID, err2 := app.AddReleaseProfile(profile)
			if err2 != nil {
				return fmt.Errorf("[%d/%d] updating release profiles: %d: (update) %v, (add) %w",
					i, len(reply.ReleaseProfiles), id, err, err2)
			}

			maps.RP = append(maps.RP, idMap{id, newID})
		}
	}

	for i, profile := range reply.QualityProfiles {
		if err := app.UpdateQualityProfile(profile); err != nil {
			id := profile.ID
			profile.ID = 0

			c.Debugf("Error Updating quality format [%d/%d] (attempting to ADD %d): %v",
				i, len(reply.QualityProfiles), id, err)

			newID, err2 := app.AddQualityProfile(profile)
			if err2 != nil {
				return fmt.Errorf("[%d/%d] updating quality profile: %d: (update) %v, (add) %w",
					i, len(reply.QualityProfiles), id, err, err2)
			}

			maps.QP = append(maps.QP, idMap{id, newID})
		}
	}

	return c.postbackSonarrRP(instance, maps)
}

// postbackSonarrRP sends the changes back to notifiarr.com.
func (c *Config) postbackSonarrRP(instance int, maps *cfMapIDpayload) error {
	if len(maps.QP) < 1 && len(maps.RP) < 1 {
		return nil
	}

	//nolint:bodyclose // already closed
	resp, _, err := c.SendData(c.BaseURL+CFSyncRoute+"?app=sonarr&updateIDs=true", &SonarrCustomFormatPayload{
		Instance: instance,
		NewMaps:  maps,
	}, false)
	if err != nil {
		c.sonarrRP[instance] = maps
		return fmt.Errorf("updating quality release ID map: %w", err)
	} else if resp.StatusCode != http.StatusOK {
		c.sonarrRP[instance] = maps
		return fmt.Errorf("updating quality release ID map: %w: %s", ErrNon200, resp.Status)
	}

	delete(c.sonarrRP, instance)

	return nil
}
