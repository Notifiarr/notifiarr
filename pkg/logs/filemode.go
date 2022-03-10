package logs

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"golift.io/rotatorr"
)

// These are used for custom logs.
// nolint:gochecknoglobals
var (
	fileMode = rotatorr.FileMode
)

// FileMode is used to unmarshal a unix file mode from the config file.
type FileMode os.FileMode

// UnmarshalText turns a unix file mode, wrapped in quotes or not, into a usable os.FileMode.
func (f *FileMode) UnmarshalText(text []byte) error {
	str := strings.TrimSpace(strings.Trim(string(text), `"'`))

	fm, err := strconv.ParseUint(str, mnd.Base8, mnd.Bits32)
	if err != nil {
		return fmt.Errorf("file_mode (%s) is invalid: %w", str, err)
	}

	*f = FileMode(os.FileMode(fm))

	return nil
}

// MarshalText satisfies an encoder.TextMarshaler interface.
func (f FileMode) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

// String creates a unix-octal version of a file mode.
func (f FileMode) String() string {
	return fmt.Sprintf("%04o", f)
}

// Mode returns the compatable os.FileMode.
func (f FileMode) Mode() os.FileMode {
	return os.FileMode(f)
}
