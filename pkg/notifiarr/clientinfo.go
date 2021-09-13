package notifiarr

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/update"
	"github.com/denisbrodbeck/machineid"
	"github.com/shirou/gopsutil/v3/host"
	"golift.io/cnfg"
	"golift.io/version"
)

// clientInfoResponse is the reply from the ClientRoute endpoint.
type clientInfoResponse struct {
	Response   string     `json:"response"` // success
	ClientInfo ClientInfo `json:"message"`
}

type ClientInfo struct {
	User struct {
		WelcomeMSG string `json:"welcome"`
		Subscriber bool   `json:"subscriber"`
		Patron     bool   `json:"patron"`
	} `json:"user"`
	Actions struct {
		Poll      bool             `json:"poll"`
		Plex      *plex.Server     `json:"plex"`      // optional
		Apps      appConfigs       `json:"apps"`      // unused yet!
		Dashboard dashConfig       `json:"dashboard"` // now in use.
		Sync      syncConfig       `json:"sync"`      // in use (cfsync)
		Gaps      gapsConfig       `json:"gaps"`      // radarr collection gaps
		Custom    []*timerConfig   `json:"custom"`    // custom GET timers
		Snapshot  *snapshot.Config `json:"snapshot"`  // optional
	} `json:"actions"`
}

// ServiceConfig comes from the services package. It's only used for display on the website.
type ServiceConfig struct {
	Interval cnfg.Duration
	Parallel int
	Disabled bool
	Checks   []*ServiceCheck
}

// ServiceCheck comes from the services package. It's only used for display on the website.
type ServiceCheck struct {
	Name     string        `json:"name"`
	Type     string        `json:"type"`
	Expect   string        `json:"expect"`
	Timeout  cnfg.Duration `json:"timeout"`
	Interval cnfg.Duration `json:"interval"`
}

type intList []int

func (l intList) Has(instance int) bool {
	for _, i := range l {
		if instance == i {
			return true
		}
	}

	return false
}

// String returns the message text for a client info response.
func (c *ClientInfo) String() string {
	if c == nil {
		return "<nil>"
	}

	return c.User.WelcomeMSG
}

// IsSub returns true if the client is a subscriber. False otherwise.
func (c *ClientInfo) IsSub() bool {
	return c != nil && c.User.Subscriber
}

// IsPatron returns true if the client is a patron. False otherwise.
func (c *ClientInfo) IsPatron() bool {
	return c != nil && c.User.Patron
}

// GetClientInfo returns an error if the API key is wrong. Returns client info otherwise.
func (c *Config) GetClientInfo(source EventType) (*ClientInfo, error) {
	c.extras.ciMutex.Lock()
	defer c.extras.ciMutex.Unlock()

	if c.extras.ClientInfo != nil {
		return c.extras.ClientInfo, nil
	}

	info, err := c.Info()
	if err != nil {
		return nil, fmt.Errorf("getting system info: %w", err)
	}

	body, err := c.SendData(ClientRoute.Path(source), info, true)
	if err != nil {
		return nil, fmt.Errorf("sending client info: %w", err)
	}

	v := clientInfoResponse{}
	if err = json.Unmarshal(body, &v); err != nil {
		return &v.ClientInfo, fmt.Errorf("parsing response: %w", err)
	}

	// Only set this if there was no error.
	c.extras.ClientInfo = &v.ClientInfo

	return c.extras.ClientInfo, nil
}

