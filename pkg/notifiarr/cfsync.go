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
	CustomFormats   []*radarr.CustomFormat   `json:"customFormats,omitempty"`
	QualityProfiles []*radarr.QualityProfile `json:"qualityProfiles,omitempty"`
	NewMaps         *cfMapIDpayload          `json:"newMaps,omitempty"`
}

// SyncRadarrCF triggers a custom format sync for Radarr.
func (c *Config) SyncRadarrCF() {
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
		payload = RadarrCustomFormatPayload{Instance: instance, NewMaps: c.radarrCFs[instance]}
	)

	payload.QualityProfiles, err = r.Radarr.GetQualityProfiles()
	if err != nil {
		return false, fmt.Errorf("getting quality profiles: %w", err)
	}

	payload.CustomFormats, err = r.Radarr.GetCustomFormats()
	if err != nil {
		return false, fmt.Errorf("getting custom formats: %w", err)
	}

	resp, _, body, err := c.SendData(c.BaseURL+CFSyncRoute+"?app=radarr", payload) //nolint:bodyclose // already closed

	switch {
	case err != nil:
		return false, fmt.Errorf("sending current formats: %w", err)
	case resp.StatusCode != http.StatusOK:
		return false, fmt.Errorf("%w: %s: %s", ErrNon200, resp.Status, string(body))
	default:
		c.Debugf("Response Payload: %s: %s", resp.Status, string(body))
	}

	c.radarrCFs[instance] = nil

	if len(body) < 1 {
		return false, nil
	} else if err := c.updateRadarrCFs(instance, r, body); err != nil {
		return false, fmt.Errorf("updating application: %w", err)
	}

	return true, nil
}

func (c *Config) updateRadarrCFs(instance int, r *apps.RadarrConfig, data []byte) error {
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

	for i, cf := range reply.CustomFormats {
		if _, err := r.UpdateCustomFormat(cf, cf.ID); err != nil {
			id := cf.ID
			cf.ID = 0

			newID, err2 := r.AddCustomFormat(cf)
			if err2 != nil {
				return fmt.Errorf("[%d/%d] updating custom format: %d: (update) %v, (add) %w",
					i, len(reply.CustomFormats), id, err, err2)
			}

			maps.CF = append(maps.CF, idMap{int64(id), int64(newID.ID)})
		}
	}

	for i, qp := range reply.QualityProfiles {
		if err := r.UpdateQualityProfile(qp); err != nil {
			id := qp.ID
			qp.ID = 0

			newID, err2 := r.AddQualityProfile(qp)
			if err2 != nil {
				return fmt.Errorf("[%d/%d] updating quality profile: %d: (update) %v, (add) %w",
					i, len(reply.QualityProfiles), id, err, err2)
			}

			maps.QP = append(maps.QP, idMap{id, newID})
		}
	}

	return c.postbackRadarrCFs(instance, maps)
}

func (c *Config) postbackRadarrCFs(instance int, maps *cfMapIDpayload) error {
	if len(maps.QP) > 0 || len(maps.CF) > 0 {
		//nolint:bodyclose // already closed.
		resp, _, body, err := c.SendData(c.BaseURL+CFSyncRoute+"?app=radarr&updateIDs=true", &RadarrCustomFormatPayload{
			Instance: instance,
			NewMaps:  maps,
		})
		if err != nil {
			c.radarrCFs[instance] = maps
			return fmt.Errorf("updating custom format ID map: %w: %s", err, string(body))
		} else if resp.StatusCode != http.StatusOK {
			c.radarrCFs[instance] = maps
			return fmt.Errorf("updating custom format ID map: %w: %s: %s", ErrNon200, resp.Status, string(body))
		}
	}

	return nil
}

/*//*****/ // Sonarr /*//*****///

// SonarrCustomFormatPayload is the payload sent and received
// to/from notifarr.com when updating custom formats for Sonarr.
type SonarrCustomFormatPayload struct {
	Instance        int                      `json:"instance"`
	ReleaseProfiles []*sonarr.ReleaseProfile `json:"releaseProfiles,omitempty"`
	QualityProfiles []*sonarr.QualityProfile `json:"qualityProfiles,omitempty"`
	NewMaps         *cfMapIDpayload          `json:"newMaps,omitempty"`
}

