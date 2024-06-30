package snapshot

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/v4/sensors"
)

func (s *Snapshot) getSystemTemps(ctx context.Context) error {
	s.System.Temps = make(map[string]float64)

	temps, err := sensors.TemperaturesWithContext(ctx)

	for _, t := range temps {
		if t.Temperature > 0 {
			s.System.Temps[t.SensorKey] = t.Temperature
		}
	}

	if err == nil {
		return nil
	}

	// Unmarshal the error for more info.
	var warns *sensors.Warnings
	if !errors.As(err, &warns) {
		return fmt.Errorf("unable to get sensor temperatures: %w", err)
	}

	errs := make([]string, len(warns.List))
	for i, w := range warns.List {
		errs[i] = fmt.Sprintf("warning %v: %v", i+1, w)
	}

	return fmt.Errorf("getting sensor temperatures: %w: %s", err, strings.Join(errs, ", "))
}
