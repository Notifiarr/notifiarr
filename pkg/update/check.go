//nolint:tagliatelle
package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"golang.org/x/mod/semver"
)

// OSsuffixMap is the OS to file suffix map for downloads.
var OSsuffixMap = map[string]string{ //nolint:gochecknoglobals
	"darwin":    "Notifiarr.dmg",
	mnd.Windows: ".exe.zip",
	"freebsd":   ".txz",
	"linux":     "", // too many variants right now.
}

// Custom errors.
var (
	ErrNoFile       = errors.New("no downloadable file found in release")
	ErrBadStatus    = errors.New("unexpected http status")
	ErrBodyTooLarge = errors.New("response body too large")
)

// LatestGH is where we find the latest release.
const LatestGH = "https://api.github.com/repos/%s/releases/latest"

const (
	// timeout is the per-attempt budget for a GitHub/unstable version check.
	timeout = 10 * time.Second
	// versionCheckAttempts is 1 try plus retries for Cloudflare 5xx blips.
	versionCheckAttempts = 3
)

// versionCheckRetry is the pause between version-check attempts.
var versionCheckRetry = time.Second

// httpStatusError is an HTTP status that was not 200.
type httpStatusError struct {
	URI    string
	Status int
}

func (e *httpStatusError) Error() string {
	return fmt.Sprintf("%s: %s: %d", ErrBadStatus, e.URI, e.Status)
}

func (e *httpStatusError) Unwrap() error {
	return ErrBadStatus
}

// Update contains running Version, Current version and Download URL for Current version.
// Outdate is true if the running version is older than the current version.
type Update struct {
	Outdate bool      // True if we're outdated, update available.
	Version string    // Version passed in externally.
	Current string    // Current release available on GH or US.
	CurrURL string    // URL of current release on GH or US.
	RelDate time.Time // Current version release date.
	RelSize int64     // Current release file size.
}

// CheckGitHub checks if the app this library lives in has an updated version on GitHub.
func CheckGitHub(ctx context.Context, userRepo string, version string) (*Update, error) {
	release, err := GetRelease(ctx, fmt.Sprintf(LatestGH, userRepo))
	if err != nil {
		return nil, err
	}

	return FillUpdate(release, version)
}

// GetRelease returns a GitHub release. See Check for an example on how to use it.
func GetRelease(ctx context.Context, uri string) (*GitHubReleasesLatest, error) {
	var release GitHubReleasesLatest
	if err := doVersionCheck(ctx, uri, &release, githubJSONLimit, nil); err != nil {
		return nil, err
	}

	return &release, nil
}

const (
	unstableJSONLimit = 1024
	githubJSONLimit   = 1024 * 1024
	maxLogBody        = 4 * 1024
)

func decodeJSONBody(ctx context.Context, resp *http.Response, uri string, dest any, maxBody int) error {
	readLimit := max(maxBody, maxLogBody)

	body, err := io.ReadAll(io.LimitReader(resp.Body, int64(readLimit)+1))
	if err != nil {
		return fmt.Errorf("reading %s response: %w", uri, err)
	}

	if len(body) > maxBody {
		logUpdateBody(ctx, resp, uri, body, ErrBodyTooLarge)
		return fmt.Errorf("%w: %s (%d bytes)", ErrBodyTooLarge, uri, len(body))
	}

	if resp.StatusCode != http.StatusOK {
		logUpdateBody(ctx, resp, uri, body, nil)
		return &httpStatusError{URI: uri, Status: resp.StatusCode}
	}

	if err = json.Unmarshal(body, dest); err != nil {
		logUpdateBody(ctx, resp, uri, body, err)
		return fmt.Errorf("decoding %s response: %w", uri, err)
	}

	return nil
}

func logUpdateBody(ctx context.Context, resp *http.Response, uri string, body []byte, decodeErr error) {
	if mnd.Log == nil {
		return
	}

	logged := body
	if len(logged) > maxLogBody {
		logged = logged[:maxLogBody]
	}

	if decodeErr != nil {
		mnd.Log.Errorf(mnd.GetID(ctx), "[UPDATE] decoding %s: status %d type %q body %q: %v",
			uri, resp.StatusCode, resp.Header.Get("Content-Type"), logged, decodeErr)
		return
	}

	mnd.Log.Errorf(mnd.GetID(ctx), "[UPDATE] %s: status %d type %q body %q",
		uri, resp.StatusCode, resp.Header.Get("Content-Type"), logged)
}

