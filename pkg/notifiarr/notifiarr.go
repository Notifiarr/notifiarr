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
	"strings"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
)

// Errors returned by this library.
var (
	ErrNon200          = fmt.Errorf("return code was not 200")
	ErrInvalidResponse = fmt.Errorf("invalid response")
)

// Notifiarr URLs.
const (
	BaseURL     = "https://notifiarr.com"
	ProdURL     = BaseURL + "/notifier.php"
	TestURL     = BaseURL + "/notifierTest.php"
	DevBaseURL  = "http://dev.notifiarr.com"
	DevURL      = DevBaseURL + "/notifier.php"
	ClientRoute = "/api/v2/user/client"
	// CFSyncRoute is the webserver route to send sync requests to.
	CFSyncRoute = "/api/v1/user/trash"
	DashRoute   = "/api/v1/user/dashboard"
	GapsRoute   = "/api/v1/user/gaps"
)

// These are used as 'source' values in json payloads sent to the webserver.
const (
	PlexCron = "plexcron"
	SnapCron = "snapcron"
	PlexHook = "plexhook"
	LogLocal = "loglocal"
)

const (
	// DefaultRetries is the number of times to attempt a request to notifiarr.com.
	// 4 means 5 total tries: 1 try + 4 retries.
	DefaultRetries = 4
	// RetryDelay is how long to Sleep between retries.
	RetryDelay = 222 * time.Millisecond
)

// success is a ssuccessful tatus message from notifiarr.com.
const success = "success"

// Payload is the outbound payload structure that is sent to Notifiarr for Plex and system snapshot data.
type Payload struct {
	Type string               `json:"eventType"`
	Plex *plex.Sessions       `json:"plex,omitempty"`
	Snap *snapshot.Snapshot   `json:"snapshot,omitempty"`
	Load *plexIncomingWebhook `json:"payload,omitempty"`
}

// Config is the input data needed to send payloads to notifiarr.
type Config struct {
	Apps         *apps.Apps       // has API key
	Plex         *plex.Server     // plex sessions
	Snap         *snapshot.Config // system snapshot data
	Retries      int
	URL          string
	BaseURL      string
	Timeout      time.Duration
	Trigger      Triggers
	*logs.Logger // log file writer
	extras
}

type extras struct {
	ciMutex    sync.Mutex
	clientInfo *ClientInfo
	client     *httpClient
	radarrCF   map[int]*cfMapIDpayload
	sonarrRP   map[int]*cfMapIDpayload
	plexTimer  *Timer
}

// Triggers allow trigger actions in the timer routine.
type Triggers struct {
	stop   chan struct{}      // Triggered by calling Stop()
	syncCF chan chan struct{} // Sync Radarr CF and Sonarr RP
	gaps   chan string        // Send Radarr Collection Gaps
	stuck  chan string        // Stuck Items
	plex   chan string        // Send Plex Sessions
	state  chan struct{}      // Dashboard State
	snap   chan string        // Snapshot
	sess   chan time.Time     // Return Plex Sessions
	sessr  chan *holder       // Session Return Channel
}

// Start (and log) snapshot and plex cron jobs if they're configured.
func (c *Config) Setup(mode string) {
	switch strings.ToLower(mode) {
	default:
		fallthrough
	case "prod", "production":
		c.URL = ProdURL
		c.BaseURL = BaseURL
	case "test", "testing":
		c.URL = TestURL
		c.BaseURL = BaseURL
	case "dev", "devel", "development":
		c.URL = DevURL
		c.BaseURL = DevBaseURL
	}

	if c.Retries < 0 {
		c.Retries = 0
	} else if c.Retries == 0 {
		c.Retries = DefaultRetries
	}

	c.extras.client = &httpClient{
		Retries: c.Retries,
		Logger:  c.ErrorLog,
		Client:  &http.Client{},
	}
}

// Start runs the timers.
func (c *Config) Start() {
	if c.Trigger.stop != nil {
		panic("notifiarr timers cannot run twice")
	}

	c.extras.radarrCF = make(map[int]*cfMapIDpayload)
	c.extras.sonarrRP = make(map[int]*cfMapIDpayload)
	c.extras.plexTimer = &Timer{}
	c.Trigger.syncCF = make(chan chan struct{})
	c.Trigger.stuck = make(chan string)
	c.Trigger.plex = make(chan string)
	c.Trigger.state = make(chan struct{})
	c.Trigger.snap = make(chan string)
	c.Trigger.sess = make(chan time.Time)
	c.Trigger.gaps = make(chan string)

	go c.runSessionHolder()
	c.startTimers()
}

// Stop snapshot and plex cron jobs.
func (c *Config) Stop() {
	if c != nil && c.Trigger.stop != nil {
		c.Trigger.stop <- struct{}{}
		close(c.Trigger.syncCF)
		close(c.Trigger.stuck)
		close(c.Trigger.plex)
		close(c.Trigger.state)
		close(c.Trigger.snap)
		close(c.Trigger.gaps)

		defer close(c.Trigger.sess)
		c.Trigger.sess = nil
	}
}
