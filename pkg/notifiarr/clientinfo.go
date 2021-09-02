package notifiarr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/shirou/gopsutil/v3/host"
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
		Plex       *plexConfig      `json:"plex"`      // unused yet
		StuckItems stuckConfig      `json:"stuck"`     // unused yet!
		Dashboard  dashConfig       `json:"dashboard"` // now in use.
		Sync       syncConfig       `json:"sync"`      // in use (cfsync)
		Gaps       gapsConfig       `json:"gaps"`      // radarr collection gaps
		Custom     []*timer         `json:"custom"`    // custom GET timers
		Snapshot   *snapshot.Config `json:"snapshot"`  // unused
	} `json:"actions"`
}

type timer struct {
	Name    string `json:"name"`     // name of action.
	Minutes int    `json:"timer"`    // how often to GET this URI.
	URI     string `json:"endpoint"` // endpoint for the URI.
	last    time.Time
}

func (t *timer) Ready() bool {
	return t.last.After(time.Now().Add(time.Duration(t.Minutes) * time.Minute))
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
func (c *Config) GetClientInfo() (*ClientInfo, error) {
	c.extras.ciMutex.Lock()
	defer c.extras.ciMutex.Unlock()

	if c.extras.clientInfo != nil {
		return c.extras.clientInfo, nil
	}

	resp, body, err := c.SendData(c.BaseURL+ClientRoute, c.Info(), true) //nolint:bodyclose // already closed.
	if err != nil {
		return nil, fmt.Errorf("POSTing client info: %w", err)
	}

	v := clientInfoResponse{}
	if err = json.Unmarshal(body, &v); err != nil {
		return &v.ClientInfo, fmt.Errorf("parsing response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &v.ClientInfo, ErrNon200
	}

	// Only set this if there was no error.
	c.extras.clientInfo = &v.ClientInfo

	return c.extras.clientInfo, nil
}

// Info is used for JSON input for our outgoing client info.
func (c *Config) Info() map[string]interface{} {
	numPlex := 0 // maybe one day we'll support more than 1 plex.
	if c.Plex.Configured() {
		numPlex = 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	hostInfo, _ := host.InfoWithContext(ctx)
	if hostInfo != nil {
		hostInfo.Hostname = "" // we do not need this.
	}

	return map[string]interface{}{
		"arch":        runtime.GOARCH,
		"build_date":  version.BuildDate,
		"docker":      os.Getenv("NOTIFIARR_IN_DOCKER") == "true",
		"go_version":  version.GoVersion,
		"gui":         ui.HasGUI(),
		"host":        hostInfo,
		"num_deluge":  len(c.Apps.Deluge),
		"num_lidarr":  len(c.Apps.Lidarr),
		"num_plex":    numPlex,
		"num_qbit":    len(c.Apps.Qbit),
		"num_radarr":  len(c.Apps.Radarr),
		"num_readarr": len(c.Apps.Readarr),
		"num_sonarr":  len(c.Apps.Sonarr),
		"os":          runtime.GOOS,
		"retries":     c.Retries,
		"revision":    version.Revision,
		"snapshots":   c.Snap,
		"stuck_dur":   stuckTimer.Seconds(),
		"timeout":     c.Timeout.Seconds(),
		"uptime_dur":  time.Since(version.Started).Round(time.Second).Seconds(),
		"version":     version.Version,
	}
}
