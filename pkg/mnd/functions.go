//nolint:gomnd
package mnd

import "fmt"

// FormatBytes converts a byte counter into a pretty UI string.
// The input val must be int, int64, uint64 or float64.
func FormatBytes(size interface{}) string { //nolint:cyclop
	var val float64

	switch valtype := size.(type) {
	case float64:
		val = valtype
	case int64:
		val = float64(valtype)
	case uint64:
		val = float64(valtype)
	case int:
		val = float64(valtype)
	default:
		panic("non-number provided to FormatBytes function")
	}

	switch {
	case val > Megabyte*Megabyte*Kilobyte*1000: // 2^60
		return fmt.Sprintf("%.2f EiB", val/float64(Megabyte*Megabyte*Megabyte))
	case val > Megabyte*Megabyte*1000: // 2^50
		return fmt.Sprintf("%.2f PiB", val/float64(Megabyte*Megabyte*Kilobyte))
	case val > Megabyte*Kilobyte*1000: // 2^40
		return fmt.Sprintf("%.2f TiB", val/float64(Megabyte*Megabyte))
	case val > Megabyte*1000: // 2^30
		return fmt.Sprintf("%.2f GiB", val/float64(Megabyte*Kilobyte))
	case val > Kilobyte*1000: // 2^20
		return fmt.Sprintf("%.1f MiB", val/float64(Megabyte))
	case val > 1000: // 2^10
		return fmt.Sprintf("%.1f KiB", val/float64(Kilobyte))
	default: // 2^1
		return fmt.Sprintf("%.0f B", val)
	}
}
