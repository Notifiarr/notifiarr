package dnclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"golift.io/starr"
	"golift.io/starr/sonarr"
)

func (c *Config) fixSonarrConfig() {
	for i := range c.Sonarr {
		if c.Sonarr[i].Timeout.Duration == 0 {
			c.Sonarr[i].Timeout.Duration = c.Timeout.Duration
		}
	}
}

// SonarrConfig represents the input data for a Sonarr server.
type SonarrConfig struct {
	*starr.Config
	*sonarr.Sonarr
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

func (c *Client) logSonarr() {
	if count := len(c.Sonarr); count == 1 {
		c.Printf(" => Sonarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Sonarr[0].URL, c.Sonarr[0].APIKey != "", c.Sonarr[0].Timeout, c.Sonarr[0].ValidSSL)
	} else {
		c.Print(" => Sonarr Config:", count, "servers")

		for _, f := range c.Sonarr {
			c.Printf(" =>    Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}
}

// getSonarr finds a Sonarr based on the passed-in ID.
// Every Sonarr handler calls this.
func (c *Client) getSonarr(id string) *SonarrConfig {
	j, _ := strconv.Atoi(id)

	for i, app := range c.Sonarr {
		if i != j-1 { // discordnotifier wants 1-indexes
			continue
		}

		return app
	}

	return nil
}

func (c *Client) sonarrRootFolders(r *http.Request) (int, interface{}) {
	// Make sure the provided sonarr id exists.
	sonar := c.getSonarr(mux.Vars(r)["id"])
	if sonar == nil {
		return http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", mux.Vars(r)["id"], ErrNoSonarr)
	}

	// Get folder list from Sonarr.
	folders, err := sonar.GetRootFolders()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting folders: %w", err)
	}

	// Format folder list into a nice path=>freesSpace map.
	p := make(map[string]int64)
	for i := range folders {
		p[folders[i].Path] = folders[i].FreeSpace
	}

	return http.StatusOK, p
}

func (c *Client) sonarrProfiles(r *http.Request) (int, interface{}) {
	// Make sure the provided sonarr id exists.
	sonar := c.getSonarr(mux.Vars(r)["id"])
	if sonar == nil {
		return http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", mux.Vars(r)["id"], ErrNoSonarr)
	}

	// Get the profiles from sonarr.
	profiles, err := sonar.GetQualityProfiles()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting profiles: %w", err)
	}

	// Format profile ID=>Name into a nice map.
	p := make(map[int]string)
	for i := range profiles {
		p[profiles[i].ID] = profiles[i].Name
	}

	return http.StatusOK, p
}

func (c *Client) sonarrLangProfiles(r *http.Request) (int, interface{}) {
	// Make sure the provided sonarr id exists.
	sonar := c.getSonarr(mux.Vars(r)["id"])
	if sonar == nil {
		return http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", mux.Vars(r)["id"], ErrNoSonarr)
	}

	// Get the profiles from sonarr.
	profiles, err := sonar.GetLanguageProfiles()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting language profiles: %w", err)
	}

	// Format profile ID=>Name into a nice map.
	p := make(map[int]string)
	for i := range profiles {
		p[profiles[i].ID] = profiles[i].Name
	}

	return http.StatusOK, p
}

func (c *Client) sonarrAddSeries(r *http.Request) (int, interface{}) {
	// Make sure the provided sonarr id exists.
	sonar := c.getSonarr(mux.Vars(r)["id"])
	if sonar == nil {
		return http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", mux.Vars(r)["id"], ErrNoSonarr)
	}

	// Extract payload and check for TMDB ID.
	payload := &sonarr.AddSeriesInput{}
	if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	} else if payload.TvdbID == 0 {
		return http.StatusUnprocessableEntity, fmt.Errorf("0: %w", ErrNoTMDB)
	}

	// Check for existing series.
	if m, err := sonar.GetSeries(payload.TvdbID); err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking series: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, fmt.Errorf("%d: %w", payload.TvdbID, ErrExists)
	}

	series, err := sonar.AddSeries(payload)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding series: %w", err)
	}

	return http.StatusCreated, series
}
