package clientinfo

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golang.org/x/crypto/bcrypt"
	"golift.io/cnfg"
)

// ClientInfo is the client's startup data received from the website.
type ClientInfo struct {
	User struct {
		// user id from notifiarr db.
		ID any `json:"id"`
		// Username is the username of the user's account on notifiarr.com. Sometimes it's empty.
		Username string `json:"username"`
		// Email is the md5 hashed email address of the user's account on notifiarr.com.
		Email string `json:"email"`
		// Password is the md5+bcrypt hashed password of the user's account on notifiarr.com.
		Password string `json:"hash"`
		// This is printed on startup and on the UI landing page.
		WelcomeMSG string `json:"welcome"`
		// Is the user a subscriber?
		Subscriber bool `json:"subscriber"`
		// Is the user a patron?
		Patron bool `json:"patron"`
		// Is the user allowed to use non-production website apis?
		DevAllowed bool `json:"devAllowed"`
		// This is the date format the user selected on the website.
		DateFormat PHPDate `json:"dateFormat"`
		// The website can use this to tell the client not to send any logs.
		StopLogs bool `json:"stopLogs"`
		// This is the URL the website uses to connect to the client.
		// It's just for info/debug here, and not used by the client.
		TunnelURL string `json:"tunnelUrl"`
		// This is the list of tunnels the website tells the client to connect to.
		Tunnels []string `json:"tunnels"`
		// List of tunnels that notifiarr.com recognizes.
		// Any of these may be used.
		Mulery []*MuleryServer `json:"mulery"`
	} `json:"user"`
	Actions        Actions `json:"actions"`
	IntegrityCheck bool    `json:"integrityCheck"`
}

// MuleryServer is data from the website. It's a tunnel's https and wss urls.
type MuleryServer struct {
	Tunnel   string `json:"tunnel"`   // ex: "https://africa.notifiarr.com/"
	Socket   string `json:"socket"`   // ex: "wss://africa.notifiarr.com/register"
	Location string `json:"location"` // ex: "Nairobi, Kenya, Africa"
}

type Actions struct {
	Plex      plex.WebsiteConfig `json:"plex"`      // Site Config for Plex.
	Apps      AllAppConfigs      `json:"apps"`      // Site Config for Starr.
	Dashboard DashConfig         `json:"dashboard"` // Site Config for Dashboard.
	Sync      SyncConfig         `json:"sync"`      // Site Config for TRaSH Sync.
	Mdblist   MdbListConfig      `json:"mdblist"`   // Site Config for MDB List.
	Gaps      GapsConfig         `json:"gaps"`      // Site Config for Radarr Gaps.
	Custom    []*CronConfig      `json:"custom"`    // Site config for Custom Crons.
	Snapshot  snapshot.Config    `json:"snapshot"`  // Site Config for System Snapshot.
}

// CronConfig defines a custom GET timer from the website.
// Used to offload crons to clients.
type CronConfig struct {
	Name     string        `json:"name"`     // name of action.
	Interval cnfg.Duration `json:"interval"` // how often to GET this URI.
	URI      string        `json:"endpoint"` // endpoint for the URI.
	Desc     string        `json:"description"`
}

// SyncConfig is the configuration returned from the notifiarr website for CF/RP TraSH sync.
type SyncConfig struct {
	Interval        cnfg.Duration `json:"interval"`        // how often to fire.
	LidarrInstances IntList       `json:"lidarrInstances"` // which instance IDs we sync
	RadarrInstances IntList       `json:"radarrInstances"` // which instance IDs we sync
	SonarrInstances IntList       `json:"sonarrInstances"` // which instance IDs we sync
	LidarrSync      []string      `json:"lidarrSync"`      // items in sync.
	SonarrSync      []string      `json:"sonarrSync"`      // items in sync.
	RadarrSync      []string      `json:"radarrSync"`      // items in sync.
}

// MdbListConfig contains the instances we send libraries for, and the interval we do it in.
type MdbListConfig struct {
	Interval cnfg.Duration `json:"interval"` // how often to fire.
	Radarr   IntList       `json:"radarr"`   // which instance IDs we sync
	Sonarr   IntList       `json:"sonarr"`   // which instance IDs we sync
}

