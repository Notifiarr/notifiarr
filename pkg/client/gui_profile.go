package client

import (
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
	PoolStats       map[string]*mulery.PoolSize    `json:"poolStats"`
	Started         time.Time                      `json:"started"`
	CmdList         []*cmdconfig.Config            `json:"cmdList"`
	CheckResults    []*services.CheckResult        `json:"checkResults"`
	CheckRunning    bool                           `json:"checkRunning"`
	CheckDisabled   bool                           `json:"checkDisabled"`
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
	clientInfo := clientinfo.Get()
	if clientInfo == nil {
		clientInfo = &clientinfo.ClientInfo{}
	}

	username, dynamic := c.getUserName(req)
	upstreamIP := strings.Trim(req.RemoteAddr[:strings.LastIndex(req.RemoteAddr, ":")], "[]")
	binary, _ := os.Executable()
	outboundIP := clientinfo.GetOutboundIP(req.Context())
	backupPath := filepath.Join(filepath.Dir(c.Flags.ConfigFile), "backups", filepath.Base(c.Flags.ConfigFile))
	ifName, netmask := getIfNameAndNetmask(outboundIP)
	hostInfo, _ := website.Site.GetHostInfo(req.Context())
	activeTunnel := ""
	poolStats := map[string]*mulery.PoolSize{}
	triggers, timers, schedules := c.triggers.GatherTriggerInfo()

	plexInfo := &plex.PMSInfo{}
	if ps := data.Get("plexStatus"); ps != nil {
		plexInfo, _ = ps.Data.(*plex.PMSInfo)
	}

	if at := data.Get("activeTunnel"); at != nil {
		activeTunnel, _ = at.Data.(string)
	}

	if c.tunnel != nil {
		poolStats = c.tunnel.PoolStats()
	}

	resp.Header().Set("Content-Type", mnd.ContentTypeJSON)

	if err := json.NewEncoder(resp).Encode(&Profile{
		PlexInfo:        plexInfo,
		Triggers:        triggers,
		Timers:          timers,
		Schedules:       schedules,
		Username:        username,
		Config:          *c.Config,
		ClientInfo:      clientInfo,
		IsWindows:       mnd.IsWindows,
		IsLinux:         mnd.IsLinux,
		IsDarwin:        mnd.IsDarwin,
		IsFreeBSD:       mnd.IsFreeBSD,
		IsDocker:        mnd.IsDocker,
		IsUnstable:      mnd.IsUnstable,
		IsSynology:      mnd.IsSynology,
		Headers:         c.getProfileHeaders(req),
		Fortune:         Fortune(),
		UpstreamIP:      upstreamIP,
		UpstreamAllowed: c.allow.Contains(req.RemoteAddr),
		UpstreamHeader:  c.Config.UIPassword.Header(),
		UpstreamType:    c.Config.UIPassword.Type(),
		Updated:         time.Now().UTC(),
		Languages:       frontend.Translations(),
		ProxyAllow:      c.allow.Contains(req.RemoteAddr),
		Flags:           c.Flags,
		Dynamic:         dynamic,
		Webauth:         c.webauth,
		LogFiles:        logs.Log.GetAllLogFilePaths(),
		ConfigFiles:     logs.GetFilePaths(c.Flags.ConfigFile, backupPath),
		CheckResults:    c.Services.GetResults(),
		CheckRunning:    c.Services.Running(),
		CheckDisabled:   c.Services.Disabled,
		SiteCrons:       c.triggers.CronTimer.List(),
		Disks:           c.getDisks(req.Context()),
		Expvar:          mnd.GetAllData(),
		HostInfo:        hostInfo,
		Started:         version.Started.Round(time.Second),
		CmdList:         c.triggers.Commands.List(),
		Program:         c.Flags.Name(),
		Version:         version.Version,
		Revision:        version.Revision,
		Branch:          version.Branch,
		BuildUser:       version.BuildUser,
		BuildDate:       version.BuildDate,
		GoVersion:       version.GoVersion,
		OS:              runtime.GOOS,
		Arch:            runtime.GOARCH,
		Binary:          binary,
		Environment:     environ(),
		Docker:          mnd.IsDocker,
		UID:             os.Getuid(),
		GID:             os.Getgid(),
		IP:              outboundIP,
		Gateway:         getGateway(),
		IfName:          ifName,
		Netmask:         netmask,
		MD5:             private.MD5(),
		ActiveTunnel:    activeTunnel,
		TunnelPoolStats: poolStats,
	}); err != nil {
		logs.Log.Errorf("Writing HTTP Response: %v", err)
	}
}

// handleProfilePost handles profile updates including authentication settings and upstream configuration.
//
//	@Summary		Update user profile
//	@Description	Updates user profile settings including authentication type, password, header, and upstream configuration.
//	@Tags			System
//	@Accept			json
//	@Produce		text/plain
//	@Param			profile	body		ProfilePost	true	"Profile update data"
//	@Success		200		{string}	string		"success message"
//	@Failure		400		{string}	string		"invalid request or password"
//	@Failure		500		{string}	string		"error saving config"
//	@Router			/profile [post]
//
//nolint:lll
func (c *Client) handleProfilePost(response http.ResponseWriter, request *http.Request) {
	post := &ProfilePost{}
	if err := json.NewDecoder(request.Body).Decode(post); err != nil {
		http.Error(response, "Invalid Request", http.StatusBadRequest)
		return
	}

	currUser, dynamic := c.getUserName(request)
	if !dynamic {
		// If the auth is currently using a password, check the password.
		if !c.checkUserPass(currUser, post.Password) {
			logs.Log.Errorf("[gui '%s' requested] Trust Profile: Invalid existing (current) password provided.", currUser)
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
		logs.Log.Errorf("[gui '%s' requested] Saving Config: %v", currUser, err)
		http.Error(response, "Saving Config: "+err.Error(), http.StatusInternalServerError)
	case post.AuthType == configfile.AuthNone:
		logs.Log.Printf("[gui '%s' requested] Disabled WebUI authentication.", currUser)
		http.Error(response, "Disabled WebUI authentication.", http.StatusOK)
		c.reloadAppNow()
	default:
		logs.Log.Printf("[gui '%s' requested] Enabled WebUI proxy authentication, header: %s", currUser, post.Header)
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

	if len(newPassw) < minPasswordLen {
		logs.Log.Errorf("[gui '%s' requested] New password must be at least %d characters.", currUser, minPasswordLen)
		http.Error(response, fmt.Sprintf("New password must be at least %d characters.",
			minPasswordLen), http.StatusBadRequest)
		return
	}

	if err := c.setUserPass(request.Context(), configfile.AuthPassword, newUser, newPassw); err != nil {
		logs.Log.Errorf("[gui '%s' requested] Saving Trust Profile: %v", currUser, err)
		http.Error(response, "Saving Trust Profile: "+err.Error(), http.StatusInternalServerError)

		return
	}

	logs.Log.Printf("[gui '%s' requested] Updated Trust Profile settings, username: %s", currUser, newUser)
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
