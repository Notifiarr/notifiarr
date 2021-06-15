// Package ui wraps some OS-specific methods so they can be used on any architecture/OS
// without compile-time errors. The packages wrapped are github.com/gen2brain/dlgs and
// github.com/getlantern/systray. The notifiarr client application only uses these on
// Windows and macOS, so they are wrapped to ignore them on macOS, linux and freeBSD.
package ui

// MenuItem is an interface to allow exposing menu items to operating systems
// that do not have a menu or a GUI.
type MenuItem interface {
	Check()
	Checked() bool
	Disable()
	Disabled() bool
	Enable()
	Hide()
	SetIcon(iconBytes []byte)
	SetTemplateIcon(templateIconBytes []byte, regularIconBytes []byte)
	SetTitle(title string)
	SetTooltip(tooltip string)
	Show()
	String() string
	Uncheck()
	Clicked() chan struct{}
}
