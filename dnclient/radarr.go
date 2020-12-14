package dnclient

import (
	"strconv"
	"strings"
	"sync"

	"golift.io/starr"
)

// RadarrConfig represents the input data for a Radarr server.
type RadarrConfig struct {
	//	Name      string `json:"name" toml:"name" xml:"name" yaml:"name"`
	Root      string `json:"root_folder" toml:"root_folder" xml:"root_folder" yaml:"root_folder"`
	QualityID int    `json:"quality_id" toml:"quality_id" xml:"quality_id" yaml:"quality_id"`
	ProfileID int    `json:"profile_id" toml:"profile_id" xml:"profile_id" yaml:"profile_id"`
	Search    bool   `json:"search" toml:"search" xml:"search" yaml:"search"`
	*starr.Config
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

func (c *Config) fixRadarrConfig() {
	for i := range c.Radarr {
		if c.Radarr[i].Timeout.Duration == 0 {
			c.Radarr[i].Timeout.Duration = c.Timeout.Duration
		}

		if c.Radarr[i].QualityID == 0 {
			c.Radarr[i].QualityID = 1
		}
	}
}

func (c *Client) logRadarr() {
	if count := len(c.Radarr); count == 1 {
		c.Printf(" => Radarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v, profileId:%d, qualityId:%d, search:%v, root:%s",
			c.Radarr[0].URL, c.Radarr[0].APIKey != "", c.Radarr[0].Timeout,
			c.Radarr[0].ValidSSL, c.Radarr[0].ProfileID, c.Radarr[0].QualityID, c.Radarr[0].Search, c.Radarr[0].Root)
	} else {
		c.Print(" => Radarr Config:", count, "servers")

		for _, f := range c.Radarr {
			c.Printf(" =>    Server: %s, apikey:%v, timeout:%v, verify ssl:%v, profileId:%d, qualityId:%d, search:%v, root:%s",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, f.ProfileID, f.QualityID, f.Search, f.Root)
		}
	}
}

func (c *Client) handleRadarr(p *IncomingPayload) (string, error) {
	if p.TMDB == 0 {
		return "TMDB ID must not be empty", nil
	} else if p.Title == "" {
		return "Title must not be empty", nil
	} else if p.Year == 0 {
		return "Year must not be empty", nil
	}

	for i, radar := range c.Radarr {
		if i != p.ID {
			continue
		}

		if p.Root == "" {
			p.Root = radar.Root
		}

		return c.addMovie(p, radar)
	}

	return "configured radarr ID not found", nil
}

func (c *Client) addMovie(p *IncomingPayload, radar *RadarrConfig) (string, error) {
	m, err := radar.Radarr3Movie(p.TMDB)
	if err != nil {
		return "error connecting to Radarr", err
	}

	if len(m) > 0 {
		return "movie already exists", nil
	}

	err = radar.Radarr3AddMovie(&starr.AddMovie{
		Title:               p.Title,
		TitleSlug:           slug(p.Title, p.TMDB),
		TmdbID:              p.TMDB,
		Year:                p.Year,
		Monitored:           true,
		QualityProfileID:    radar.QualityID,
		ProfileID:           radar.ProfileID,
		MinimumAvailability: "released",
		AddMovieOptions:     starr.AddMovieOptions{SearchForMovie: radar.Search},
		RootFolderPath:      p.Root,
	})
	if err != nil {
		return "error adding movie", err
	}

	return "movie added", nil
}

func slug(s string, i int) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "-") + "-" + strconv.Itoa(i)
}
