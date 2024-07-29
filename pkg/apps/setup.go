// Package apps provides the _incoming_ HTTP methods for notifiarr.com integrations.
// Methods are included for Radarr, Readrr, Lidarr and Sonarr. This library also
// holds the site API Key and the base HTTP server abstraction used throughout
// the Notifiarr client application. The configuration should be derived from
// a config file; a Router and an Error Log logger must also be provided.
package apps

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/CAFxX/httpcompression"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/gorilla/mux"
	"golang.org/x/time/rate"
	"golift.io/cnfg"
	"golift.io/starr"
)

// Apps is the input configuration to relay requests to Starr apps.
//
//nolint:lll
type Apps struct {
	APIKey       string            `json:"apiKey"                 toml:"api_key"      xml:"api_key"      yaml:"apiKey"`
	ExKeys       []string          `json:"extraKeys"              toml:"extra_keys"   xml:"extra_keys"   yaml:"extraKeys"`
	URLBase      string            `json:"urlbase"                toml:"urlbase"      xml:"urlbase"      yaml:"urlbase"`
	MaxBody      int               `json:"maxBody"                toml:"max_body"     xml:"max_body"     yaml:"maxBody"`
	Serial       bool              `json:"serial"                 toml:"serial"       xml:"serial"       yaml:"serial"`
	Sonarr       []*SonarrConfig   `json:"sonarr,omitempty"       toml:"sonarr"       xml:"sonarr"       yaml:"sonarr,omitempty"`
	Radarr       []*RadarrConfig   `json:"radarr,omitempty"       toml:"radarr"       xml:"radarr"       yaml:"radarr,omitempty"`
	Lidarr       []*LidarrConfig   `json:"lidarr,omitempty"       toml:"lidarr"       xml:"lidarr"       yaml:"lidarr,omitempty"`
	Readarr      []*ReadarrConfig  `json:"readarr,omitempty"      toml:"readarr"      xml:"readarr"      yaml:"readarr,omitempty"`
	Prowlarr     []*ProwlarrConfig `json:"prowlarr,omitempty"     toml:"prowlarr"     xml:"prowlarr"     yaml:"prowlarr,omitempty"`
	Deluge       []*DelugeConfig   `json:"deluge,omitempty"       toml:"deluge"       xml:"deluge"       yaml:"deluge,omitempty"`
	Qbit         []*QbitConfig     `json:"qbit,omitempty"         toml:"qbit"         xml:"qbit"         yaml:"qbit,omitempty"`
	Rtorrent     []*RtorrentConfig `json:"rtorrent,omitempty"     toml:"rtorrent"     xml:"rtorrent"     yaml:"rtorrent,omitempty"`
	SabNZB       []*SabNZBConfig   `json:"sabnzbd,omitempty"      toml:"sabnzbd"      xml:"sabnzbd"      yaml:"sabnzbd,omitempty"`
	NZBGet       []*NZBGetConfig   `json:"nzbget,omitempty"       toml:"nzbget"       xml:"nzbget"       yaml:"nzbget,omitempty"`
	Transmission []*XmissionConfig `json:"transmission,omitempty" toml:"transmission" xml:"transmission" yaml:"transmission,omitempty"`
	Tautulli     *TautulliConfig   `json:"tautulli,omitempty"     toml:"tautulli"     xml:"tautulli"     yaml:"tautulli,omitempty"`
	Plex         *PlexConfig       `json:"plex"                   toml:"plex"         xml:"plex"         yaml:"plex"`
	Router       *mux.Router       `json:"-"                      toml:"-"            xml:"-"            yaml:"-"`
	mnd.Logger   `json:"-"                      toml:"-"            xml:"-"`
	keys         map[string]struct{} // for fast key lookup.
	compress     func(h http.Handler) http.Handler
}

type ExtraConfig struct {
	Name     string        `json:"name"     toml:"name"      xml:"name"`
	Timeout  cnfg.Duration `json:"timeout"  toml:"timeout"   xml:"timeout"`
	Interval cnfg.Duration `json:"interval" toml:"interval"  xml:"interval"`
	ValidSSL bool          `json:"validSsl" toml:"valid_ssl" xml:"valid_ssl"`
	Deletes  int           `json:"deletes"  toml:"deletes"   xml:"deletes"`
	delLimit *rate.Limiter
}

// Errors sent to client web requests.
var (
	ErrNoTMDB     = errors.New("TMDB ID must not be empty")
	ErrNoGRID     = errors.New("GRID ID must not be empty")
	ErrNoTVDB     = errors.New("TVDB ID must not be empty")
	ErrNoMBID     = errors.New("MBID ID must not be empty")
	ErrNoRadarr   = fmt.Errorf("configured %s ID not found", starr.Radarr)
	ErrNoSonarr   = fmt.Errorf("configured %s ID not found", starr.Sonarr)
	ErrNoLidarr   = fmt.Errorf("configured %s ID not found", starr.Lidarr)
	ErrNoReadarr  = fmt.Errorf("configured %s ID not found", starr.Readarr)
	ErrNoProwlarr = fmt.Errorf("configured %s ID not found", starr.Prowlarr)
	ErrNotFound   = errors.New("the request returned an empty payload")
	ErrNonZeroID  = errors.New("provided ID must be non-zero")
	// ErrWrongCount is returned when an app returns the wrong item count.
	ErrWrongCount = errors.New("wrong item count returned")
	ErrInvalidApp = errors.New("invalid application configuration provided")
	ErrRateLimit  = errors.New("rate limit reached")
)

// Setup creates request interfaces and sets the timeout for each server.
// This is part of the config/startup init.
func (a *Apps) Setup() error { //nolint:cyclop
	a.APIKey = strings.TrimSpace(a.APIKey)
	a.compress, _ = httpcompression.DefaultAdapter()

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
		if len(key) > 3 { //nolint:mnd
			a.keys[key] = struct{}{}
		}
	}

	a.lidarrHandlers()
	a.prowlarrHandlers()
	a.radarrHandlers()
	a.readarrHandlers()
	a.sonarrHandlers()
}

// DelOK returns true if the delete limit isn't reached.
func (e *ExtraConfig) DelOK() bool {
	return e.Deletes > 0 && e.delLimit.Allow()
}

// apiError attempts to parse the api error message to create a better return code.
func apiError(backupCode int, msg string, err error) (int, error) {
	var reqerr *starr.ReqError

	if errors.As(err, &reqerr) {
		return reqerr.Code, fmt.Errorf("%s: %w", msg, reqerr)
	}

	return backupCode, fmt.Errorf("%s: %w", msg, err)
}
