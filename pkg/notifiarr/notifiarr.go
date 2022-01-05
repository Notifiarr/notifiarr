// Package notifiarr provides a standard interface for sending data to notifiarr.com.
// Several methods are exported to make POSTing data to notifarr easier. This package
// also handles the incoming Plex webhook as well as the "crontab" timers for plex
// sessions, snapshots, dashboard state, custom format sync for Radarr and release
// profile sync for Sonarr.
// This package's cofiguration is provided by the configfile  package.
package notifiarr

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/shirou/gopsutil/v3/host"
	"golift.io/cnfg"
)

// Errors returned by this library.
var (
	ErrNon200          = fmt.Errorf("return code was not 200")
	ErrInvalidResponse = fmt.Errorf("invalid response")
	ErrInvalidApp      = fmt.Errorf("invalid application provided")
)

const (
	ModeDev  = "development"
	ModeProd = "production"
)

const (
	// DefaultRetries is the number of times to attempt a request to notifiarr.com.
	// 4 means 5 total tries: 1 try + 4 retries.
	DefaultRetries = 4
	// RetryDelay is how long to Sleep between retries.
	RetryDelay = 222 * time.Millisecond
)

// Config is the input data needed to send payloads to notifiarr.
type Config struct {
	Apps     *apps.Apps       // has API key
	Plex     *plex.Server     // plex sessions
	Snap     *snapshot.Config // system snapshot data
	Services *ServiceConfig
	Retries  int
	BaseURL  string
	Timeout  cnfg.Duration
	Trigger  Triggers
	MaxBody  int
	Sighup   chan os.Signal

	*logs.Logger // log file writer
	extras
}

type extras struct {
	ciMutex sync.Mutex
	*ClientInfo
	client    *httpClient
	radarrCF  map[int]*cfMapIDpayload
	sonarrRP  map[int]*cfMapIDpayload
	plexTimer *Timer
	hostInfo  *host.InfoStat
}

// Triggers allow trigger actions in the timer routine.
type Triggers struct {
	stop  *action        // Triggered by calling Stop()
	sess  chan time.Time // Return Plex Sessions
	sessr chan *holder   // Session Return Channel
	List  []*action      // List of action triggers
}

// Start (and log) snapshot and plex cron jobs if they're configured.
func (c *Config) Setup(mode string) string {
	switch strings.ToLower(mode) {
	default:
		c.BaseURL = BaseURL
	case "prod", ModeProd:
		c.BaseURL = BaseURL
		mode = ModeProd
	case "dev", "devel", ModeDev, "test", "testing":
		c.BaseURL = DevBaseURL
		mode = ModeDev
	}

	if c.Retries < 0 {
		c.Retries = 0
	} else if c.Retries == 0 {
		c.Retries = DefaultRetries
	}

	if c.extras.client == nil {
		c.extras.client = &httpClient{
			Retries: c.Retries,
			Logger:  c.ErrorLog,
			Client:  &http.Client{},
		}
	}

	return mode
}

// Start runs the timers.
func (c *Config) Start() {
	if c.Trigger.stop != nil {
		panic("notifiarr timers cannot run more than once")
	}

	c.Trigger.sess = make(chan time.Time, 1)
	c.extras.radarrCF = make(map[int]*cfMapIDpayload)
	c.extras.sonarrRP = make(map[int]*cfMapIDpayload)
	c.extras.plexTimer = &Timer{}

	// Order is important here.
	go c.runSessionHolder()
	c.logSnapshotStartup()
	c.makeBaseTriggers()
	c.makeCorruptionTriggers()
	c.makeBackupTriggers()
	c.setPlexTimers()

	if _, err := c.GetClientInfo(EventStart); err == nil { // only run this if we have clientinfo
		c.setClientInfoTimerTriggers() // sync, gaps, dashboard, custom
	}

	c.makeCustomClientInfoTimerTriggers()
	c.runTimers()
}

func (c *Config) makeBaseTriggers() {
	c.Trigger.stop = &action{
		Name: TrigStop,
		C:    make(chan EventType),
	}
	c.Trigger.add(&action{
		Name: TrigStuckItems,
		Fn:   c.sendStuckQueueItems,
		C:    make(chan EventType, 1),
		T:    time.NewTicker(stuckDur),
	}, &action{
		Name: TrigPlexSessions,
		Fn:   c.sendPlexSessions,
		C:    make(chan EventType, 1),
	}, &action{
		Name: TrigCollectionGaps,
		Fn:   c.sendGaps,
		C:    make(chan EventType, 1),
	}, &action{
		Name: TrigCFSync,
		Fn:   c.syncCF,
		C:    make(chan EventType, 1),
	}, &action{
		Name: TrigDashboard,
		Fn:   c.sendDashboardState,
		C:    make(chan EventType, 1),
	}, &action{
		Name: TrigSnapshot,
		T:    tickerOrNil(c.Snap.Interval.Duration),
		Fn:   c.sendSnapshot,
		C:    make(chan EventType, 1),
	})
}

