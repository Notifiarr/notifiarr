package logs

import (
	"debug/pe"
	"os"
)

func getFileOwner(_ os.FileInfo) string {
	return ""
}

// Sometimes we compile with -H=windowsgui and sometimes without.
// Having this function allows us to detect which, so we can turn on/off console logging.
func hasConsoleWindow() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}

	file, err := pe.Open(exe)
	if err != nil {
		return false
	}
	defer file.Close()

	const windowsTerminal = 3

	if header, ok := file.OptionalHeader.(*pe.OptionalHeader64); ok {
		return header.Subsystem == windowsTerminal
	} else if header, ok := file.OptionalHeader.(*pe.OptionalHeader32); ok {
		return header.Subsystem == windowsTerminal
	}

	return false
}
