package tautulli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
)

// Config is the Tautulli configuration.
type Config struct {
	URL          string `json:"url"    toml:"url"     xml:"url"`
	APIKey       string `json:"apiKey" toml:"api_key" xml:"api_key"`
	*http.Client `json:"-"      toml:"-"       xml:"-"`
}

// GetURLInto gets a url and unmarshals the contents into the provided interface pointer.
func (c *Config) GetURLInto(ctx context.Context, params url.Values, into interface{}) error {
	err := func() error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL+"/api/v2", nil)
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}

		req.URL.RawQuery = params.Encode()

		resp, err := c.Client.Do(req)
		if err != nil {
			return fmt.Errorf("making request: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response (%s): %w: %s", resp.Status, err, string(body))
		}

		if err := json.Unmarshal(body, into); err != nil {
			return fmt.Errorf("decoding response (%s): %w: %s", resp.Status, err, string(body))
		}

		return nil
	}()
	if err != nil {
		return fmt.Errorf("%s", strings.ReplaceAll(err.Error(), c.APIKey, "<redacted>")) //nolint:goerr113
	}

	return nil
}

// GetUsers returns the Tautulli users.
func (c *Config) GetUsers(ctx context.Context) (*Users, error) {
	if c == nil || c.Client == nil {
		return &Users{}, nil
	}

	params := url.Values{}
	params.Add("cmd", "get_users")
	params.Add("apikey", c.APIKey)

	var users Users

	err := c.GetURLInto(ctx, params, &users)
	if err != nil {
		return nil, err
	}

	return &users, nil
}

// Info represent the data returned by the get_tautulli_info command.
//
//nolint:tagliatelle
type Info struct {
	InstallType         string `json:"tautulli_install_type"`
	Version             string `json:"tautulli_version"`
	Branch              string `json:"tautulli_branch"`
	Commit              string `json:"tautulli_commit"`
	Platform            string `json:"tautulli_platform"`
	PlatformRelease     string `json:"tautulli_platform_release"`
	PlatformVersion     string `json:"tautulli_platform_version"`
	PlatformLinuxDistro string `json:"tautulli_platform_linux_distro"`
	PlatformDeviceName  string `json:"tautulli_platform_device_name"`
	PythonVersion       string `json:"tautulli_python_version"`
}

// GetInfo returns the Tautulli app info.
func (c *Config) GetInfo(ctx context.Context) (*Info, error) {
	if c == nil || c.Client == nil {
		return nil, nil //nolint:nilnil
	}

	params := url.Values{}
	params.Add("cmd", "get_tautulli_info")
	params.Add("apikey", c.APIKey)

	var output struct {
		Resp struct {
			Result string `json:"result"`
			Data   *Info  `json:"data"`
		} `json:"response"`
	}

	if err := c.GetURLInto(ctx, params, &output); err != nil {
		return nil, err
	}

	return output.Resp.Data, nil
}

// Users is the entire get_users API response.
type Users struct {
	Response struct {
		Result  string `json:"result"`  // success, error
		Message string `json:"message"` // error msg
		Data    []User `json:"data"`
	} `json:"response"`
}

// User is the user data from the get_users API call.
//
//nolint:tagliatelle
type User struct {
	RowID           int64    `json:"row_id"`
	UserID          int64    `json:"user_id"`
	Username        string   `json:"username"`
	FriendlyName    string   `json:"friendly_name"`
	Thumb           string   `json:"thumb"`
	Email           string   `json:"email"`
	ServerToken     string   `json:"server_token"`
	SharedLibraries []string `json:"shared_libraries"`
	FilterAll       string   `json:"filter_all"`
	FilterMovies    string   `json:"filter_movies"`
	FilterTv        string   `json:"filter_tv"`
	FilterMusic     string   `json:"filter_music"`
	FilterPhotos    string   `json:"filter_photos"`
	IsActive        int      `json:"is_active"`     // 1,0 (bool)
	IsAdmin         int      `json:"is_admin"`      // 1,0 (bool)
	IsHomeUser      int      `json:"is_home_user"`  // 1,0 (bool)
	IsAllowSync     int      `json:"is_allow_sync"` // 1,0 (bool)
	IsRestricted    int      `json:"is_restricted"` // 1,0 (bool)
	DoNotify        int      `json:"do_notify"`     // 1,0 (bool)
	KeepHistory     int      `json:"keep_history"`  // 1,0 (bool)
	AllowGuest      int      `json:"allow_guest"`   // 1,0 (bool)
}

// MapIDName returns a map of plex user ID => Feiendly Name (or username) for Tautulli users.
func (u *Users) MapIDName() map[string]string {
	if u == nil {
		return nil
	}

	nameMap := map[string]string{}

	for _, user := range u.Response.Data {
		if user.FriendlyName == "" {
			nameMap[strconv.FormatInt(user.UserID, mnd.Base10)] = user.Username
			nameMap[user.Username] = user.Username
		} else {
			nameMap[strconv.FormatInt(user.UserID, mnd.Base10)] = user.FriendlyName
			nameMap[user.Username] = user.FriendlyName
		}
	}

	return nameMap
}
