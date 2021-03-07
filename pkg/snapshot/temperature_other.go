// +build !freebsd,!openbsd,!netbsd

package snapshot

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/v3/host"
)

func (s *Snapshot) getSystemTemps(ctx context.Context, run bool) error {
	if !run {
		return nil
	}

	s.System.Temps = make(map[string]float64)

	temps, err := host.SensorsTemperaturesWithContext(ctx)
	if err != nil {
		return fmt.Errorf("unable to get sensor temperatures: %w", err)
	}

	for _, t := range temps {
		if t.Temperature > 0 {
			s.System.Temps[t.SensorKey] = t.Temperature
		}
	}

	return nil
}
