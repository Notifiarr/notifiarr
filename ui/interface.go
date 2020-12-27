package ui

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
