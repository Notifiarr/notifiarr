package apps

import (
	"context"
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
func (t *TautulliConfig) GetUsers(ctx context.Context) (*TautulliUsers, error) {
	if t == nil || t.URL == "" {
		return &TautulliUsers{}, nil
	}

	params := url.Values{}
	params.Add("cmd", "get_users")
	params.Add("apikey", t.APIKey)

	var users TautulliUsers

	err := GetURLInto(ctx, "Tautulli", t.Timeout.Duration, t.VerifySSL, t.URL+"/api/v2", params, &users)
	if err != nil {
		return nil, err
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
//nolint:tagliatelle
type TautulliUser struct {
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

// MapEmailName returns a map of email => name for Tautulli users.
func (t *TautulliUsers) MapEmailName() map[string]string {
	if t == nil {
		return nil
	}

	nameMap := map[string]string{}

	for _, user := range t.Response.Data {
		// user.FriendlyName always seems to be set, so this first if-block is safety only.
		if user.FriendlyName == "" && user.Email != "" && user.Username != "" {
			nameMap[user.Email] = user.Username
			continue
		} else if user.FriendlyName == "" {
			// This user has no mapability.
			continue
		}

		// We only need username or email, not both, but in the order username then email.
		if user.Username != "" {
			nameMap[user.Username] = user.FriendlyName
		} else if user.Email != "" {
			nameMap[user.Email] = user.FriendlyName
		}
	}

	return nameMap
}
