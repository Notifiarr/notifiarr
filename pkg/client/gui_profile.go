package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/frontend"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/private"
	"github.com/Notifiarr/notifiarr/pkg/services"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/triggers/commands/cmdconfig"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/crontimer"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"github.com/shirou/gopsutil/v4/host"
	mulery "golift.io/mulery/client"
	"golift.io/version"
)

// Profile is the data returned by the profile GET endpoint.
// Basically everything.
type Profile struct {
	Username        string                 `json:"username"`
	Config          configfile.Config      `json:"config"`
	ClientInfo      *clientinfo.ClientInfo `json:"clientInfo"`
	IsWindows       bool                   `json:"isWindows"`
	IsLinux         bool                   `json:"isLinux"`
	IsDarwin        bool                   `json:"isDarwin"`
	IsDocker        bool                   `json:"isDocker"`
	IsUnstable      bool                   `json:"isUnstable"`
	IsFreeBSD       bool                   `json:"isFreeBsd"`
	IsSynology      bool                   `json:"isSynology"`
	Headers         http.Header            `json:"headers"`
	Fortune         string                 `json:"fortune"`
	UpstreamIP      string                 `json:"upstreamIp"`
	UpstreamAllowed bool                   `json:"upstreamAllowed"`
	UpstreamHeader  string                 `json:"upstreamHeader"`
	UpstreamType    configfile.AuthType    `json:"upstreamType"`
	Languages       frontend.Languages     `json:"languages"`
	Triggers        []common.TriggerInfo   `json:"triggers"`
	Timers          []common.TriggerInfo   `json:"timers"`
	Schedules       []common.TriggerInfo   `json:"schedules"`
	SiteCrons       []*crontimer.Timer     `json:"siteCrons"`
	PlexInfo        *plex.PMSInfo          `json:"plexInfo"`
	PlexAge         time.Time              `json:"plexAge"`
	// LoggedIn is only used by the front end. Backend does not set or use it.
	LoggedIn        bool                           `json:"loggedIn"`
	Updated         time.Time                      `json:"updated"`
	Flags           *configfile.Flags              `json:"flags"`
	Dynamic         bool                           `json:"dynamic"`
	Webauth         bool                           `json:"webauth"`
	Msg             string                         `json:"msg,omitempty"`
	LogFiles        *logs.LogFileInfos             `json:"logFileInfo"`
	ConfigFiles     *logs.LogFileInfos             `json:"configFileInfo"`
	Expvar          mnd.AllData                    `json:"expvar"`
	HostInfo        *host.InfoStat                 `json:"hostInfo"`
	Disks           map[string]*snapshot.Partition `json:"disks"`
	ProxyAllow      bool                           `json:"proxyAllow"`
	Started         time.Time                      `json:"started"`
	CmdList         []*cmdconfig.Config            `json:"cmdList"`
	Program         string                         `json:"program"`
	Version         string                         `json:"version"`
	Revision        string                         `json:"revision"`
	Branch          string                         `json:"branch"`
	BuildUser       string                         `json:"buildUser"`
	BuildDate       string                         `json:"buildDate"`
	GoVersion       string                         `json:"goVersion"`
	OS              string                         `json:"os"`
	Arch            string                         `json:"arch"`
	Binary          string                         `json:"binary"`
	Environment     map[string]string              `json:"environment"`
	Docker          bool                           `json:"docker"`
	UID             int                            `json:"uid"`
	GID             int                            `json:"gid"`
	IP              string                         `json:"ip"`
	Gateway         string                         `json:"gateway"`
	IfName          string                         `json:"ifName"`
	Netmask         string                         `json:"netmask"`
	MD5             string                         `json:"md5"`
	ActiveTunnel    string                         `json:"activeTunnel"`
	TunnelPoolStats map[string]*mulery.PoolSize    `json:"tunnelPoolStats"`
}

