package snapshot

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/v3/host"
)

func (s *Snapshot) getSystemTemps(ctx context.Context) error {
	s.System.Temps = make(map[string]float64)

	temps, err := host.SensorsTemperaturesWithContext(ctx)

	for _, t := range temps {
		if t.Temperature > 0 {
			s.System.Temps[t.SensorKey] = t.Temperature
		}
	}

	/*
		https://github.com/shirou/gopsutil/issues/1377
		^^ see this for why this code is now commented out.

		if err == nil {
			return nil
		}

		var warns *host.Warnings
		if !errors.As(err, &warns) {
			return fmt.Errorf("unable to get sensor temperatures: %w", err)
		}

		errs := make([]string, len(warns.List))
		for i, w := range warns.List {
			errs[i] = fmt.Sprintf("warning %v: %v", i+1, w)
		}

		return fmt.Errorf("getting sensor temperatures: %w: %s", err, strings.Join(errs, ", "))
	*/

	if err != nil {
		return fmt.Errorf("warning getting sensor temperatures: %w", err)
	}

	return nil
}
