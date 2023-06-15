// Package apps provides the _incoming_ HTTP methods for notifiarr.com integrations.
// Methods are included for Radarr, Readrr, Lidarr and Sonarr. This library also
// holds the site API Key and the base HTTP server abstraction used throughout
// the Notifiarr client application. The configuration should be derived from
// a config file; a Router and an Error Log logger must also be provided.
package apps

import (
	"fmt"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/gorilla/mux"
	"golift.io/cnfg"
	"golift.io/starr"
)

// Apps is the input configuration to relay requests to Starr apps.
//
//nolint:lll
type Apps struct {
	APIKey       string            `json:"apiKey" toml:"api_key" xml:"api_key" yaml:"apiKey"`
	ExKeys       []string          `json:"extraKeys" toml:"extra_keys" xml:"extra_keys" yaml:"extraKeys"`
	URLBase      string            `json:"urlbase" toml:"urlbase" xml:"urlbase" yaml:"urlbase"`
	MaxBody      int               `toml:"max_body" xml:"max_body" json:"maxBody"`
	Serial       bool              `json:"serial" toml:"serial" xml:"serial" yaml:"serial"`
	Sonarr       []*SonarrConfig   `json:"sonarr,omitempty" toml:"sonarr" xml:"sonarr" yaml:"sonarr,omitempty"`
	Radarr       []*RadarrConfig   `json:"radarr,omitempty" toml:"radarr" xml:"radarr" yaml:"radarr,omitempty"`
	Lidarr       []*LidarrConfig   `json:"lidarr,omitempty" toml:"lidarr" xml:"lidarr" yaml:"lidarr,omitempty"`
	Readarr      []*ReadarrConfig  `json:"readarr,omitempty" toml:"readarr" xml:"readarr" yaml:"readarr,omitempty"`
	Prowlarr     []*ProwlarrConfig `json:"prowlarr,omitempty" toml:"prowlarr" xml:"prowlarr" yaml:"prowlarr,omitempty"`
	Deluge       []*DelugeConfig   `json:"deluge,omitempty" toml:"deluge" xml:"deluge" yaml:"deluge,omitempty"`
	Qbit         []*QbitConfig     `json:"qbit,omitempty" toml:"qbit" xml:"qbit" yaml:"qbit,omitempty"`
	Rtorrent     []*RtorrentConfig `json:"rtorrent,omitempty" toml:"rtorrent" xml:"rtorrent" yaml:"rtorrent,omitempty"`
	SabNZB       []*SabNZBConfig   `json:"sabnzbd,omitempty" toml:"sabnzbd" xml:"sabnzbd" yaml:"sabnzbd,omitempty"`
	NZBGet       []*NZBGetConfig   `json:"nzbget,omitempty" toml:"nzbget" xml:"nzbget" yaml:"nzbget,omitempty"`
	Transmission []*XmissionConfig `json:"transmission,omitempty" toml:"transmission" xml:"transmission" yaml:"transmission,omitempty"`
	Tautulli     *TautulliConfig   `json:"tautulli,omitempty" toml:"tautulli" xml:"tautulli" yaml:"tautulli,omitempty"`
	Plex         *PlexConfig       `json:"plex" toml:"plex" xml:"plex" yaml:"plex"`
	Router       *mux.Router       `json:"-" toml:"-" xml:"-" yaml:"-"`
	mnd.Logger   `toml:"-" xml:"-" json:"-"`
	keys         map[string]struct{} `toml:"-"` // for fast key lookup.
}

type ExtraConfig struct {
	Name     string        `toml:"name" xml:"name" json:"name"`
	Timeout  cnfg.Duration `toml:"timeout" xml:"timeout" json:"timeout"`
	Interval cnfg.Duration `toml:"interval" xml:"interval" json:"interval"`
	ValidSSL bool          `toml:"valid_ssl" xml:"valid_ssl" json:"validSsl"`
}

// Errors sent to client web requests.
var (
	ErrNoTMDB     = fmt.Errorf("TMDB ID must not be empty")
	ErrNoGRID     = fmt.Errorf("GRID ID must not be empty")
	ErrNoTVDB     = fmt.Errorf("TVDB ID must not be empty")
	ErrNoMBID     = fmt.Errorf("MBID ID must not be empty")
	ErrNoRadarr   = fmt.Errorf("configured %s ID not found", starr.Radarr)
	ErrNoSonarr   = fmt.Errorf("configured %s ID not found", starr.Sonarr)
	ErrNoLidarr   = fmt.Errorf("configured %s ID not found", starr.Lidarr)
	ErrNoReadarr  = fmt.Errorf("configured %s ID not found", starr.Readarr)
	ErrNoProwlarr = fmt.Errorf("configured %s ID not found", starr.Prowlarr)
	ErrNotFound   = fmt.Errorf("the request returned an empty payload")
	ErrNonZeroID  = fmt.Errorf("provided ID must be non-zero")
	// ErrWrongCount is returned when an app returns the wrong item count.
	ErrWrongCount = fmt.Errorf("wrong item count returned")
	ErrInvalidApp = fmt.Errorf("invalid application configuration provided")
)

// Setup creates request interfaces and sets the timeout for each server.
// This is part of the config/startup init.
func (a *Apps) Setup() error { //nolint:cyclop
	a.APIKey = strings.TrimSpace(a.APIKey)

	if err := a.setupLidarr(); err != nil {
		return err
	}

	if err := a.setupProwlarr(); err != nil {
		return err
	}

	if err := a.setupRadarr(); err != nil {
		return err
	}

	if err := a.setupReadarr(); err != nil {
		return err
	}

	if err := a.setupSonarr(); err != nil {
		return err
	}

	if err := a.setupDeluge(); err != nil {
		return err
	}

	if err := a.setupNZBGet(); err != nil {
		return err
	}

	if err := a.setupQbit(); err != nil {
		return err
	}

	if err := a.setupSabNZBd(); err != nil {
		return err
	}

	if err := a.setupRtorrent(); err != nil {
		return err
	}

	if err := a.setupTransmission(); err != nil {
		return err
	}

	if a.Tautulli == nil {
		a.Tautulli = &TautulliConfig{}
	}

	if a.Plex == nil {
		a.Plex = &PlexConfig{Config: &plex.Config{}}
	}

	a.Tautulli.Setup(a.MaxBody, a.Logger)
	a.Plex.Setup(a.MaxBody, a.Logger)

	return nil
}

// InitHandlers activates all our handlers. This is part of the web server init.
func (a *Apps) InitHandlers() {
	a.keys = make(map[string]struct{})
	for _, key := range append(a.ExKeys, a.APIKey) {
		if len(key) > 3 { //nolint:gomnd
			a.keys[key] = struct{}{}
		}
	}

	a.lidarrHandlers()
	a.prowlarrHandlers()
	a.radarrHandlers()
	a.readarrHandlers()
	a.sonarrHandlers()
}
