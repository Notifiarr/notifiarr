package notifiarr

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"golift.io/starr/radarr"
	"golift.io/starr/sonarr"
)

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
	if ci, err := c.GetClientInfo(); err != nil {
		c.Debugf("Cannot sync Radarr Custom Formats. Error: %v", err)
		return
	} else if ci.Message.CFSync < 1 {
		return
	}

	for i, r := range c.Apps.Radarr {
		if r.DisableCF || r.URL == "" || r.APIKey == "" {
			continue
		}

		switch synced, err := c.syncRadarrCF(i+1, r); {
		case err != nil:
			c.Errorf("Radarr Custom Formats sync request for '%d:%s' failed: %v", i+1, r.URL, err)
		case synced:
			c.Printf("Synced Custom Formats from Notifiarr for Radarr: %d:%s", i+1, r.URL)
		default:
			c.Printf("Sent Custom Formats sync request to Notifiarr for Radarr: %d:%s", i+1, r.URL)
		}
	}
}

func (c *Config) syncRadarrCF(instance int, r *apps.RadarrConfig) (bool, error) {
	var (
		err     error
		payload = RadarrCustomFormatPayload{Instance: instance, Name: r.Name, NewMaps: c.radarrCF[instance]}
	)

	payload.QualityProfiles, err = r.Radarr.GetQualityProfiles()
	if err != nil {
		return false, fmt.Errorf("getting quality profiles: %w", err)
	}

	payload.CustomFormats, err = r.Radarr.GetCustomFormats()
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
	} else if err := c.updateRadarrCF(instance, r, body); err != nil {
		return false, fmt.Errorf("updating application: %w", err)
	}

	return true, nil
}

func (c *Config) updateRadarrCF(instance int, r *apps.RadarrConfig, data []byte) error {
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
		len(reply.QualityProfiles), len(reply.CustomFormats), instance, r.URL)

	if payload.Response != success {
		return fmt.Errorf("%w: %s", ErrInvalidResponse, payload.Response)
	}

	maps := &cfMapIDpayload{QP: []idMap{}, CF: []idMap{}, Instance: instance}

	for i, profile := range reply.CustomFormats {
		id := profile.ID
		if _, err := r.UpdateCustomFormat(profile, id); err != nil {
			profile.ID = 0

			c.Debugf("Error Updating custom format [%d/%d] (attempting to ADD %d): %v",
				i, len(reply.CustomFormats), id, err)

			newID, err2 := r.AddCustomFormat(profile)
			if err2 != nil {
				return fmt.Errorf("[%d/%d] updating custom format: %d: (update) %v, (add) %w",
					i, len(reply.CustomFormats), id, err, err2)
			}

			maps.CF = append(maps.CF, idMap{int64(id), int64(newID.ID)})
		}
	}

	for i, profile := range reply.QualityProfiles {
		if err := r.UpdateQualityProfile(profile); err != nil {
			id := profile.ID
			profile.ID = 0

			c.Debugf("Error Updating quality profile [%d/%d] (attempting to ADD %d): %v",
				i, len(reply.QualityProfiles), id, err)

			newID, err2 := r.AddQualityProfile(profile)
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
	if ci, err := c.GetClientInfo(); err != nil {
		c.Debugf("Cannot sync Sonarr Release Profiles. Error: %v", err)
		return
	} else if ci.Message.RPSync < 1 {
		return
	}

	for i, s := range c.Apps.Sonarr {
		if s.DisableCF || s.URL == "" || s.APIKey == "" {
			continue
		}

		switch synced, err := c.syncSonarrRP(i+1, s); {
		case err != nil:
			c.Errorf("Sonarr Release Profiles sync for '%d:%s' failed: %v", i+1, s.URL, err)
		case synced:
			c.Printf("Synced Release Profiles from Notifiarr for Sonarr: %d:%s", i+1, s.URL)
		default:
			c.Printf("Sent Release Profiles sync request to Notifiarr for Sonarr: %d:%s", i+1, s.URL)
		}
	}
}

func (c *Config) syncSonarrRP(instance int, s *apps.SonarrConfig) (bool, error) {
	var (
		err     error
		payload = SonarrCustomFormatPayload{Instance: instance, Name: s.Name, NewMaps: c.sonarrRP[instance]}
	)

	payload.QualityProfiles, err = s.Sonarr.GetQualityProfiles()
	if err != nil {
		return false, fmt.Errorf("getting quality profiles: %w", err)
	}

	payload.ReleaseProfiles, err = s.Sonarr.GetReleaseProfiles()
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
	} else if err := c.updateSonarrRP(instance, s, body); err != nil {
		return false, fmt.Errorf("updating application: %w", err)
	}

	return true, nil
}

func (c *Config) updateSonarrRP(instance int, s *apps.SonarrConfig, data []byte) error {
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
		len(reply.QualityProfiles), len(reply.ReleaseProfiles), instance, s.URL)

	if payload.Response != success {
		return fmt.Errorf("%w: %s", ErrInvalidResponse, payload.Response)
	}

	maps := &cfMapIDpayload{RP: []idMap{}, QP: []idMap{}, Instance: instance}

	for i, profile := range reply.ReleaseProfiles {
		if err := s.UpdateReleaseProfile(profile); err != nil {
			id := profile.ID
			profile.ID = 0

			c.Debugf("Error Updating release profile [%d/%d] (attempting to ADD %d): %v",
				i, len(reply.ReleaseProfiles), id, err)

			newID, err2 := s.AddReleaseProfile(profile)
			if err2 != nil {
				return fmt.Errorf("[%d/%d] updating release profiles: %d: (update) %v, (add) %w",
					i, len(reply.ReleaseProfiles), id, err, err2)
			}

			maps.RP = append(maps.RP, idMap{id, newID})
		}
	}

	for i, profile := range reply.QualityProfiles {
		if err := s.UpdateQualityProfile(profile); err != nil {
			id := profile.ID
			profile.ID = 0

			c.Debugf("Error Updating quality format [%d/%d] (attempting to ADD %d): %v",
				i, len(reply.QualityProfiles), id, err)

			newID, err2 := s.AddQualityProfile(profile)
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