// handleProfile returns the current user's username in a JSON response.
//
//	@Summary		Get user profile
//	@Description	Returns comprehensive profile information including config, triggers, system info, and user settings.
//	@Tags			System
//	@Produce		json
//	@Success		200	{object}	Profile	"comprehensive profile data"
//	@Failure		401	{string}	string	"unauthorized"
//	@Router			/profile [get]
//
//nolint:funlen
func (c *Client) handleProfile(resp http.ResponseWriter, req *http.Request) {
	profile := &Profile{
		Config:     *c.Config,
		IsWindows:  mnd.IsWindows,
		IsLinux:    mnd.IsLinux,
		IsDarwin:   mnd.IsDarwin,
		IsFreeBSD:  mnd.IsFreeBSD,
		IsDocker:   mnd.IsDocker,
		IsUnstable: mnd.IsUnstable,
		IsSynology: mnd.IsSynology,
		Flags:      c.Flags,
		Webauth:    c.webauth,
		UpstreamIP: strings.Trim(req.RemoteAddr[:strings.LastIndex(req.RemoteAddr, ":")], "[]"),
		Version:    version.Version,
		Revision:   version.Revision,
		Branch:     version.Branch,
		BuildUser:  version.BuildUser,
		BuildDate:  version.BuildDate,
		GoVersion:  version.GoVersion,
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		Docker:     mnd.IsDocker,
	}

	profile.UpstreamAllowed = c.allow.Contains(req.RemoteAddr)
	profile.UpstreamHeader = c.Config.UIPassword.Header()
	profile.UpstreamType = c.Config.UIPassword.Type()
	profile.Updated = time.Now().UTC()
	profile.Languages = frontend.Translations()
	profile.ProxyAllow = c.allow.Contains(req.RemoteAddr)
	profile.Headers = c.getProfileHeaders(req)
	profile.Environment = environ()
	profile.Fortune = Fortune()
	profile.Username, profile.Dynamic = c.getUserName(req)
	profile.Binary, _ = os.Executable()
	profile.IP = clientinfo.GetOutboundIP(req.Context())
	profile.IfName, profile.Netmask = getIfNameAndNetmask(profile.IP)
	profile.LogFiles = logs.Log.GetAllLogFilePaths()
	profile.SiteCrons = c.triggers.CronTimer.List()
	profile.Expvar = mnd.GetAllData()
	profile.Started = version.Started.Round(time.Second)
	profile.CmdList = c.triggers.Commands.List()
	profile.Program = c.Flags.Name()
	profile.UID = os.Getuid()
	profile.GID = os.Getgid()
	profile.Gateway = getGateway()
	profile.MD5 = private.MD5()
	profile.ConfigFiles = logs.GetFilePaths(c.Flags.ConfigFile,
		filepath.Join(filepath.Dir(c.Flags.ConfigFile), "backups", filepath.Base(c.Flags.ConfigFile)))
	profile.HostInfo, _ = website.GetHostInfo(req.Context())
	profile.Triggers, profile.Timers, profile.Schedules = c.triggers.GatherTriggerInfo()
	profile.Disks = getDisks(req.Context(), c.Config.Snapshot.ZFSPools)

	profile.ClientInfo = clientinfo.Get()
	if profile.ClientInfo == nil {
		profile.ClientInfo = &clientinfo.ClientInfo{}
	}

	profile.PlexInfo = &plex.PMSInfo{}
	profile.PlexAge = time.Time{}
	if ps := data.Get("plexStatus"); ps != nil {
		profile.PlexAge = ps.Time
		profile.PlexInfo, _ = ps.Data.(*plex.PMSInfo)
	}

	if at := data.Get("activeTunnel"); at != nil {
		profile.ActiveTunnel, _ = at.Data.(string)
	}

	if c.tunnel != nil {
		profile.TunnelPoolStats = c.tunnel.PoolStats()
	}

	resp.Header().Set("Content-Type", mnd.ContentTypeJSON)

	if err := json.NewEncoder(resp).Encode(profile); err != nil {
		logs.Log.Errorf(mnd.GetID(req.Context()), "Writing HTTP Response: %v", err)
	}
}

// handleProfileNoAPIKey handles a minimal profile response for the UI when no API key is set.
func (c *Client) handleProfileNoAPIKey(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", mnd.ContentTypeJSON)
	profile := &Profile{Updated: time.Now().UTC()}
	profile.Config.APIKey = c.Config.APIKey
	profile.Config.URLBase = c.Config.URLBase

	if err := json.NewEncoder(resp).Encode(profile); err != nil {
		logs.Log.Errorf(mnd.GetID(req.Context()), "Writing HTTP Response: %v", err)
	}
}