// DashConfig is the configuration returned from the notifiarr website for the dashboard configuration.
type DashConfig struct {
	Interval     cnfg.Duration `json:"interval"` // how often to fire.
	Deluge       bool          `json:"deluge"`
	Lidarr       bool          `json:"lidarr"`
	Qbit         bool          `json:"qbit"`
	Radarr       bool          `json:"radarr"`
	Readarr      bool          `json:"readarr"`
	Sonarr       bool          `json:"sonarr"`
	Plex         bool          `json:"plex"`
	SabNZB       bool          `json:"sabnzbd"`
	NZBGet       bool          `json:"nzbget"`
	Rtorrent     bool          `json:"rtorrent"`
	Transmission bool          `json:"transmission"`
}

// AppConfig is the data that comes from the website for each Starr app.
type AppConfig struct {
	Instance int           `json:"instance"`
	Name     string        `json:"name"`
	Corrupt  string        `json:"corrupt"`
	Backup   string        `json:"backup"`
	Interval cnfg.Duration `json:"interval"`
	Stuck    bool          `json:"stuck"`
	Finished bool          `json:"finished"`
}

// InstanceConfig allows binding methods to a list of instance configurations.
type InstanceConfig []*AppConfig

// AllAppConfigs is the configuration returned from the notifiarr website for Starr apps.
type AllAppConfigs struct {
	Lidarr   InstanceConfig `json:"lidarr"`
	Prowlarr InstanceConfig `json:"prowlarr"`
	Radarr   InstanceConfig `json:"radarr"`
	Readarr  InstanceConfig `json:"readarr"`
	Sonarr   InstanceConfig `json:"sonarr"`
}

// GapsConfig is the configuration returned from the notifiarr website for Radarr Collection Gaps.
type GapsConfig struct {
	Instances IntList       `json:"instances"`
	Interval  cnfg.Duration `json:"interval"`
}

// IntList has a method to abstract lookups.
type IntList []int

// Has returns true if the list has an instance ID.
func (l IntList) Has(instance int) bool {
	return slices.Contains(l, instance)
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

// SaveClientInfo returns an error if the API key is wrong. Caches and returns client info otherwise.
func (c *Config) SaveClientInfo(ctx context.Context, startup bool) (*ClientInfo, error) {
	event := website.EventSignal
	if startup {
		event = website.EventStart
	}

	body, err := website.GetData(&website.Request{
		ReqID:      mnd.GetID(ctx),
		Route:      website.ClientRoute,
		Event:      event,
		Payload:    c.Info(ctx, startup),
		LogPayload: true,
	})
	if err != nil {
		return nil, fmt.Errorf("sending client info: %w", err)
	}

	clientInfo := ClientInfo{}
	if err = json.Unmarshal(body.Details.Response, &clientInfo); err != nil {
		return &clientInfo, fmt.Errorf("parsing response: %w, %s", err, string(body.Details.Response))
	}

	// Only set this if there was no error.
	data.Save("clientInfo", &clientInfo)

	return &clientInfo, nil
}

func Get() *ClientInfo {
	data := data.Get("clientInfo")
	if data == nil || data.Data == nil {
		return nil
	}

	cinfo, _ := data.Data.(*ClientInfo)

	return cinfo
}

func (i InstanceConfig) Finished(instance int) bool {
	for _, app := range i {
		if app.Instance == instance {
			return app.Finished
		}
	}

	return false
}

func (i InstanceConfig) Stuck(instance int) bool {
	for _, app := range i {
		if app.Instance == instance {
			return app.Stuck
		}
	}

	return false
}

func (i InstanceConfig) Backup(instance int) string {
	for _, app := range i {
		if app.Instance == instance {
			return app.Backup
		}
	}

	return mnd.Disabled
}

func (i InstanceConfig) Corrupt(instance int) string {
	for _, app := range i {
		if app.Instance == instance {
			return app.Corrupt
		}
	}

	return mnd.Disabled
}

// CheckPassword checks if a password is valid for a given username, using website credentials.
func CheckPassword(username, password string) bool {
	ci := Get() //nolint:varnamelen
	if ci == nil {
		return false
	}

	// The email address comes from the website as an md5 string.
	md5user := md5.Sum([]byte(username))
	if hex.EncodeToString(md5user[:]) != ci.User.Email && ci.User.Username != username {
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(ci.User.Password), []byte(password)) == nil
}
