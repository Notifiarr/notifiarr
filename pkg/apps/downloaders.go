package apps

import (
	"fmt"
	"time"

	"golift.io/cnfg"
	"golift.io/deluge"
	"golift.io/nzbget"
	"golift.io/qbit"
)

/* Notifiarr Client provides minimal support for pulling data from Download clients. */

type DelugeConfig struct {
	*deluge.Config
	Name           string        `toml:"name" xml:"name" json:"name"`
	Interval       cnfg.Duration `toml:"interval" xml:"interval" json:"interval"`
	*deluge.Deluge `toml:"-" xml:"-" json:"-"`
}

type QbitConfig struct {
	*qbit.Config
	Name       string        `toml:"name" xml:"name" json:"name"`
	Interval   cnfg.Duration `toml:"interval" xml:"interval" json:"interval"`
	*qbit.Qbit `toml:"-" xml:"-" json:"-"`
}

type TautulliConfig struct {
	Name     string        `toml:"name" xml:"name" json:"name"`
	Interval cnfg.Duration `toml:"interval" xml:"interval" json:"interval"`
	Timeout  cnfg.Duration `toml:"timeout" xml:"timeout" json:"timeout"`
	URL      string        `toml:"url" xml:"url" json:"url"`
	APIKey   string        `toml:"api_key" xml:"api_key" json:"apiKey"`
}

type NZBGetConfig struct {
	*nzbget.Config
	Name           string        `toml:"name" xml:"name"`
	Interval       cnfg.Duration `toml:"interval" xml:"interval"`
	*nzbget.NZBGet `toml:"-" xml:"-" json:"-"`
}

func (a *Apps) setupDeluge(timeout time.Duration) error {
	for idx := range a.Deluge {
		if a.Deluge[idx] == nil || a.Deluge[idx].Config == nil || a.Deluge[idx].Config.URL == "" {
			return fmt.Errorf("%w: missing url: Deluge config %d", ErrInvalidApp, idx+1)
		}

		// a.Deluge[i].Debugf = a.DebugLog.Printf
		if err := a.Deluge[idx].setup(timeout); err != nil {
			return err
		}
	}

	return nil
}

func (d *DelugeConfig) setup(timeout time.Duration) (err error) {
	_, d.Deluge, err = deluge.NewNoAuth(d.Config)

	if err != nil {
		return fmt.Errorf("deluge setup failed: %w", err)
	}

	if d.Timeout.Duration == 0 {
		d.Timeout.Duration = timeout
	}

	return nil
}

func (a *Apps) setupQbit(timeout time.Duration) error {
	for idx := range a.Qbit {
		if a.Qbit[idx].Config == nil || a.Qbit[idx].URL == "" {
			return fmt.Errorf("%w: missing url: Qbit config %d", ErrInvalidApp, idx+1)
		}

		// a.Qbit[i].Debugf = a.DebugLog.Printf
		if err := a.Qbit[idx].setup(timeout); err != nil {
			return err
		}
	}

	return nil
}

func (q *QbitConfig) setup(timeout time.Duration) (err error) {
	q.Qbit, err = qbit.NewNoAuth(q.Config)
	if err != nil {
		return fmt.Errorf("qbit setup failed: %w", err)
	}

	if q.Timeout.Duration == 0 {
		q.Timeout.Duration = timeout
	}

	return nil
}

func (a *Apps) setupNZBGet(timeout time.Duration) error {
	for idx := range a.NZBGet {
		if a.NZBGet[idx].Config == nil || a.NZBGet[idx].URL == "" {
			return fmt.Errorf("%w: missing url: NZBGet config %d", ErrInvalidApp, idx+1)
		}

		a.NZBGet[idx].NZBGet = nzbget.New(a.NZBGet[idx].Config)

		if a.NZBGet[idx].Timeout.Duration == 0 {
			a.NZBGet[idx].Timeout.Duration = timeout
		}
	}

	return nil
}