// handleProfilePost handles profile updates including authentication settings and upstream configuration.
//
//	@Summary		Update user profile
//	@Description	Updates login credentials, authentication settings, and upstream trusted networks.
//	@Tags			System
//	@Accept			json
//	@Produce		text/plain
//	@Param			profile	body		ProfilePost	true	"Profile update data"
//	@Success		200		{string}	string		"success message"
//	@Failure		400		{string}	string		"invalid request or password"
//	@Failure		500		{string}	string		"error saving config"
//	@Router			/profile [post]
//
//nolint:cyclop
func (c *Client) handleProfilePost(response http.ResponseWriter, request *http.Request) {
	post := &ProfilePost{}
	if err := json.NewDecoder(request.Body).Decode(post); err != nil {
		http.Error(response, "Invalid Request", http.StatusBadRequest)
		return
	}

	currUser, dynamic := c.getUserName(request)
	if !dynamic {
		// If the auth is currently using a password, check the password.
		if !c.Config.UIPassword.Valid(currUser, post.Password) &&
			(c.Config.UIPassword.IsCrypted() || !clientinfo.CheckPassword(currUser, post.Password)) {
			logs.Log.Errorf(mnd.GetID(request.Context()),
				"[gui '%s' requested] Trust Profile: Invalid existing (current) password provided.", currUser)
			http.Error(response, "Invalid existing (current) password provided.", http.StatusBadRequest)
			return
		}
	}

	// Upstreams is only read on reload, but this is still not thread safe
	// because two people could click save at the same time.
	c.Lock()
	c.Config.Upstreams = strings.Fields(post.Upstreams)
	c.Unlock()

	if post.NewPass == "" {
		post.NewPass = post.Password
	}

	if post.AuthType == configfile.AuthPassword {
		c.handleProfilePostPassword(response, request, post.Username, post.NewPass)
		return
	}

	switch err := c.setUserPass(request.Context(), post.AuthType, post.Header, ""); {
	case err != nil:
		logs.Log.Errorf(mnd.GetID(request.Context()), "[gui '%s' requested] Saving Config: %v", currUser, err)
		http.Error(response, "Saving Config: "+err.Error(), http.StatusInternalServerError)
	case post.AuthType == configfile.AuthNone:
		logs.Log.Printf(mnd.GetID(request.Context()), "[gui '%s' requested] Disabled WebUI authentication.", currUser)
		http.Error(response, "Disabled WebUI authentication.", http.StatusOK)
		c.reloadAppNow()
	default:
		logs.Log.Printf(mnd.GetID(request.Context()),
			"[gui '%s' requested] Enabled WebUI proxy authentication, header: %s", currUser, post.Header)
		c.setSession(post.Username, response, request)
		http.Error(response, "Enabled WebUI proxy authentication. Header: "+post.Header, http.StatusOK)
		c.reloadAppNow()
	}
}

// ProfilePost is the data sent to the profile POST endpoint when updating the trust profile.
type ProfilePost struct {
	Username  string              `json:"username"`
	Password  string              `json:"password"`
	AuthType  configfile.AuthType `json:"authType"`
	Header    string              `json:"header"`
	NewPass   string              `json:"newPass"`
	Upstreams string              `json:"upstreams"`
}

func (c *Client) handleProfilePostPassword(
	response http.ResponseWriter,
	request *http.Request,
	newUser, newPassw string,
) {
	currUser, _ := c.getUserName(request)

	if err := c.setUserPass(request.Context(), configfile.AuthPassword, newUser, newPassw); err != nil {
		logs.Log.Errorf(mnd.GetID(request.Context()), "[gui '%s' requested] Saving Trust Profile: %v", currUser, err)
		http.Error(response, "Saving Trust Profile: "+err.Error(), http.StatusInternalServerError)

		return
	}

	logs.Log.Printf(mnd.GetID(request.Context()),
		"[gui '%s' requested] Updated Trust Profile settings, username: %s", currUser, newUser)
	c.setSession(newUser, response, request)
	http.Error(response, "Trust Profile saved.", http.StatusOK)
	c.reloadAppNow()
}

