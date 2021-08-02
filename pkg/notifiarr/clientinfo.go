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

// ClientInfo is the reply from the ClienRoute endpoint.
type ClientInfo struct {
	Status  string `json:"status"`
	Message struct {
		Text       string `json:"text"`
		Subscriber bool   `json:"subscriber"`
		Patron     bool   `json:"patron"`
		CFSync     int64  `json:"cfSync"`
		RPSync     int64  `json:"rpSync"`
	} `json:"message"`
}

// String returns the message text for a client info response.
func (c *ClientInfo) String() string {
	if c == nil {
		return ""
	}

	return c.Message.Text
}

// IsSub returns true if the client is a subscriber. False otherwise.
func (c *ClientInfo) IsSub() bool {
	return c != nil && c.Message.Subscriber
}

// IsPatron returns true if the client is a patron. False otherwise.
func (c *ClientInfo) IsPatron() bool {
	return c != nil && c.Message.Patron
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
		"cfsync_dur":  cfSyncTimer.Seconds(),
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