func (c *Config) setPlexTimers() {
	if !c.Plex.Configured() {
		return
	}

	if c.Plex.Interval.Duration > 0 {
		// Add a little splay to the timers to not hit plex at the same time too often.
		c.Printf("==> Plex Sessions Collection Started, URL: %s, interval:%s timeout:%s webhook_cooldown:%v delay:%v",
			c.Plex.URL, c.Plex.Interval, c.Plex.Timeout, c.Plex.Cooldown, c.Plex.Delay)
		c.Trigger.get(TrigPlexSessions).T = time.NewTicker(c.Plex.Interval.Duration + 139*time.Millisecond) // nolint:wsl
	}

	if c.Plex.MoviesPC != 0 || c.Plex.SeriesPC != 0 {
		c.Printf("==> Plex Completed Items Started, URL: %s, interval:1m timeout:%s movies:%d%% series:%d%%",
			c.Plex.URL, c.Plex.Timeout, c.Plex.MoviesPC, c.Plex.SeriesPC)
		c.Trigger.add( // this has no name, which keeps it from logging _every minute_
			&action{SFn: c.checkPlexFinishedItems, T: time.NewTicker(time.Minute + 179*time.Millisecond)})
	}
}

func (c *Config) setClientInfoTimerTriggers() {
	if c.Actions.Gaps.Interval.Duration > 0 && len(c.Apps.Radarr) > 0 {
		c.Trigger.get(TrigCollectionGaps).T = time.NewTicker(c.Actions.Gaps.Interval.Duration)
		c.Printf("==> Collection Gaps Timer Enabled, interval:%s", c.Actions.Gaps.Interval)
	}

	if c.Actions.Sync.Interval.Duration > 0 && (len(c.Apps.Radarr) > 0 || len(c.Apps.Sonarr) > 0) {
		c.Trigger.get(TrigCFSync).T = time.NewTicker(c.Actions.Sync.Interval.Duration)
		c.Printf("==> Keeping %d Radarr Custom Formats and %d Sonarr Release Profiles synced, interval:%s",
			c.Actions.Sync.Radarr, c.Actions.Sync.Sonarr, c.Actions.Sync.Interval)
	}

	if c.Actions.Dashboard.Interval.Duration > 0 {
		c.Trigger.get(TrigDashboard).T = time.NewTicker(c.Actions.Dashboard.Interval.Duration)
		c.Printf("==> Sending Current State Data for Dashboard, interval:%s", c.Actions.Dashboard.Interval)
	}

	if len(c.Actions.Custom) > 0 { // This is not directly triggerable.
		c.Printf("==> Custom Timers Enabled: %d timers provided", len(c.Actions.Custom))
	}
}

func (c *Config) makeCustomClientInfoTimerTriggers() {
	// This poller is sorta shoehorned in here for lack of a better place to put it.
	if c.ClientInfo == nil || c.ClientInfo.Actions.Poll {
		c.Printf("==> Started Notifiarr Poller, have_clientinfo:%v interval:%s timeout:%s",
			c.ClientInfo != nil, cnfg.Duration{Duration: pollDur.Round(time.Second)}, c.Timeout)
		c.Trigger.add(&action{Name: TrigPollSite, Fn: c.pollForReload, T: time.NewTicker(pollDur)})
	}

	if c.ClientInfo == nil {
		return
	}

	for _, custom := range c.Actions.Custom {
		custom.setup(c)

		var ticker *time.Ticker

		if custom.Interval.Duration < time.Minute {
			c.Errorf("Website provided custom cron interval under 1 minute. Ignored! Interval: %s Name: %s, URI: %s",
				custom.Interval, custom.Name, custom.URI)
		} else {
			ticker = time.NewTicker(custom.Interval.Duration)
		}

		c.Trigger.add(&action{
			Name: TriggerName(fmt.Sprintf("Running Custom Cron Timer '%s' GET %s", custom.Name, custom.URI)),
			Fn:   custom.run,
			C:    custom.ch,
			T:    ticker,
		})
	}
}

// Stop all internal cron timers and Triggers.
func (c *Config) Stop(event EventType) {
	if c == nil {
		return
	}

	c.Print("==> Stopping Notifiarr Timers.")

	if c.Trigger.stop == nil {
		c.Error("==> Notifiarr Timers cannot be stopped: not running!")
		return
	}

	c.Trigger.stop.C <- event
	defer close(c.Trigger.sess)
	c.Trigger.sess = nil
}
