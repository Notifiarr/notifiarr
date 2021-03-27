package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Go-Lift-TV/notifiarr/pkg/apps"
	"github.com/Go-Lift-TV/notifiarr/pkg/logs"
	"github.com/Go-Lift-TV/notifiarr/pkg/notifiarr"
	"golift.io/cnfg"
)

// Defaults.
const (
	DefaultInterval = 10 * time.Minute
	MinimumInterval = DefaultInterval / 2
	MinimumTimeout  = time.Second
	DefaultTimeout  = 10 * MinimumTimeout
	MaximumParallel = 10
	DefaultBuffer   = 1000
)

var (
	ErrNoName      = fmt.Errorf("service check is missing a unique name")
	ErrNoCheck     = fmt.Errorf("service check is missing a check value")
	ErrInvalidType = fmt.Errorf("service check type must be one of %s, %s", CheckTCP, CheckHTTP)
	ErrBadTCP      = fmt.Errorf("tcp checks must have an ip:port or host:port combo; the :port is required")
)

// Config for this plugin.
type Config struct {
	Interval     cnfg.Duration     `toml:"interval"`
	Parallel     uint              `toml:"parallel"`
	Disabled     bool              `toml:"disabled"`
	Apps         *apps.Apps        `toml:"-"`
	Notify       *notifiarr.Config `toml:"-"`
	*logs.Logger `json:"-"`        // log file writer
	services     map[string]*Service
	checks       chan *Service
	done         chan struct{}
	stopChan     chan struct{}
}

// CheckType locks us into a few specific types of checks.
type CheckType string

// These are our supported Check Types.
const (
	CheckHTTP CheckType = "http"
	CheckTCP  CheckType = "tcp"
	CheckPING CheckType = "ping"
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

// CheckResult is sent to Notifiarr.
type CheckResult struct {
	Name   string     `json:"name"`   // "Radarr"
	State  CheckState `json:"state"`  // 0 = OK, 1 = Warn, 2 = Crit, 3 = Unknown
	Output string     `json:"output"` // metadata message
	Type   CheckType  `json:"type"`   // http, tcp, ping
	Time   time.Time  `json:"time"`   // when it was checked
}

// Service is a thing we check and report results for.
type Service struct {
	Name      string        `toml:"name"`    // Radarr
	Type      CheckType     `toml:"type"`    // http
	Value     string        `toml:"check"`   // http://some.url
	Expect    string        `toml:"expect"`  // 200
	Timeout   cnfg.Duration `toml:"timeout"` // 10s
	output    string
	state     CheckState
	lastCheck time.Time
}

func (c *Config) Start(services []*Service) error {
	if c.Disabled {
		return nil
	}

	if err := c.setup(services); err != nil {
		return err
	}

	c.startCheckers()

	go c.run()
	c.Printf("==> Service Checker Started! %d services, interval: %s", len(c.services), c.Interval)

	return nil
}

func (c *Config) setup(services []*Service) error {
	c.services = make(map[string]*Service)
	c.checks = make(chan *Service, DefaultBuffer)
	c.done = make(chan struct{})
	c.stopChan = make(chan struct{})

	services = append(services, c.collectApps()...)

	for i := range services {
		if err := services[i].validate(); err != nil {
			return err
		}

		// Add this validated service to our service map.
		c.services[services[i].Name] = services[i]
	}

	if c.Parallel > MaximumParallel {
		c.Parallel = MaximumParallel
	} else if c.Parallel == 0 {
		c.Parallel = 1
	}

	if c.Interval.Duration == 0 {
		c.Interval.Duration = DefaultInterval
	} else if c.Interval.Duration < MinimumInterval {
		c.Interval.Duration = MinimumInterval
	}

	return nil
}

func (c *Config) collectApps() []*Service {
	svcs := []*Service{}

	for _, a := range c.Apps.Lidarr {
		if a.Name != "" {
			svcs = append(svcs, &Service{
				Name:    a.Name,
				Type:    CheckHTTP,
				Value:   a.URL,
				Expect:  "200",
				Timeout: cnfg.Duration{Duration: a.Timeout.Duration},
			})
		}
	}

	for _, a := range c.Apps.Radarr {
		if a.Name != "" {
			svcs = append(svcs, &Service{
				Name:    a.Name,
				Type:    CheckHTTP,
				Value:   a.URL,
				Expect:  "200",
				Timeout: cnfg.Duration{Duration: a.Timeout.Duration},
			})
		}
	}

	for _, a := range c.Apps.Readarr {
		if a.Name != "" {
			svcs = append(svcs, &Service{
				Name:    a.Name,
				Type:    CheckHTTP,
				Value:   a.URL,
				Expect:  "200",
				Timeout: cnfg.Duration{Duration: a.Timeout.Duration},
			})
		}
	}

	for _, a := range c.Apps.Sonarr {
		if a.Name != "" {
			svcs = append(svcs, &Service{
				Name:    a.Name,
				Type:    CheckHTTP,
				Value:   a.URL,
				Expect:  "200",
				Timeout: cnfg.Duration{Duration: a.Timeout.Duration},
			})
		}
	}

	return svcs
}

func (c *Config) run() {
	ticker := time.NewTicker(c.Interval.Duration)
	defer func() {
		ticker.Stop()
		c.done <- struct{}{}
	}()

	for {
		select {
		case <-ticker.C:
			c.reportChecks()
		case <-c.stopChan:
			return
		}
	}
}

func (c *Config) reportChecks() {
	for s := range c.services {
		c.checks <- c.services[s]
	}

	for range c.services {
		<-c.done
	}

	svcs := []*CheckResult{}

	for _, s := range c.services {
		c.Printf("Service Checked: %s, state: %s, output: %s", s.Name, s.state, s.output)

		svcs = append(svcs, &CheckResult{
			Name:   s.Name,
			State:  s.state,
			Output: s.output,
			Type:   s.Type,
			Time:   s.lastCheck,
		})
	}

	data, _ := json.MarshalIndent(&struct {
		Type string         `json:"eventType"`
		Svcs []*CheckResult `json:"services"`
	}{Type: "service_checks", Svcs: svcs}, "", " ")

	c.Debug("Sending Payload:", string(data))

	if _, err := c.Notify.SendJSON(notifiarr.BaseURL, data); err != nil {
		c.Error("Sending service check update to Notifiarr:", err)
	}
}

func (c *Config) Stop() {
	if c.stopChan == nil {
		return
	}

	close(c.stopChan)
	c.stopChan = nil
	<-c.done

	close(c.checks)
	c.checks = nil

	for i := uint(0); i < c.Parallel; i++ {
		<-c.done
	}

	close(c.done)
	c.done = nil
}

func (c CheckState) String() string {
	switch c {
	default:
		fallthrough
	case StateUnknown:
		return "Unknown"
	case StateCritical:
		return "Critical"
	case StateWarning:
		return "Warning"
	case StateOK:
		return "OK"
	}
}
