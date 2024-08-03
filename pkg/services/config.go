package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/website"
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
	ErrNoName      = errors.New("service check is missing a unique name")
	ErrNoCheck     = errors.New("service check is missing a check value")
	ErrInvalidType = fmt.Errorf("service check type must be one of %s, %s, %s, %s, %s",
		CheckTCP, CheckHTTP, CheckPROC, CheckPING, CheckICMP)
	ErrBadTCP = errors.New("tcp checks must have an ip:port or host:port combo; the :port is required")
)

// Config for this Services plugin comes from a config file.
type Config struct {
	Interval    cnfg.Duration     `json:"interval" toml:"interval" xml:"interval"`
	Parallel    uint              `json:"parallel" toml:"parallel" xml:"parallel"`
	Disabled    bool              `json:"disabled" toml:"disabled" xml:"disabled"`
	LogFile     string            `json:"logFile"  toml:"log_file" xml:"log_file"`
	Apps        *apps.Apps        `json:"-"        toml:"-"`
	website     *website.Server   `json:"-"        toml:"-"`
	Plugins     *snapshot.Plugins `json:"-"        toml:"-"` // pass this in so we can service-check mysql
	mnd.Logger  `json:"-"`        // log file writer
	services    map[string]*Service
	checks      chan *Service
	done        chan bool
	stopChan    chan struct{}
	triggerChan chan website.EventType
	checkChan   chan triggerCheck
	stopLock    sync.Mutex
}

// CheckType locks us into a few specific types of checks.
type CheckType string

// These are our supported Check Types.
const (
	CheckHTTP CheckType = "http"
	CheckTCP  CheckType = "tcp"
	CheckPING CheckType = "ping"
	CheckICMP CheckType = "icmp"
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

// Results is sent to website.
type Results struct {
	Type     string            `json:"eventType"`
	What     website.EventType `json:"what"`
	Interval float64           `json:"interval"`
	Svcs     []*CheckResult    `json:"services"`
}

// CheckResult represents the status of a service.
type CheckResult struct {
	Name        string         `json:"name"`     // "Radarr"
	State       CheckState     `json:"state"`    // 0 = OK, 1 = Warn, 2 = Crit, 3 = Unknown
	Output      *Output        `json:"output"`   // metadata message must never be nil.
	Type        CheckType      `json:"type"`     // http, tcp, ping
	Time        time.Time      `json:"time"`     // when it was checked, rounded to Microseconds
	Since       time.Time      `json:"since"`    // how long it has been in this state, rounded to Microseconds
	Interval    float64        `json:"interval"` // interval in seconds
	Metadata    map[string]any `json:"metadata"` // arbitrary info about the service or result.
	Check       string         `json:"-"`
	Expect      string         `json:"-"`
	IntervalDur time.Duration  `json:"-"`
}

// Service is a thing we check and report results for.
type Service struct {
	Name     string         `json:"name"     toml:"name"     xml:"name"`     // Radarr
	Type     CheckType      `json:"type"     toml:"type"     xml:"type"`     // http
	Value    string         `json:"value"    toml:"check"    xml:"check"`    // http://some.url
	Expect   string         `json:"expect"   toml:"expect"   xml:"expect"`   // 200
	Timeout  cnfg.Duration  `json:"timeout"  toml:"timeout"  xml:"timeout"`  // 10s
	Interval cnfg.Duration  `json:"interval" toml:"interval" xml:"interval"` // 1m
	Tags     map[string]any `json:"tags"     toml:"tags"     xml:"tags"`     // copied to Metadata.
	validSSL bool           // can be set for https checks.
	svc      service
}

type service struct {
	Output       *Output    `json:"output"`
	State        CheckState `json:"state"`
	Since        time.Time  `json:"since"`
	LastCheck    time.Time  `json:"lastCheck"`
	log          mnd.Logger
	proc         *procExpect // only used for process checks.
	ping         *pingExpect // only used for icmp/udp ping checks.
	sync.RWMutex `json:"-"`
}

type Output struct {
	str string // output string
	esc bool   // html escaped?
}

func (o *Output) String() string {
	switch {
	case o == nil:
		return ""
	case o.esc:
		return html.UnescapeString(o.str)
	default:
		return o.str
	}
}

func (o *Output) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.str) //nolint:wrapcheck // do not unescape it.
}

func (o *Output) UnmarshalJSON(input []byte) error {
	return json.Unmarshal(input, &o.str) //nolint:wrapcheck
}
