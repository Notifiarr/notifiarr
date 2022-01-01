package apps

import (
	"fmt"
	"strings"
	"time"

	"golift.io/cnfg"
	"golift.io/starr"
	"golift.io/starr/prowlarr"
)

// prowlarrHandlers is called once on startup to register the web API paths.
func (a *Apps) prowlarrHandlers() {
}

// ProwlarrConfig represents the input data for a Prowlarr server.
type ProwlarrConfig struct {
	Name     string        `toml:"name" xml:"name"`         // if set, turn on service checks.
	Interval cnfg.Duration `toml:"interval" xml:"interval"` // service check interval.
	Corrupt  string        `toml:"corrupt" xml:"corrupt"`
	Backup   string        `toml:"backup" xml:"backup"`
	*starr.Config
	*prowlarr.Prowlarr
	Errorf func(string, ...interface{}) `toml:"-" xml:"-"`
}

func (a *Apps) setupProwlarr(timeout time.Duration) error {
	for i, prowl := range a.Prowlarr {
		if prowl.Config == nil || prowl.Config.URL == "" {
			return fmt.Errorf("%w: missing url: Prowlarr config %d", ErrInvalidApp, i+1)
		}

		prowl.Debugf = a.DebugLog.Printf
		prowl.Errorf = a.ErrorLog.Printf
		prowl.setup(timeout)
	}

	return nil
}

func (r *ProwlarrConfig) setup(timeout time.Duration) {
	r.Prowlarr = prowlarr.New(r.Config)
	if r.Timeout.Duration == 0 {
		r.Timeout.Duration = timeout
	}

	r.URL = strings.TrimRight(r.URL, "/")

	if u, err := r.GetURL(); err != nil {
		r.Errorf("Checking Prowlarr Path: %v", err)
	} else if u := strings.TrimRight(u, "/"); u != r.URL {
		r.Errorf("Prowlarr URL fixed: %s -> %s (continuing)", r.URL, u)
		r.URL = u
	}
}
