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
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/klauspost/compress/gzhttp"
	"golang.org/x/time/rate"
	"golift.io/cnfg"
	"golift.io/starr"
)

// AppsConfig is the input configuration to relay requests to Starr apps.
//
//nolint:lll
type AppsConfig struct {
	BaseConfig
	Sonarr       []StarrConfig    `json:"sonarr,omitempty"       toml:"sonarr"       xml:"sonarr"       yaml:"sonarr,omitempty"`
	Radarr       []StarrConfig    `json:"radarr,omitempty"       toml:"radarr"       xml:"radarr"       yaml:"radarr,omitempty"`
	Lidarr       []StarrConfig    `json:"lidarr,omitempty"       toml:"lidarr"       xml:"lidarr"       yaml:"lidarr,omitempty"`
	Readarr      []StarrConfig    `json:"readarr,omitempty"      toml:"readarr"      xml:"readarr"      yaml:"readarr,omitempty"`
	Prowlarr     []StarrConfig    `json:"prowlarr,omitempty"     toml:"prowlarr"     xml:"prowlarr"     yaml:"prowlarr,omitempty"`
	Deluge       []DelugeConfig   `json:"deluge,omitempty"       toml:"deluge"       xml:"deluge"       yaml:"deluge,omitempty"`
	Qbit         []QbitConfig     `json:"qbit,omitempty"         toml:"qbit"         xml:"qbit"         yaml:"qbit,omitempty"`
	Rtorrent     []RtorrentConfig `json:"rtorrent,omitempty"     toml:"rtorrent"     xml:"rtorrent"     yaml:"rtorrent,omitempty"`
	SabNZB       []SabNZBConfig   `json:"sabnzbd,omitempty"      toml:"sabnzbd"      xml:"sabnzbd"      yaml:"sabnzbd,omitempty"`
	NZBGet       []NZBGetConfig   `json:"nzbget,omitempty"       toml:"nzbget"       xml:"nzbget"       yaml:"nzbget,omitempty"`
	Transmission []XmissionConfig `json:"transmission,omitempty" toml:"transmission" xml:"transmission" yaml:"transmission,omitempty"`
	Tautulli     TautulliConfig   `json:"tautulli"               toml:"tautulli"     xml:"tautulli"     yaml:"tautulli"`
	Plex         PlexConfig       `json:"plex"                   toml:"plex"         xml:"plex"         yaml:"plex"`
	PlexServer   []PlexConfig     `json:"plexServer,omitempty"   toml:"plex_server"  xml:"plex_server"  yaml:"plexServer,omitempty"`
}

type BaseConfig struct {
	APIKey  string   `json:"apiKey"    toml:"api_key"    xml:"api_key"    yaml:"apiKey"`
	ExKeys  []string `json:"extraKeys" toml:"extra_keys" xml:"extra_keys" yaml:"extraKeys"`
	URLBase string   `json:"urlbase"   toml:"urlbase"    xml:"urlbase"    yaml:"urlbase"`
	MaxBody int      `json:"maxBody"   toml:"max_body"   xml:"max_body"   yaml:"maxBody"`
	Serial  bool     `json:"serial"    toml:"serial"     xml:"serial"     yaml:"serial"`
}

type Apps struct {
	BaseConfig
	Lidarr       []Lidarr
	Radarr       []Radarr
	Readarr      []Readarr
	Sonarr       []Sonarr
	Prowlarr     []Prowlarr
	Deluge       []Deluge
	NZBGet       []NZBGet
	Qbit         []Qbit
	Rtorrent     []Rtorrent
	SabNZB       []SabNZB
	Transmission []Xmission
	Tautulli     Tautulli
	Plex         []Plex
	Router       *mux.Router
	keys         map[string]struct{} // for fast key lookup.
	compress     func(h http.Handler) http.HandlerFunc
}

type ExtraConfig struct {
	Name     string        `json:"name"     toml:"name"      xml:"name"`
	Timeout  cnfg.Duration `json:"timeout"  toml:"timeout"   xml:"timeout"`
	Interval cnfg.Duration `json:"interval" toml:"interval"  xml:"interval"`
	ValidSSL bool          `json:"validSsl" toml:"valid_ssl" xml:"valid_ssl"`
	Deletes  int           `json:"deletes"  toml:"deletes"   xml:"deletes"`
}

type StarrConfig struct {
	starr.Config
	ExtraConfig
}

