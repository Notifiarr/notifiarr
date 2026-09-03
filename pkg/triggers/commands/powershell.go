package commands

import (
	"slices"
	"strings"
)

// wrapPowerShell makes powershell.exe -File failures visible to cmd.Run().
// Windows PowerShell treats cmdlet errors (Start-Process, etc.) as non-terminating,
// so powershell.exe -File exits 0 and custom commands show 0 failures.
// Rewriting -File into -Command with $ErrorActionPreference='Stop' turns those
// into a non-zero process exit. Existing -Command / -EncodedCommand is left alone.
func wrapPowerShell(args []string) []string {
	if len(args) < 2 || !isPowerShell(args[0]) {
		return args
	}

	flags := args[1:]
	if hasPSSwitch(flags, "command", "c", "encodedcommand", "e", "ec") {
		return args
	}

	fileIdx := indexPSSwitch(flags, "file", "f")
	if fileIdx < 0 || fileIdx+1 >= len(flags) {
		return args
	}

	script := flags[fileIdx+1]
	extra := flags[fileIdx+2:]
	kept := append([]string{}, flags[:fileIdx]...)

	if !hasPSSwitch(kept, "noninteractive") {
		kept = append(kept, "-NonInteractive")
	}

	var cmd strings.Builder
	cmd.WriteString("$ErrorActionPreference='Stop'; & ")
	cmd.WriteString(psQuote(script))

	for _, arg := range extra {
		cmd.WriteByte(' ')
		cmd.WriteString(psQuote(arg))
	}

	return append(append([]string{args[0]}, kept...), "-Command", cmd.String())
}

func isPowerShell(exe string) bool {
	name := exe
	if idx := strings.LastIndexAny(exe, `/\`); idx >= 0 {
		name = exe[idx+1:]
	}

	switch strings.ToLower(name) {
	case "powershell.exe", "powershell", "pwsh.exe", "pwsh":
		return true
	default:
		return false
	}
}

func normalizePSArg(arg string) string {
	return strings.ToLower(strings.TrimLeft(arg, "-/"))
}

func hasPSSwitch(args []string, names ...string) bool {
	return indexPSSwitch(args, names...) >= 0
}

func indexPSSwitch(args []string, names ...string) int {
	for idx, arg := range args {
		if slices.Contains(names, normalizePSArg(arg)) {
			return idx
		}
	}

	return -1
}

func psQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