func doVersionCheck(ctx context.Context, uri string, dest any, maxBody int, after func(*http.Response)) error {
	var lastErr error

	for attempt := 1; attempt <= versionCheckAttempts; attempt++ {
		err := doVersionCheckOnce(ctx, uri, dest, maxBody, after)
		if err == nil {
			return nil
		}

		lastErr = err
		if attempt == versionCheckAttempts || !retryableVersionCheck(err) {
			return err
		}

		if mnd.Log != nil {
			mnd.Log.Errorf(mnd.GetID(ctx), "[UPDATE] [%d/%d] version check failed, retrying in %s: %v",
				attempt, versionCheckAttempts, versionCheckRetry, err)
		}

		timer := time.NewTimer(versionCheckRetry)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("version check: %w", ctx.Err())
		case <-timer.C:
		}
	}

	return lastErr
}

func doVersionCheckOnce(ctx context.Context, uri string, dest any, maxBody int, after func(*http.Response)) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return fmt.Errorf("requesting %s: %w", uri, err)
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return fmt.Errorf("querying %s: %w", uri, err)
	}
	defer resp.Body.Close()

	if err = decodeJSONBody(ctx, resp, uri, dest, maxBody); err != nil {
		return err
	}

	if after != nil {
		after(resp)
	}

	return nil
}

func retryableVersionCheck(err error) bool {
	switch {
	case err == nil, errors.Is(err, context.Canceled), errors.Is(err, ErrBodyTooLarge):
		return false
	}

	if statusErr, ok := errors.AsType[*httpStatusError](err); ok {
		return statusErr.Status >= http.StatusInternalServerError ||
			statusErr.Status == http.StatusTooManyRequests ||
			statusErr.Status == http.StatusRequestTimeout
	}

	// Transport errors, per-attempt timeouts, and Cloudflare 200 error pages.
	return true
}

// FillUpdate compares a current version with the latest GitHub release.
func FillUpdate(release *GitHubReleasesLatest, version string) (*Update, error) {
	update := &Update{
		RelDate: release.PublishedAt,
		CurrURL: release.HTMLURL,
		Current: release.TagName,
		Version: "v" + strings.TrimPrefix(version, "v"),
		Outdate: semver.Compare("v"+strings.TrimPrefix(release.TagName, "v"),
			"v"+strings.TrimPrefix(version, "v")) > 0,
	}

	arch := runtime.GOARCH
	switch arch {
	case "arm":
		arch = "armhf"
	case "386":
		arch = "i386"
	}

	suffix := OSsuffixMap[runtime.GOOS]
	if mnd.IsFreeBSD || mnd.IsLinux {
		suffix = arch + suffix
	}

	for _, file := range release.Assets {
		if strings.HasSuffix(file.BrowserDownloadURL, suffix) {
			update.CurrURL = file.BrowserDownloadURL
			update.RelDate = file.UpdatedAt
			update.RelSize = file.Size

			break
		}
	}

	if release.HTMLURL == update.CurrURL {
		return update, fmt.Errorf("%w: %s", ErrNoFile, update.CurrURL)
	}

	return update, nil
}

// GitHubReleasesLatest is the output from the releases/latest API on GitHub.
type GitHubReleasesLatest struct {
	URL             string    `json:"url"`
	AssetsURL       string    `json:"assets_url"`
	UploadURL       string    `json:"upload_url"`
	HTMLURL         string    `json:"html_url"`
	ID              int64     `json:"id"`
	Author          GHuser    `json:"author"`
	NodeID          string    `json:"node_id"`
	TagName         string    `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Draft           bool      `json:"draft"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     time.Time `json:"published_at"`
	Assets          []GHasset `json:"assets"`
	TarballURL      string    `json:"tarball_url"`
	ZipballURL      string    `json:"zipball_url"`
	Body            string    `json:"body"`
}

// GHasset is part of GitHubReleasesLatest.
type GHasset struct {
	URL                string    `json:"url"`
	ID                 int64     `json:"id"`
	NodeID             string    `json:"node_id"`
	Name               string    `json:"name"`
	Label              string    `json:"label"`
	Uploader           GHuser    `json:"uploader"`
	ContentType        string    `json:"content_type"`
	State              string    `json:"state"`
	Size               int64     `json:"size"`
	DownloadCount      int       `json:"download_count"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	BrowserDownloadURL string    `json:"browser_download_url"`
}

// GHuser is part of GitHubReleasesLatest.
type GHuser struct {
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