// SyncSonarrCF triggers a custom format sync for Sonarr.
func (c *Config) SyncSonarrCF() {
	for i, s := range c.Apps.Sonarr {
		if s.DisableCF || s.URL == "" || s.APIKey == "" {
			continue
		}

		switch synced, err := c.syncSonarrCF(i+1, s); {
		case err != nil:
			c.Errorf("Sonarr Release Profiles sync for '%d:%s' failed: %v", i+1, s.URL, err)
		case synced:
			c.Printf("Synced Release Profiles from Notifiarr for Sonarr: %d:%s", i+1, s.URL)
		default:
			c.Printf("Sent Release Profiles sync request to Notifiarr for Sonarr: %d:%s", i+1, s.URL)
		}
	}
}

func (c *Config) syncSonarrCF(instance int, s *apps.SonarrConfig) (bool, error) {
	var (
		err     error
		payload = SonarrCustomFormatPayload{Instance: instance, NewMaps: c.sonarrCFs[instance]}
	)

	payload.QualityProfiles, err = s.Sonarr.GetQualityProfiles()
	if err != nil {
		return false, fmt.Errorf("getting quality profiles: %w", err)
	}

	payload.ReleaseProfiles, err = s.Sonarr.GetReleaseProfiles()
	if err != nil {
		return false, fmt.Errorf("getting release profiles: %w", err)
	}

	resp, _, body, err := c.SendData(c.BaseURL+CFSyncRoute+"?app=sonarr", payload) //nolint:bodyclose // already closed

	switch {
	case err != nil:
		return false, fmt.Errorf("sending current profiles: %w", err)
	case resp.StatusCode != http.StatusOK:
		return false, fmt.Errorf("%w: %s: %s", ErrNon200, resp.Status, string(body))
	default:
		c.Debugf("Response Payload: %s: %s", resp.Status, string(body))
	}

	c.sonarrCFs[instance] = nil

	if len(body) < 1 {
		return false, nil
	} else if err := c.updateSonarrCFs(instance, s, body); err != nil {
		return false, fmt.Errorf("updating application: %w", err)
	}

	return true, nil
}

func (c *Config) updateSonarrCFs(instance int, s *apps.SonarrConfig, data []byte) error {
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

	maps := &cfMapIDpayload{QP: []idMap{}, RP: []idMap{}, Instance: instance}

	for i, cf := range reply.ReleaseProfiles {
		if err := s.UpdateReleaseProfile(cf); err != nil {
			newID, err2 := s.AddReleaseProfile(cf)
			if err2 != nil {
				return fmt.Errorf("[%d/%d] updating release profiles: %d: (update) %v, (add) %w",
					i, len(reply.ReleaseProfiles), cf.ID, err, err2)
			}

			maps.RP = append(maps.RP, idMap{cf.ID, newID})
		}
	}

	for i, qp := range reply.QualityProfiles {
		if err := s.UpdateQualityProfile(qp); err != nil {
			id := qp.ID
			qp.ID = 0

			newID, err2 := s.AddQualityProfile(qp)
			if err2 != nil {
				return fmt.Errorf("[%d/%d] updating quality profile: %d: (update) %v, (add) %w",
					i, len(reply.QualityProfiles), id, err, err2)
			}

			maps.QP = append(maps.QP, idMap{id, newID})
		}
	}

	return c.postbackSonarrCFs(instance, maps)
}

func (c *Config) postbackSonarrCFs(instance int, maps *cfMapIDpayload) error {
	if len(maps.QP) > 0 || len(maps.RP) > 0 {
		//nolint:bodyclose // already closed
		resp, _, body, err := c.SendData(c.BaseURL+CFSyncRoute+"?app=sonarr&updateIDs=true", &SonarrCustomFormatPayload{
			Instance: instance,
			NewMaps:  maps,
		})

		if err != nil {
			c.sonarrCFs[instance] = maps
			return fmt.Errorf("updating release profiles ID map: %w: %s", err, string(body))
		} else if resp.StatusCode != http.StatusOK {
			c.sonarrCFs[instance] = maps
			return fmt.Errorf("updating custom format ID map: %w: %s: %s", ErrNon200, resp.Status, string(body))
		}
	}

	return nil
}
