package services

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/notifiarr"
	"golift.io/cnfg"
)

// Defaults.
const (
	DefaultSendInterval  = 10 * time.Minute
	MinimumSendInterval  = DefaultSendInterval / 2
	DefaultCheckInterval = MinimumSendInterval
	MinimumCheckInterval = 10 * time.Second
	MinimumTimeout       = time.Second
	DefaultTimeout       = 10 * MinimumTimeout
	MaximumParallel      = 10
	DefaultBuffer        = 1000
	NotifiarrEventType   = "service_checks"
)

var (
	ErrNoName      = fmt.Errorf("service check is missing a unique name")
	ErrNoCheck     = fmt.Errorf("service check is missing a check value")
	ErrInvalidType = fmt.Errorf("service check type must be one of %s, %s, %s", CheckTCP, CheckHTTP, CheckPROC)
	ErrBadTCP      = fmt.Errorf("tcp checks must have an ip:port or host:port combo; the :port is required")
)

// Config for this plugin.
type Config struct {
	Interval     cnfg.Duration     `toml:"interval" xml:"interval"`
	Parallel     uint              `toml:"parallel" xml:"parallel"`
	Disabled     bool              `toml:"disabled" xml:"disabled"`
	Apps         *apps.Apps        `toml:"-"`
	Notify       *notifiarr.Config `toml:"-"`
	*logs.Logger `json:"-"`        // log file writer
	services     map[string]*Service
	checks       chan *Service
	done         chan bool
	stopChan     chan struct{}
	mu           sync.Mutex
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
	Type     string         `json:"eventType"`
	What     string         `json:"what"`
	Interval float64        `json:"interval"`
	Svcs     []*CheckResult `json:"services"`
}

// CheckResult represents the status of a service.
type CheckResult struct {
	Name     string     `json:"name"`   // "Radarr"
	State    CheckState `json:"state"`  // 0 = OK, 1 = Warn, 2 = Crit, 3 = Unknown
	Output   string     `json:"output"` // metadata message
	Type     CheckType  `json:"type"`   // http, tcp, ping
	Time     time.Time  `json:"time"`   // when it was checked, rounded to Microseconds
	Since    time.Time  `json:"since"`  // how long it has been in this state, rounded to Microseconds
	Interval float64    `json:"interval"`
}

// Service is a thing we check and report results for.
type Service struct {
	Name      string        `toml:"name"`     // Radarr
	Type      CheckType     `toml:"type"`     // http
	Value     string        `toml:"check"`    // http://some.url
	Expect    string        `toml:"expect"`   // 200
	Timeout   cnfg.Duration `toml:"timeout"`  // 10s
	Interval  cnfg.Duration `toml:"interval"` // 1m
	log       *logs.Logger
	output    string
	state     CheckState
	since     time.Time
	lastCheck time.Time
	proc      *procExpect // only used for process checks.
}

func (c *Config) Start(services []*Service) error {
	services = append(services, c.collectApps()...)
	if c.Disabled || len(services) == 0 {
		return nil
	}

	if err := c.setup(services); err != nil {
		return err
	}

	c.start()

	return nil
}

// start runs Parallel checkers and the check reporter.
func (c *Config) start() {
	for i := uint(0); i < c.Parallel; i++ {
		go func() {
			for check := range c.checks {
				c.done <- check.check()
			}

			c.done <- false
		}()
	}

	go c.runServiceChecker()
}

func (c *Config) runServiceChecker() {
	c.Printf("==> Service Checker Started! %d services, interval: %s", len(c.services), c.Interval)
	c.RunChecks(true)
	c.SendResults(notifiarr.ProdURL, &Results{
		What: "start",
		Svcs: c.GetResults(),
	})

	ticker := time.NewTicker(c.Interval.Duration)
	second := time.NewTicker(time.Millisecond * 4159) //nolint:gomnd

	defer func() {
		second.Stop()
		ticker.Stop()
		c.done <- false
	}()

	for {
		select {
		case <-second.C:
			c.RunChecks(false)
		case <-ticker.C:
			c.SendResults(notifiarr.ProdURL, &Results{
				What: "timer",
				Svcs: c.GetResults(),
			})
		case <-c.stopChan:
			return
		}
	}
}

func (c *Config) setup(services []*Service) error {
	c.services = make(map[string]*Service)
	c.checks = make(chan *Service, DefaultBuffer)
	c.done = make(chan bool)
	c.stopChan = make(chan struct{})

	for i := range services {
		services[i].log = c.Logger
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
		c.Interval.Duration = DefaultSendInterval
	} else if c.Interval.Duration < MinimumSendInterval {
		c.Interval.Duration = MinimumSendInterval
	}

	return nil
}

// collectApps turns app configs into service checks if they have a name.
func (c *Config) collectApps() []*Service {
	svcs := []*Service{}

	for _, a := range c.Apps.Lidarr {
		if a.Name != "" {
			svcs = append(svcs, &Service{
				Name:     a.Name,
				Type:     CheckHTTP,
				Value:    a.URL,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: a.Timeout.Duration},
				Interval: a.Interval,
			})
		}
	}

	for _, a := range c.Apps.Radarr {
		if a.Name != "" {
			svcs = append(svcs, &Service{
				Name:     a.Name,
				Type:     CheckHTTP,
				Value:    a.URL,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: a.Timeout.Duration},
				Interval: a.Interval,
			})
		}
	}

	for _, a := range c.Apps.Readarr {
		if a.Name != "" {
			svcs = append(svcs, &Service{
				Name:     a.Name,
				Type:     CheckHTTP,
				Value:    a.URL,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: a.Timeout.Duration},
				Interval: a.Interval,
			})
		}
	}

	for _, a := range c.Apps.Sonarr {
		if a.Name != "" {
			svcs = append(svcs, &Service{
				Name:     a.Name,
				Type:     CheckHTTP,
				Value:    a.URL,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: a.Timeout.Duration},
				Interval: a.Interval,
			})
		}
	}

	return svcs
}

// RunChecks runs checks that are due. Passing true, runs them even if they're not due.
// Returns true if a service state changed.
func (c *Config) RunChecks(forceAll bool) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	stateChange := false

	for s := range c.services {
		if forceAll || c.services[s].lastCheck.Add(c.services[s].Interval.Duration).Before(time.Now()) {
			count++
			c.checks <- c.services[s]
		}
	}

	for ; count > 0; count-- {
		if sc := <-c.done; sc {
			stateChange = true
		}
	}

	return stateChange
}

func (c *Config) GetResults() []*CheckResult {
	c.mu.Lock()
	defer c.mu.Unlock()

	svcs := make([]*CheckResult, len(c.services))
	count := 0

	for _, s := range c.services {
		svcs[count] = &CheckResult{
			Interval: s.Interval.Duration.Seconds(),
			Name:     s.Name,
			State:    s.state,
			Output:   s.output,
			Type:     s.Type,
			Time:     s.lastCheck,
			Since:    s.since,
		}
		count++
	}

	return svcs
}

// SendResults sends a set of Results to Notifiarr.
func (c *Config) SendResults(url string, results *Results) {
	results.Type = NotifiarrEventType
	results.Interval = c.Interval.Seconds()

	data, _ := json.MarshalIndent(results, "", " ")
	if _, err := c.Notify.SendJSON(url, data); err != nil {
		c.Errorf("Sending service check update to %s: %v", url, err)
	} else {
		c.Printf("Sent %d service check states to %s", len(results.Svcs), url)
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
