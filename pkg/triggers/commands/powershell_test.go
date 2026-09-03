package commands //nolint:testpackage

import (
	"slices"
	"testing"
)

func TestWrapPowerShellFile(t *testing.T) {
	t.Parallel()

	exe := `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`
	script := `C:\Users\ee\Desktop\clientCommands\plexRestart.ps1`

	got := wrapPowerShell([]string{exe, "-File", script})
	want := []string{exe, "-NonInteractive", "-Command", "$ErrorActionPreference='Stop'; & '" + script + "'"}

	if !slices.Equal(got, want) {
		t.Fatalf("wrapPowerShell() = %#v, want %#v", got, want)
	}
}

func TestWrapPowerShellFileExtras(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "minus f with script args",
			in:   []string{"powershell.exe", "-f", `C:\a.ps1`, "one", "two"},
			want: []string{
				"powershell.exe", "-NonInteractive", "-Command",
				`$ErrorActionPreference='Stop'; & 'C:\a.ps1' 'one' 'two'`,
			},
		},
		{
			name: "preserves preceding flags",
			in:   []string{"pwsh.exe", "-ExecutionPolicy", "Bypass", "-File", `D:\run.ps1`},
			want: []string{
				"pwsh.exe", "-ExecutionPolicy", "Bypass", "-NonInteractive", "-Command",
				`$ErrorActionPreference='Stop'; & 'D:\run.ps1'`,
			},
		},
		{
			name: "does not duplicate noninteractive",
			in:   []string{"powershell", "-NonInteractive", "-File", `C:\x.ps1`},
			want: []string{
				"powershell", "-NonInteractive", "-Command",
				`$ErrorActionPreference='Stop'; & 'C:\x.ps1'`,
			},
		},
		{
			name: "escapes single quotes in path",
			in:   []string{"powershell.exe", "-File", `C:\O'Brien\run.ps1`},
			want: []string{
				"powershell.exe", "-NonInteractive", "-Command",
				`$ErrorActionPreference='Stop'; & 'C:\O''Brien\run.ps1'`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := wrapPowerShell(test.in)
			if !slices.Equal(got, test.want) {
				t.Fatalf("wrapPowerShell() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestWrapPowerShellSkip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []string
	}{
		{name: "not powershell", in: []string{"taskkill", "/IM", "Plex Media Server.exe", "/F"}},
		{name: "minus command", in: []string{"powershell.exe", "-Command", "Get-Process"}},
		{name: "minus c", in: []string{"pwsh", "-c", "Get-Date"}},
		{name: "encoded command", in: []string{"powershell.exe", "-EncodedCommand", "QQA="}},
		{name: "file without path", in: []string{"powershell.exe", "-File"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := wrapPowerShell(test.in)
			if !slices.Equal(got, test.in) {
				t.Fatalf("wrapPowerShell() = %#v, want unchanged %#v", got, test.in)
			}
		})
	}
}
