package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/frontend"
	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
)

// Profile is the data returned by the profile GET endpoint.
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
	LoggedIn bool      `json:"loggedIn"`
	Updated  time.Time `json:"updated"`
}

// handleProfile returns the current user's username in a JSON response.
func (c *Client) handleProfile(w http.ResponseWriter, req *http.Request) {
	username, _ := c.getUserName(req)
	clientInfo := clientinfo.Get()
	upstreamIP := strings.Trim(req.RemoteAddr[:strings.LastIndex(req.RemoteAddr, ":")], "[]")

	if err := json.NewEncoder(w).Encode(&Profile{
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
		UpstreamAllowed: c.Config.Allow.Contains(req.RemoteAddr),
		UpstreamHeader:  c.Config.UIPassword.Header(),
		UpstreamType:    c.Config.UIPassword.Type(),
		Updated:         time.Now().UTC(),
		Languages:       frontend.Translations(),
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
			c.Errorf("[gui '%s' requested] Invalid existing (current) password provided. '%s'", currUser, post.Password)
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

	switch err := c.setUserPass(request.Context(), post.AuthType.Type(), post.Header, ""); {
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

	if err := c.setUserPass(request.Context(), "password", newUser, newPassw); err != nil {
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
