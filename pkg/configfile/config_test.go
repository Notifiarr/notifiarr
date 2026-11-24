package configfile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/stretchr/testify/require"
	"golift.io/cnfgfile"
)

func TestWriteExampleLeavesClientNameUnset(t *testing.T) { //nolint:paralleltest // toggles configfile.ForceAllTmpl
	configfile.ForceAllTmpl = true
	t.Cleanup(func() { configfile.ForceAllTmpl = false })

	cfg := configfile.NewConfig()
	file := filepath.Join(t.TempDir(), "notifiarr.conf")

	_, err := cfg.Write(t.Context(), file, false)
	require.NoError(t, err)
	require.Empty(t, cfg.ClientName)

	body, err := os.ReadFile(file)
	require.NoError(t, err)
	require.Contains(t, string(body), `#client_name = "my-client"`)
	require.NotContains(t, string(body), "client_name = '''")
}

func TestWritePersistsDefaultClientName(t *testing.T) {
	t.Parallel()

	cfg := configfile.NewConfig()
	file := filepath.Join(t.TempDir(), "notifiarr.conf")

	wrote, err := cfg.Write(t.Context(), file, false)
	require.NoError(t, err)
	require.Equal(t, file, wrote)
	require.NotEmpty(t, cfg.ClientName)

	loaded := configfile.NewConfig()
	require.NoError(t, cnfgfile.Unmarshal(loaded, file))
	require.Equal(t, cfg.ClientName, loaded.ClientName)

	info, err := host.InfoWithContext(t.Context())
	require.NoError(t, err)
	require.Equal(t, website.CleanHostname(info), cfg.ClientName)
}

func TestWriteKeepsClientName(t *testing.T) {
	t.Parallel()

	cfg := configfile.NewConfig()
	cfg.ClientName = "  library  "
	file := filepath.Join(t.TempDir(), "notifiarr.conf")

	_, err := cfg.Write(t.Context(), file, false)
	require.NoError(t, err)
	require.Equal(t, "library", cfg.ClientName)

	loaded := configfile.NewConfig()
	require.NoError(t, cnfgfile.Unmarshal(loaded, file))
	require.Equal(t, "library", loaded.ClientName)
}
