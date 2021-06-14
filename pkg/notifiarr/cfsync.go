package notifiarr

import (
	"encoding/json"
	"fmt"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"golift.io/starr/radarr"
	"golift.io/starr/sonarr"
)

const (
	// CFSyncRoute is the webserver route to send sync requests to.
	CFSyncRoute = "/api/v1/user/trash"
)

/*//*****/ // Radarr /*//*****///

// RadarrCustomFormatPayload is the payload sent and received
// to/from notifarr.com when updating custom formats for Radarr.
type RadarrCustomFormatPayload struct {
	App             string
	CustomFormats   []*radarr.CustomFormat   `json:"customFormats"`
	QualityProfiles []*radarr.QualityProfile `json:"qualityProfiles"`
}

// SyncRadarrCF triggers a custom format sync for Radarr.
func (c *Config) SyncRadarrCF() {
	for _, r := range c.Apps.Radarr {
		if r.DisableCF || r.URL == "" || r.APIKey == "" {
			continue
		}

		switch synced, err := c.syncRadarrCF(r); {
		case err != nil:
			c.Errorf("Radarr CF sync request for '%s' failed: %v", r.URL, err)
		case synced:
			c.Printf("Sent Custom Format sync request to Notifiarr for Radarr: %s", r.URL)
		default:
			c.Printf("Updated Custom Formats from Notifiarr for Radarr: %s", r.URL)
		}
	}
}

func (c *Config) syncRadarrCF(r *apps.RadarrConfig) (bool, error) {
	var (
		err     error
		payload = RadarrCustomFormatPayload{App: "radarr"}
	)

	payload.QualityProfiles, err = r.Radarr.GetQualityProfiles()
	if err != nil {
		return false, fmt.Errorf("getting quality profiles: %w", err)
	}

	payload.CustomFormats, err = r.Radarr.GetCustomFormats()
	if err != nil {
		return false, fmt.Errorf("getting custom formats: %w", err)
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("json marshalling: %w", err)
	}

	b, err = c.SendJSON(c.BaseURL+CFSyncRoute, b)
	if err != nil {
		return false, fmt.Errorf("sending current formats: %w", err)
	}

	if len(b) < 1 {
		return false, nil
	}

	if err := c.updateRadarrCFs(r, b); err != nil {
		return false, fmt.Errorf("updating application: %w", err)
	}

	return true, nil
}

func (c *Config) updateRadarrCFs(r *apps.RadarrConfig, data []byte) error {
	var payload RadarrCustomFormatPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("invalid response: %w", err)
	}

	for i, cf := range payload.CustomFormats {
		if _, err := r.UpdateCustomFormat(cf, cf.ID); err != nil {
			return fmt.Errorf("[%d/%d] updating custom format: %d: %w", i, len(payload.CustomFormats), cf.ID, err)
		}
	}

	for i, qp := range payload.QualityProfiles {
		if err := r.UpdateQualityProfile(qp); err != nil {
			return fmt.Errorf("[%d/%d] updating quality profile: %d: %w", i, len(payload.QualityProfiles), qp.ID, err)
		}
	}

	return nil
}

/*//*****/ // Sonarr /*//*****///

// SonarrCustomFormatPayload is the payload sent and received
// to/from notifarr.com when updating custom formats for Sonarr.
type SonarrCustomFormatPayload struct {
	App string
	// CustomFormats   []*sonarr.CustomFormat   `json:"customFormats"`
	QualityProfiles []*sonarr.QualityProfile `json:"qualityProfiles"`
}

// SyncSonarrCF triggers a custom format sync for Sonarr.
func (c *Config) SyncSonarrCF() {
	for _, s := range c.Apps.Sonarr {
		if s.DisableCF || s.URL == "" || s.APIKey == "" {
			continue
		}

		switch synced, err := c.syncSonarrCF(s); {
		case err != nil:
			c.Errorf("Sonarr CF sync for '%s' failed: %v", s.URL, err)
		case synced:
			c.Printf("Sent Custom Format sync request to Notifiarr for Sonarr: %s", s.URL)
		default:
			c.Printf("Updated Custom Formats from Notifiarr for Sonarr: %s", s.URL)
		}
	}
}

func (c *Config) syncSonarrCF(s *apps.SonarrConfig) (bool, error) {
	var (
		err     error
		payload = SonarrCustomFormatPayload{App: "sonarr"}
	)

	payload.QualityProfiles, err = s.Sonarr.GetQualityProfiles()
	if err != nil {
		return false, fmt.Errorf("getting quality profiles: %w", err)
	}

	/* // Sonarr has no Custom Formats (yet?)
	payload.CustomFormats, err = s.Sonarr.GetCustomFormats()
	if err != nil {
		return false, fmt.Errorf("getting custom formats: %w", err)
	} /**/

	b, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("json marshalling: %w", err)
	}

	b, err = c.SendJSON(c.BaseURL+CFSyncRoute, b)
	if err != nil {
		return false, fmt.Errorf("sending current formats: %w", err)
	}

	if len(b) < 1 {
		return false, nil
	}

	if err := c.updateSonarrCFs(s, b); err != nil {
		return false, fmt.Errorf("updating application: %w", err)
	}

	return true, nil
}

func (c *Config) updateSonarrCFs(s *apps.SonarrConfig, data []byte) error {
	var payload SonarrCustomFormatPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("invalid response: %w", err)
	}

	/* // Sonarr has no Custom Formats
	for _, cf := range payload.CustomFormats {
		if _, err := s.UpdateCustomFormat(cf, cf.ID); err != nil {
			return fmt.Errorf("updating custom format: %d: %w", cf.ID, err)
		}
	} /**/

	for i, qp := range payload.QualityProfiles {
		if err := s.UpdateQualityProfile(qp); err != nil {
			return fmt.Errorf("[%d/%d] updating quality profile: %d: %w", i, len(payload.QualityProfiles), qp.ID, err)
		}
	}

	return nil
}
