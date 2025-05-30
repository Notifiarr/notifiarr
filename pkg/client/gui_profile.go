package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/frontend"
	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/private"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
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
	// LoggedIn is only used by the front end. Backend does not set or use it.
	LoggedIn        bool                          `json:"loggedIn"`
	Updated         time.Time                     `json:"updated"`
	Flags           *configfile.Flags             `json:"flags"`
	Dynamic         bool                          `json:"dynamic"`
	Webauth         bool                          `json:"webauth"`
	Msg             string                        `json:"msg,omitempty"`
	LogFiles        *logs.LogFileInfos            `json:"logFileInfo"`
	ConfigFiles     *logs.LogFileInfos            `json:"configFileInfo"`
	Expvar          mnd.AllData                   `json:"expvar"`
	HostInfo        *host.InfoStat                `json:"hostInfo"`
	Disks           map[string]snapshot.Partition `json:"disks"`
	ProxyAllow      bool                          `json:"proxyAllow"`
	PoolStats       map[string]*mulery.PoolSize   `json:"poolStats"`
	Started         time.Time                     `json:"started"`
	Program         string                        `json:"program"`
	Version         string                        `json:"version"`
	Revision        string                        `json:"revision"`
	Branch          string                        `json:"branch"`
	BuildUser       string                        `json:"buildUser"`
	BuildDate       string                        `json:"buildDate"`
	GoVersion       string                        `json:"goVersion"`
	OS              string                        `json:"os"`
	Arch            string                        `json:"arch"`
	Binary          string                        `json:"binary"`
	Environment     map[string]string             `json:"environment"`
	Docker          bool                          `json:"docker"`
	UID             int                           `json:"uid"`
	GID             int                           `json:"gid"`
	IP              string                        `json:"ip"`
	Gateway         string                        `json:"gateway"`
	IfName          string                        `json:"ifName"`
	Netmask         string                        `json:"netmask"`
	MD5             string                        `json:"md5"`
	ActiveTunnel    string                        `json:"activeTunnel"`
	TunnelPoolStats map[string]*mulery.PoolSize   `json:"tunnelPoolStats"`
}

// handleProfile returns the current user's username in a JSON response.
//
//nolint:funlen
func (c *Client) handleProfile(resp http.ResponseWriter, req *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			c.Errorf("handleProfile panic: %v\n%#v\n", r, c.Config)
			panic(r)
		}
	}()

	clientInfo := clientinfo.Get()
	if clientInfo == nil {
		clientInfo = &clientinfo.ClientInfo{}
	}

	username, dynamic := c.getUserName(req)
	upstreamIP := strings.Trim(req.RemoteAddr[:strings.LastIndex(req.RemoteAddr, ":")], "[]")
	binary, _ := os.Executable()
	outboundIP := clientinfo.GetOutboundIP()
	backupPath := filepath.Join(filepath.Dir(c.Flags.ConfigFile), "backups", filepath.Base(c.Flags.ConfigFile))
	ifName, netmask := getIfNameAndNetmask(outboundIP)
	hostInfo, _ := website.Site.GetHostInfo(req.Context())
	activeTunnel := ""
	poolStats := map[string]*mulery.PoolSize{}

	if at := data.Get("activeTunnel"); at != nil {
		activeTunnel, _ = at.Data.(string)
	}

	if c.tunnel != nil {
		poolStats = c.tunnel.PoolStats()
	}

	if err := json.NewEncoder(resp).Encode(&Profile{
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
		LogFiles:        c.Logger.GetAllLogFilePaths(),
		ConfigFiles:     logs.GetFilePaths(c.Flags.ConfigFile, backupPath),
		//Disks:           c.getDisks(req.Context()), // TODO: split disks from snapshot.
		Expvar:          mnd.GetAllData(),
		HostInfo:        hostInfo,
		Started:         version.Started.Round(time.Second),
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
		c.Errorf("Writing HTTP Response: %v", err)
	}
}

func (c *Client) handleProfilePost(response http.ResponseWriter, request *http.Request) {
	post, err := c.getProfilePostData(request)
	if err != nil {
		c.Errorf("%v", err)
		http.Error(response, "Invalid Request", http.StatusBadRequest)
		return
	}

	currUser, dynamic := c.getUserName(request)
	if !dynamic {
		// If the auth is currently using a password, check the password.
		if !c.checkUserPass(currUser, post.Password) {
			c.Errorf("[gui '%s' requested] Trust Profile: Invalid existing (current) password provided.", currUser)
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
		c.Errorf("[gui '%s' requested] Saving Config: %v", currUser, err)
		http.Error(response, "Saving Config: "+err.Error(), http.StatusInternalServerError)
	case post.AuthType == configfile.AuthNone:
		c.Printf("[gui '%s' requested] Disabled WebUI authentication.", currUser)
		http.Error(response, "Disabled WebUI authentication.", http.StatusOK)
		c.reloadAppNow()
	default:
		c.Printf("[gui '%s' requested] Enabled WebUI proxy authentication, header: %s", currUser, post.Header)
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

func (c *Client) getProfilePostData(request *http.Request) (*ProfilePost, error) {
	post := &ProfilePost{}

	// The New UI uses JSON, the old UI uses form data.
	if request.Header.Get("Content-Type") != "application/json" {
		// If the request is not JSON, we're using the old form data.
		at, _ := strconv.Atoi(request.PostFormValue("AuthType"))
		post.AuthType = configfile.AuthType(at)
		post.Password = request.PostFormValue("Password")
		post.Header = request.PostFormValue("AuthHeader")
		post.Username = request.PostFormValue("NewUsername")
		post.NewPass = request.PostFormValue("NewPassword")
		post.Upstreams = request.PostFormValue("Upstreams")
		switch request.PostFormValue("AuthType") {
		case "password":
			post.AuthType = configfile.AuthPassword
		case "nopass":
			post.AuthType = configfile.AuthNone
		case "header":
			post.AuthType = configfile.AuthHeader
		}
		return post, nil
	}

	if err := json.NewDecoder(request.Body).Decode(&post); err != nil {
		return nil, fmt.Errorf("decoding request json: %w", err)
	}

	return post, nil
}

func (c *Client) handleProfilePostPassword(
	response http.ResponseWriter,
	request *http.Request,
	newUser, newPassw string,
) {
	currUser, _ := c.getUserName(request)

	if len(newPassw) < minPasswordLen {
		c.Errorf("[gui '%s' requested] New password must be at least %d characters.", currUser, minPasswordLen)
		http.Error(response, fmt.Sprintf("New password must be at least %d characters.",
			minPasswordLen), http.StatusBadRequest)
		return
	}

	if err := c.setUserPass(request.Context(), configfile.AuthPassword, newUser, newPassw); err != nil {
		c.Errorf("[gui '%s' requested] Saving Trust Profile: %v", currUser, err)
		http.Error(response, "Saving Trust Profile: "+err.Error(), http.StatusInternalServerError)

		return
	}

	c.Printf("[gui '%s' requested] Updated Trust Profile settings, username: %s", currUser, newUser)
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
