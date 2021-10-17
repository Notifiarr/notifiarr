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

func (q *QbitConfig) setup(timeout time.Duration) (err error) {
	q.Qbit, err = qbit.New(q.Config)
	if err != nil {
		return fmt.Errorf("qbit setup failed: %w", err)
	}

	if q.Timeout.Duration == 0 {
		q.Timeout.Duration = timeout
	}

	return nil
}