// Info is used for JSON input for our outgoing client info.
func (c *Config) Info() (map[string]interface{}, error) {
	var (
		plexConfig interface{}
		numPlex    = 0 // maybe one day we'll support more than 1 plex.
	)

	if c.Plex.Configured() {
		numPlex = 1
		plexConfig = map[string]interface{}{
			"seriesPc":   c.Plex.SeriesPC,
			"moviesPc":   c.Plex.MoviesPC,
			"cooldown":   c.Plex.Cooldown,
			"accountMap": c.Plex.AccountMap,
			"interval":   c.Plex.Interval,
			"noActivity": c.Plex.NoActivity,
		}
	}

	hostInfo, err := c.GetHostInfoUID()

	return map[string]interface{}{
		"client": map[string]interface{}{
			"arch":      runtime.GOARCH,
			"buildDate": version.BuildDate,
			"goVersion": version.GoVersion,
			"os":        runtime.GOOS,
			"revision":  version.Revision,
			"version":   version.Version,
			"uptimeSec": time.Since(version.Started).Round(time.Second).Seconds(),
			"started":   version.Started,
			"docker":    os.Getenv("NOTIFIARR_IN_DOCKER") == "true",
			"gui":       ui.HasGUI(),
		},
		"host": hostInfo,
		"num": map[string]interface{}{
			"deluge":  len(c.Apps.Deluge),
			"lidarr":  len(c.Apps.Lidarr),
			"plex":    numPlex,
			"qbit":    len(c.Apps.Qbit),
			"radarr":  len(c.Apps.Radarr),
			"readarr": len(c.Apps.Readarr),
			"sonarr":  len(c.Apps.Sonarr),
		},
		"config": map[string]interface{}{
			"globalTimeout": c.Timeout.String(),
			"retries":       c.Retries,
			"plex":          plexConfig,
			"snapshots":     c.Snap,
			"apps":          c.getAppConfigs(),
		},
		"internal": map[string]interface{}{
			"stuckDur": stuckDur.String(),
			"pollDur":  pollDur.String(),
		},
		"services": c.Services,
	}, err
}

// GetHostInfoUID attempts to make a unique machine identifier...
func (c *Config) GetHostInfoUID() (*host.InfoStat, error) {
	c.extras.hiMutex.Lock()
	defer c.extras.hiMutex.Unlock()

	if c.extras.hostInfo != nil {
		return c.extras.hostInfo, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) //nolint:gomnd
	defer cancel()

	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting host info: %w", err)
	}

	uid, err := machineid.ProtectedID(hostInfo.Hostname)
	if err != nil {
		return nil, fmt.Errorf("getting machine ID: %w", err)
	}

	hostInfo.HostID = uid // this is where we put the unique ID.

	return hostInfo, nil
}

func (c *Config) pollForReload() {
	if c.ClientInfo != nil && !c.ClientInfo.Actions.Poll {
		return
	}

	info, err := c.Info()
	if err != nil {
		c.Errorf("Getting System Info: %v", err)
		return
	}

	body, err := c.SendData(ClientRoute.Path(EventPoll), info, true)
	if err != nil {
		c.Errorf("Polling Notifiarr: %v", err)
		return
	}

	var v struct {
		Reload bool `json:"reload"`
	}

	if err = json.Unmarshal(body, &v); err != nil {
		c.Errorf("Polling Notifiarr: %v", err)
		return
	}

	if v.Reload {
		c.Printf("Website indicated new configurations; reloading to pick them up!")
		c.Sighup <- &update.Signal{Text: "poll triggered reload"}
	} else if c.ClientInfo == nil {
		c.Printf("API Key checked out, reloading to pick up configuration from website!")
		c.Sighup <- &update.Signal{Text: "client info reload"}
	}
}

func (c *Config) getAppConfigs() interface{} {
	apps := make(map[string][]map[string]interface{})

	for i, app := range c.Apps.Lidarr {
		apps["lidarr"] = append(apps["lidarr"], map[string]interface{}{
			"name":     app.Name,
			"instance": i + 1,
			"checkQ":   app.CheckQ,
			"stuckOn":  app.StuckItem,
			"interval": app.Interval,
		})
	}

	for i, app := range c.Apps.Radarr {
		apps["radarr"] = append(apps["radarr"], map[string]interface{}{
			"name":     app.Name,
			"instance": i + 1,
			"checkQ":   app.CheckQ,
			"stuckOn":  app.StuckItem,
			"interval": app.Interval,
		})
	}

	for i, app := range c.Apps.Readarr {
		apps["readarr"] = append(apps["readarr"], map[string]interface{}{
			"name":     app.Name,
			"instance": i + 1,
			"checkQ":   app.CheckQ,
			"stuckOn":  app.StuckItem,
			"interval": app.Interval,
		})
	}

	for i, app := range c.Apps.Sonarr {
		apps["sonarr"] = append(apps["sonarr"], map[string]interface{}{
			"name":     app.Name,
			"instance": i + 1,
			"checkQ":   app.CheckQ,
			"stuckOn":  app.StuckItem,
			"interval": app.Interval,
		})
	}

	return apps
}
