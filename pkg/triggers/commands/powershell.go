package commands

import (
	"regexp"
	"slices"
	"strings"
)

// psParamName is a PowerShell parameter name that must stay unquoted so it
// binds like `powershell.exe -File script.ps1 -Force`. Quoted `'-Force'` is a
// string and lands positionally. Colon values are quoted separately so `;|&()`
// cannot splice into the synthesized -Command string.
var psParamName = regexp.MustCompile(`^-[A-Za-z][A-Za-z0-9]*$`)

// wrapPowerShell makes powershell.exe -File failures visible to cmd.Run().
// Windows PowerShell treats cmdlet errors (Start-Process, etc.) as non-terminating,
// so powershell.exe -File exits 0 and custom commands show 0 failures.
// Rewriting -File into -Command with $ErrorActionPreference='Stop' turns those
// into a non-zero process exit. Existing host-level -Command / -EncodedCommand
// (before -File) is left alone; switches after -File belong to the script.
func wrapPowerShell(args []string) []string {
	if len(args) < 2 || !isPowerShell(args[0]) {
		return args
	}

	flags := args[1:]
	fileIdx := indexPSSwitch(flags, "file", "f")
	if fileIdx < 0 || fileIdx+1 >= len(flags) {
		return args
	}

	if hasPSSwitch(flags[:fileIdx], "command", "c", "encodedcommand", "e", "ec") {
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
	cmd.WriteString(psScript(script))

	for _, arg := range extra {
		cmd.WriteByte(' ')
		cmd.WriteString(psArg(arg))
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

// psScript keeps -File path resolution: a bare name is found in the current
// directory by -File, but the call operator needs an explicit relative path.
func psScript(s string) string {
	if strings.ContainsAny(s, `/\:`) {
		return psQuote(s)
	}

	return psQuote("./" + s)
}

func psArg(arg string) string {
	if name, value, ok := strings.Cut(arg, ":"); ok && psParamName.MatchString(name) {
		return name + ":" + psColonValue(value)
	}

	if psParamName.MatchString(arg) {
		return arg
	}

	return psQuote(arg)
}

func psColonValue(value string) string {
	switch strings.ToLower(value) {
	case "$true", "$false", "$null":
		return value
	default:
		return psQuote(value)
	}
}
