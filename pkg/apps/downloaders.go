package apps

import (
	"fmt"
	"time"

	"golift.io/cnfg"
	"golift.io/deluge"
	"golift.io/qbit"
)

/* Notifiarr Client provides minimal support for pulling data from Download clients. */

type DelugeConfig struct {
	Name     string        `toml:"name"`
	Interval cnfg.Duration `toml:"interval"`
	*deluge.Config
	*deluge.Deluge
}

type QbitConfig struct {
	Name     string        `toml:"name"`
	Interval cnfg.Duration `toml:"interval"`
	*qbit.Config
	*qbit.Qbit
}

type TautulliConfig struct {
	Name     string        `toml:"name"`
	Interval cnfg.Duration `toml:"interval"`
	Timeout  cnfg.Duration `toml:"timeout"`
	URL      string        `toml:"url"`
	APIKey   string        `toml:"api_key"`
}

func (a *Apps) setupDeluge(timeout time.Duration) error {
	for i := range a.Deluge {
		if a.Deluge[i] == nil || a.Deluge[i].Config == nil || a.Deluge[i].Config.URL == "" {
			return fmt.Errorf("%w: missing url: Deluge config %d", ErrInvalidApp, i+1)
		}

		// a.Deluge[i].Debugf = a.DebugLog.Printf
		if err := a.Deluge[i].setup(timeout); err != nil {
			return err
		}
	}

	return nil
}

func (d *DelugeConfig) setup(timeout time.Duration) (err error) {
	d.Deluge, err = deluge.New(*d.Config)
	if err != nil {
		return fmt.Errorf("deluge setup failed: %w", err)
	}

	if d.Timeout.Duration == 0 {
		d.Timeout.Duration = timeout
	}

	return nil
}

func (a *Apps) setupQbit(timeout time.Duration) error {
	for i := range a.Qbit {
		if a.Qbit[i].Config == nil || a.Qbit[i].URL == "" {
			return fmt.Errorf("%w: missing url: Qbit config %d", ErrInvalidApp, i+1)
		}

		// a.Qbit[i].Debugf = a.DebugLog.Printf
		if err := a.Qbit[i].setup(timeout); err != nil {
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