type StarrApp struct {
	StarrConfig
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

// CheckURLs validates the configuration for each app.
func CheckURLs(config *AppsConfig) error { //nolint:cyclop,gocognit,funlen
	for idx, app := range config.Lidarr {
		if err := checkUrl(app.URL, starr.Lidarr.String(), idx); err != nil {
			return err
		}
	}

	for idx, app := range config.Prowlarr {
		if err := checkUrl(app.URL, starr.Prowlarr.String(), idx); err != nil {
			return err
		}
	}

	for idx, app := range config.Radarr {
		if err := checkUrl(app.URL, starr.Radarr.String(), idx); err != nil {
			return err
		}
	}

	for idx, app := range config.Readarr {
		if err := checkUrl(app.URL, starr.Readarr.String(), idx); err != nil {
			return err
		}
	}

	for idx, app := range config.Sonarr {
		if err := checkUrl(app.URL, starr.Sonarr.String(), idx); err != nil {
			return err
		}
	}

	for idx, app := range config.Deluge {
		if err := checkUrl(app.URL, "Deluge", idx); err != nil {
			return err
		}
	}

	for idx, app := range config.Rtorrent {
		if err := checkUrl(app.URL, "Rtorrent", idx); err != nil {
			return err
		}
	}

	for idx, app := range config.SabNZB {
		if err := checkUrl(app.URL, "SabNZB", idx); err != nil {
			return err
		}
	}

	for idx, app := range config.NZBGet {
		if err := checkUrl(app.URL, "NZBGet", idx); err != nil {
			return err
		}
	}

	for idx, app := range config.Qbit {
		if err := checkUrl(app.URL, "Qbit", idx); err != nil {
			return err
		}
	}

	for idx, app := range config.Transmission {
		if err := checkUrl(app.URL, "Transmission", idx); err != nil {
			return err
		}
	}

	if config.Tautulli.URL != "" {
		if err := checkUrl(config.Tautulli.URL, "Tautulli", 0); err != nil {
			return err
		}
	}

	// Check primary Plex config (backward compat).
	if config.Plex.URL != "" {
		if err := checkUrl(config.Plex.URL, "Plex", 0); err != nil {
			return err
		}
	}

	// Check additional Plex servers.
	for idx, app := range config.PlexServer {
		if err := checkUrl(app.URL, "Plex", idx+1); err != nil {
			return err
		}
	}

	return nil
}

// New creates request interfaces and sets the timeout for each server.
// This is part of the config/startup init.
func New(config *AppsConfig) (*Apps, error) { //nolint:cyclop,funlen
	var err error

	apps := &Apps{
		BaseConfig: config.BaseConfig,
	}

	apps.APIKey = strings.TrimSpace(apps.APIKey)
	apps.compress = gzhttp.GzipHandler

	apps.Lidarr, err = config.setupLidarr()
	if err != nil {
		return nil, err
	}

	apps.Prowlarr, err = config.setupProwlarr()
	if err != nil {
		return nil, err
	}

	apps.Radarr, err = config.setupRadarr()
	if err != nil {
		return nil, err
	}

	apps.Readarr, err = config.setupReadarr()
	if err != nil {
		return nil, err
	}

	apps.Sonarr, err = config.setupSonarr()
	if err != nil {
		return nil, err
	}

	apps.Deluge, err = config.setupDeluge()
	if err != nil {
		return nil, err
	}

	apps.NZBGet, err = config.setupNZBGet()
	if err != nil {
		return nil, err
	}

	apps.Qbit, err = config.setupQbit()
	if err != nil {
		return nil, err
	}

	apps.SabNZB, err = config.setupSabNZBd()
	if err != nil {
		return nil, err
	}

	apps.Rtorrent, err = config.setupRtorrent()
	if err != nil {
		return nil, err
	}

	apps.Transmission, err = config.setupTransmission()
	if err != nil {
		return nil, err
	}

	apps.Tautulli = config.Tautulli.Setup(apps.MaxBody)
	apps.Plex = config.setupPlex()

	return apps, nil
}

// setupPlex merges the primary Plex config with any additional plex_server configs.
func (a *AppsConfig) setupPlex() []Plex {
	var instances []Plex

	// Add primary Plex instance if configured (backward compat).
	if a.Plex.Enabled() {
		instances = append(instances, a.Plex.Setup(a.MaxBody))
	}

	// Add additional Plex servers.
	for _, cfg := range a.PlexServer {
		if cfg.Enabled() {
			instances = append(instances, cfg.Setup(a.MaxBody))
		}
	}

	return instances
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
func (e StarrApp) DelOK() bool {
	return e.Deletes > 0 && e.delLimit.Allow()
}

// Enabled returns true if the Sonarr instance is enabled and usable.
func (s StarrConfig) Enabled() bool {
	return s.URL != "" && s.APIKey != "" && s.Timeout.Duration >= 0
}

// apiError attempts to parse the api error message to create a better return code.
func apiError(backupCode int, msg string, err error) (int, error) {
	var reqerr *starr.ReqError

	if errors.As(err, &reqerr) {
		return reqerr.Code, fmt.Errorf("%s: %w", msg, reqerr)
	}

	return backupCode, fmt.Errorf("%s: %w", msg, err)
}

func checkUrl(appURL, app string, idx int) error {
	if appURL == "" {
		return fmt.Errorf("%w: missing url: %s config %d", ErrInvalidApp, app, idx+1)
	} else if !strings.HasPrefix(appURL, "http://") && !strings.HasPrefix(appURL, "https://") {
		return fmt.Errorf("%w: URL must begin with http:// or https://: %s config %d", ErrInvalidApp, app, idx+1)
	}

	if _, err := url.Parse(appURL); err != nil {
		return fmt.Errorf("%w: invalid URL: %s config %d", err, app, idx+1)
	}

	return nil
}
