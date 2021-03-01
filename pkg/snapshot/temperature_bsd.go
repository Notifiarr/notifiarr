// +build freebsd openbsd netbsd

package snapshot

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/v3/host"
	"golang.org/x/sys/unix"
)

func (s *Snapshot) getSystemTemps(ctx context.Context, run bool) error {
	if !run {
		return nil
	}

	s.System.Temps = make(map[string]float64)

	temps, err := host.SensorsTemperaturesWithContext(ctx)
	if err != nil {
		return s.getSystemTempsFreeBSD()
	}

	for _, t := range temps {
		if t.Temperature > 0 {
			s.System.Temps[t.SensorKey] = t.Temperature
		}
	}

	return nil
}

// The host library may not support BSD, so try it ourselves.
// nolint: gomnd
func (s *Snapshot) getSystemTempsFreeBSD() error {
	temp, err := unix.SysctlUint32("dev.cpu.0.temperature")
	if err != nil {
		return fmt.Errorf("unable to get cpu temperature: %w", err)
	}

	// Convert from Kelvin * 10 to Celsius.
	s.System.Temps = map[string]float64{"cpu0": float64(int32(temp)-2732) / 10}

	for i := 1; i < 8; i++ {
		temp, err := unix.SysctlUint32(fmt.Sprintf("dev.cpu.%d.temperature", i))
		if err == nil {
			s.System.Temps[fmt.Sprintf("cpu%d", i)] = float64(int32(temp)-2732) / 10
		}
	}

	return nil
}
