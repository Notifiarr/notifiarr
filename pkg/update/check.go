// Package update checks for an available update on GitHub.
// It has baked in assumptions, but is mostly portable.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

// Latest is where we find the latest release.
const Latest = "https://api.github.com/repos/%s/releases/latest"

// Update contains running Version, Current version and Download URL for Current version.
// Outdate is true if the running version is older than the current version.
type Update struct {
	Outdate bool
	Date    time.Time
	Version string
	Current string
	URL     string
}

// Check checks if the app this library lives in has an updated version on GitHub.
func Check(userRepo string, version string) (*Update, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5) // nolint:gomnd
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(Latest, userRepo), nil)
	if err != nil {
		return nil, fmt.Errorf("requesting github: %w", err)
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("querying github: %w", err)
	}

	release, err := decodeBody(resp)
	if err != nil {
		return nil, fmt.Errorf("decoding github response: %w", err)
	}

	return fillUpdate(release, version), nil
}

func decodeBody(resp *http.Response) (*gitHubReleasesLatest, error) {
	defer resp.Body.Close()

	var release gitHubReleasesLatest

	return &release, json.NewDecoder(resp.Body).Decode(&release)
}

func fillUpdate(release *gitHubReleasesLatest, version string) *Update {
	u := &Update{
		Current: release.TagName,
		Version: "v" + version,
		Outdate: semver.Compare("v"+strings.TrimPrefix(release.TagName, "v"),
			"v"+strings.TrimPrefix(version, "v")) > 0,
		URL:  release.HTMLURL,
		Date: release.PublishedAt,
	}

	arch := runtime.GOARCH
	if arch == "arm" {
		arch = "armhf"
	} else if arch == "386" {
		arch = "i386"
	}

	switch runtime.GOOS {
	case "darwin":
		return u.getFile(release.Assets, ".exe.zip")
	case "windows":
		return u.getFile(release.Assets, ".dmg")
	case "freebsd":
		return u.getFile(release.Assets, arch+".txz")
	case "linux":
		fallthrough // :( too many variants
	default:
		return u
	}
}

func (u *Update) getFile(assets []*asset, suffix string) *Update {
	for _, file := range assets {
		if strings.HasSuffix(file.BrowserDownloadURL, suffix) {
			u.URL = file.BrowserDownloadURL
			u.Date = file.UpdatedAt
		}
	}

	return u
}

type gitHubReleasesLatest struct {
	URL             string    `json:"url"`
	AssetsURL       string    `json:"assets_url"`
	UploadURL       string    `json:"upload_url"`
	HTMLURL         string    `json:"html_url"`
	ID              int64     `json:"id"`
	Author          *user     `json:"author"`
	NodeID          string    `json:"node_id"`
	TagName         string    `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Draft           bool      `json:"draft"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     time.Time `json:"published_at"`
	Assets          []*asset  `json:"assets"`
	TarballURL      string    `json:"tarball_url"`
	ZipballURL      string    `json:"zipball_url"`
	Body            string    `json:"body"`
}

type asset struct {
	URL                string    `json:"url"`
	ID                 int64     `json:"id"`
	NodeID             string    `json:"node_id"`
	Name               string    `json:"name"`
	Label              string    `json:"label"`
	Uploader           *user     `json:"uploader"`
	ContentType        string    `json:"content_type"`
	State              string    `json:"state"`
	Size               int       `json:"size"`
	DownloadCount      int       `json:"download_count"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	BrowserDownloadURL string    `json:"browser_download_url"`
}

type user struct {
	Login             string `json:"login"`
	ID                int64  `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}
