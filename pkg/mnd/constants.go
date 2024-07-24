// Package mnd provides reusable constants for the Notifiarr application packages.
package mnd

import (
	"runtime"
	"time"
)

// Application Constants.
const (
	Mode0755      = 0o755
	Mode0750      = 0o750
	Mode0600      = 0o600
	Mode0644      = 0o644
	Kilobyte      = 1024
	Megabyte      = Kilobyte * Kilobyte
	KB100         = Kilobyte * 100
	OneDay        = 24 * time.Hour
	Base10        = 10
	Base8         = 8
	Bits64        = 64
	Bits32        = 32
	True          = "true"
	False         = "false"
	Windows       = "windows"
	Disabled      = "disabled"
	Success       = "success"
	Status        = "Status"
	BytesReceived = " Bytes Received"
	BytesSent     = " Bytes Sent"
	Matched       = " Matched"
	Requests      = " Requests"
	Deleted       = "deleted:"
	HelpLink      = "Notifiarr Discord: https://notifiarr.com/discord"
	UserRepo      = "Notifiarr/notifiarr"
	BugIssue      = "This is a bug please report it on github: https://github.com/" + UserRepo + "/issues/new"
	Synology      = "/etc/synoinfo.conf" // Synology is the path to the syno config file.
	IsLinux       = runtime.GOOS == "linux"
	IsWindows     = runtime.GOOS == Windows
	IsFreeBSD     = runtime.GOOS == "freebsd"
	IsDarwin      = runtime.GOOS == "darwin"
)

// Application Defaults.
const (
	Title            = "Notifiarr"
	DefaultName      = "notifiarr"
	DefaultLogFileMb = 100
	DefaultLogFiles  = 0 // delete none
	DefaultEnvPrefix = "DN"
	DefaultTimeout   = time.Minute
	DefaultBindAddr  = "0.0.0.0:5454"
)
