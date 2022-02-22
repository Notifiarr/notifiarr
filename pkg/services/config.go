package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/notifiarr"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"golift.io/cnfg"
)

// Services Defaults.
const (
	DefaultSendInterval  = 10 * time.Minute
	MinimumSendInterval  = DefaultSendInterval / 2
	DefaultCheckInterval = MinimumSendInterval
	MinimumCheckInterval = 10 * time.Second
	MinimumTimeout       = time.Second
	DefaultTimeout       = 10 * MinimumTimeout
	MaximumParallel      = 10
	DefaultBuffer        = 1000
)

// Errors returned by this Services package.
var (
	ErrNoName      = fmt.Errorf("service check is missing a unique name")
	ErrNoCheck     = fmt.Errorf("service check is missing a check value")
	ErrInvalidType = fmt.Errorf("service check type must be one of %s, %s, %s", CheckTCP, CheckHTTP, CheckPROC)
	ErrBadTCP      = fmt.Errorf("tcp checks must have an ip:port or host:port combo; the :port is required")
)

// Config for this Services plugin comes from a config file.
type Config struct {
	Interval     cnfg.Duration     `toml:"interval" xml:"interval" json:"interval"`
	Parallel     uint              `toml:"parallel" xml:"parallel" json:"parallel"`
	Disabled     bool              `toml:"disabled" xml:"disabled" json:"disabled"`
	LogFile      string            `toml:"log_file" xml:"log_file" json:"logFile"`
	Apps         *apps.Apps        `toml:"-" json:"-"`
	Notifiarr    *notifiarr.Config `toml:"-" json:"-"`
	Plugins      *snapshot.Plugins `toml:"-" json:"-"`
	*logs.Logger `json:"-"`        // log file writer
	services     map[string]*Service
	checks       chan *Service
	done         chan bool
	stopChan     chan struct{}
	triggerChan  chan notifiarr.EventType
	checkChan    chan triggerCheck
	stopLock     sync.Mutex
}

// CheckType locks us into a few specific types of checks.
type CheckType string

// These are our supported Check Types.
const (
	CheckHTTP CheckType = "http"
	CheckTCP  CheckType = "tcp"
	CheckPING CheckType = "ping"
	CheckPROC CheckType = "process"
)

// CheckState represents the current state of a service check.
type CheckState uint

// Supported check states.
const (
	StateOK CheckState = iota
	StateWarning
	StateCritical
	StateUnknown
)

// Results is sent to Notifiarr.
type Results struct {
	Type     string              `json:"eventType"`
	What     notifiarr.EventType `json:"what"`
	Interval float64             `json:"interval"`
	Svcs     []*CheckResult      `json:"services"`
}

// CheckResult represents the status of a service.
type CheckResult struct {
	Name        string        `json:"name"`   // "Radarr"
	State       CheckState    `json:"state"`  // 0 = OK, 1 = Warn, 2 = Crit, 3 = Unknown
	Output      string        `json:"output"` // metadata message
	Type        CheckType     `json:"type"`   // http, tcp, ping
	Time        time.Time     `json:"time"`   // when it was checked, rounded to Microseconds
	Since       time.Time     `json:"since"`  // how long it has been in this state, rounded to Microseconds
	Interval    float64       `json:"interval"`
	Check       string        `json:"-"`
	Expect      string        `json:"-"`
	IntervalDur time.Duration `json:"-"`
}

// Service is a thing we check and report results for.
type Service struct {
	Name      string        `toml:"name" xml:"name" json:"name"`             // Radarr
	Type      CheckType     `toml:"type" xml:"type" json:"type"`             // http
	Value     string        `toml:"check" xml:"check" json:"value"`          // http://some.url
	Expect    string        `toml:"expect" xml:"expect" json:"expect"`       // 200
	Timeout   cnfg.Duration `toml:"timeout" xml:"timeout" json:"timeout"`    // 10s
	Interval  cnfg.Duration `toml:"interval" xml:"interval" json:"interval"` // 1m
	log       *logs.Logger
	output    string
	state     CheckState
	since     time.Time
	lastCheck time.Time
	proc      *procExpect // only used for process checks.
}
