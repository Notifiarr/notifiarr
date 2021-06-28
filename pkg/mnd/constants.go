// Package con provides re-usable constants for the Notifiarr application packages.
package mnd

import "time"

// Application Constants.
const (
	Mode0755 = 0755
	Mode0750 = 0750
	Mode0600 = 0600
	Megabyte = 1024 * 1024
	OneDay   = 24 * time.Hour
	Base10   = 10
	Base8    = 8
	Bits64   = 64
	Bits32   = 32
	Windows  = "windows"
	HelpLink = "GoLift Discord: https://golift.io/discord"
	UserRepo = "Notifiarr/notifiarr"
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