//nolint:funlen // is what it is.
func (c *Client) getProfileHeaders(req *http.Request) http.Header {
	// ignoredHeaders is a list of headers that should be filtered out from profile requests.
	// We display a list of headers on the Trust Profile page for the user to select
	// their auth header when configuring an auth proxy. We hide headers that we know
	// are not auth headers.
	ignoredHeaders := map[string]struct{}{
		"accept":                    {},
		"accept-encoding":           {},
		"accept-language":           {},
		"cache-control":             {},
		"content-length":            {},
		"content-type":              {},
		"cdn-loop":                  {},
		"cf-connecting-ip":          {},
		"cf-ipcity":                 {},
		"cf-ipcontinent":            {},
		"cf-ipcountry":              {},
		"cf-iplatitude":             {},
		"cf-iplongitude":            {},
		"cf-metro-code":             {},
		"cf-postal-code":            {},
		"cf-ray":                    {},
		"cf-region":                 {},
		"cf-region-code":            {},
		"cf-timezone":               {},
		"cf-visitor":                {},
		"connection":                {},
		"cookie":                    {},
		"dnt":                       {},
		"expect":                    {},
		"pragma":                    {},
		"priority":                  {},
		"referer":                   {},
		"sec-ch-ua":                 {},
		"sec-ch-ua-mobile":          {},
		"sec-ch-ua-platform":        {},
		"sec-fetch-dest":            {},
		"sec-fetch-mode":            {},
		"sec-fetch-site":            {},
		"strict-transport-security": {},
		"te":                        {},
		"upgrade-insecure-requests": {},
		"user-agent":                {},
		"x-content-type-options":    {},
		"x-forwarded-for":           {},
		"x-forwarded-host":          {},
		"x-forwarded-method":        {},
		"x-forwarded-port":          {},
		"x-forwarded-proto":         {},
		"x-forwarded-server":        {},
		"x-forwarded-ssl":           {},
		"x-forwarded-uri":           {},
		"x-noticlient-username":     {},
		"x-original-method":         {},
		"x-original-uri":            {},
		"x-original-url":            {},
		"x-real-ip":                 {},
		"x-redacted-uri":            {},
		"x-request-id":              {},
	}
	headers := http.Header{}

	for name, values := range req.Header {
		if _, ok := ignoredHeaders[strings.ToLower(name)]; !ok {
			for _, value := range values {
				headers.Add(name, value)
			}
		}
	}

	return headers
}

// ServicesConfig is the data returned by the services/config GET endpoint.
type ServicesConfig struct {
	Results  []*services.CheckResult `json:"results"`
	Running  bool                    `json:"running"`
	Disabled bool                    `json:"disabled"`
}

// handleServicesConfig returns the services config information including results and running status.
//
//	@Summary		Get services config
//	@Description	Returns services config information including results and running status.
//	@Tags			System
//	@Produce		json
//	@Success		200	{object}	ServicesConfig	"services config data"
//	@Failure		401	{string}	string	"unauthorized"
//	@Router			/services/config [get]
func (c *Client) handleServicesConfig(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", mnd.ContentTypeJSON)

	err := json.NewEncoder(resp).Encode(&ServicesConfig{
		Results:  c.Services.GetResults(),
		Running:  c.Services.Running(),
		Disabled: c.Services.Disabled,
	})
	if err != nil {
		logs.Log.Errorf(mnd.GetID(req.Context()), "Writing HTTP Response: %v", err)
	}
}

func (c *Client) setUserPass(ctx context.Context, authType configfile.AuthType, username, password string) error {
	c.Lock()
	defer c.Unlock()

	current := c.Config.UIPassword

	var err error

	switch authType {
	case configfile.AuthPassword:
		err = c.Config.UIPassword.Set(username, password)
	case configfile.AuthHeader:
		err = c.Config.UIPassword.SetHeader(username)
	case configfile.AuthNone:
		err = c.Config.UIPassword.SetNoAuth(username)
	case configfile.AuthWebsite:
		err = c.Config.UIPassword.Set("", "")
	}

	if err != nil {
		c.Config.UIPassword = current
		return fmt.Errorf("saving new auth settings: %w", err)
	}

	config, err := c.Config.CopyConfig()
	if err != nil {
		return fmt.Errorf("copying config: %w", err)
	}

	if err := c.saveNewConfig(ctx, config); err != nil {
		c.Config.UIPassword = current
		return err
	}

	return nil
}
