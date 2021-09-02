package notifiarr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/shirou/gopsutil/v3/host"
	"golift.io/version"
)

// ClientInfo is the reply from the ClientRoute endpoint.
type ClientInfo struct {
	User struct {
		WelcomeMSG string `json:"welcome"`
		Subscriber bool   `json:"subscriber"`
		Patron     bool   `json:"patron"`
	} `json:"user"`
	Actions struct {
		Sync struct {
			Minutes int    `json:"timer"`    // how often to fire in minutes.
			URI     string `json:"endpoint"` // "api/v1/user/sync"
			Radarr  int64  `json:"radarr"`   // items in sync
			Sonarr  int64  `json:"sonarr"`   // items in sync
		}
		Gaps   gaps `json:"gaps"`
		Custom []*timer
	}
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

// String returns the message text for a client info response.
func (c *ClientInfo) String() string {
	if c == nil {
		return ""
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
	if c.extras.clientInfo != nil {
		return c.extras.clientInfo, nil
	}

	resp, body, err := c.SendData(c.BaseURL+ClientRoute, c.Info(), true) //nolint:bodyclose // already closed.
	if err != nil {
		return nil, fmt.Errorf("POSTing client info: %w", err)
	}

	v := ClientInfo{}
	if err = json.Unmarshal(body, &v); err != nil {
		return &v, fmt.Errorf("parsing response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &v, ErrNon200
	}

	// Only set this if there was no error.
	c.extras.clientInfo = &v

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
		"dash_dur":    c.DashDur.Seconds(),
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
		"snap_dur":    c.Snap.Interval.Seconds(),
		"snap_tout":   c.Snap.Timeout.Seconds(),
		"stuck_dur":   stuckTimer.Seconds(),
		"timeout":     c.Timeout.Seconds(),
		"uptime_dur":  time.Since(version.Started).Round(time.Second).Seconds(),
		"version":     version.Version,
	}
}
