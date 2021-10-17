package apps

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (t *TautulliConfig) setup(timeout time.Duration) {
	if t == nil {
		return
	}

	if t.Timeout.Duration == 0 {
		t.Timeout.Duration = timeout
	}

	t.URL = strings.TrimRight(t.URL, "/")
}

// GetUsers returns the Tautulli users.
func (t *TautulliConfig) GetUsers() (*TautulliUsers, error) {
	if t == nil || t.URL == "" {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), t.Timeout.Duration)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.URL+"/api/v2", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	params := url.Values{}
	params.Add("cmd", "get_users")
	params.Add("apikey", t.APIKey)
	req.URL.RawQuery = params.Encode()

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	var users TautulliUsers
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("parsing json: %w", err)
	}

	return &users, nil
}

// TautulliUsers is the entire get_users API response.
type TautulliUsers struct {
	Response struct {
		Result  string         `json:"result"`  // success, error
		Message string         `json:"message"` // error msg
		Data    []TautulliUser `json:"data"`
	} `json:"response"`
}

// TautulliUser is the user data from the get_users API call.
type TautulliUser struct {
	RowID           int64  `json:"row_id"`
	UserID          int64  `json:"user_id"`
	Username        string `json:"username"`
	FriendlyName    string `json:"friendly_name"`
	Thumb           string `json:"thumb"`
	Email           string `json:"email"`
	ServerToken     string `json:"server_token"`
	SharedLibraries string `json:"shared_libraries"`
	FilterAll       string `json:"filter_all"`
	FilterMovies    string `json:"filter_movies"`
	FilterTv        string `json:"filter_tv"`
	FilterMusic     string `json:"filter_music"`
	FilterPhotos    string `json:"filter_photos"`
	IsActive        int    `json:"is_active"`     // 1,0 (bool)
	IsAdmin         int    `json:"is_admin"`      // 1,0 (bool)
	IsHomeUser      int    `json:"is_home_user"`  // 1,0 (bool)
	IsAllowSync     int    `json:"is_allow_sync"` // 1,0 (bool)
	IsRestricted    int    `json:"is_restricted"` // 1,0 (bool)
	DoNotify        int    `json:"do_notify"`     // 1,0 (bool)
	KeepHistory     int    `json:"keep_history"`  // 1,0 (bool)
	AllowGuest      int    `json:"allow_guest"`   // 1,0 (bool)
}

// MapEmailName returns a map of email => name for Tautulli users.
func (t *TautulliUsers) MapEmailName() map[string]string {
	if t == nil {
		return nil
	}

	m := map[string]string{}

	for _, user := range t.Response.Data {
		if user.Email != "" && user.FriendlyName != "" {
			if user.Email == user.FriendlyName && user.Username != "" {
				m[user.Email] = user.Username
			} else {
				m[user.Email] = user.FriendlyName
			}
		}
	}

	return m
}
