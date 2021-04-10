// Package update checks for an available update on GitHub.
// It has baked in assumptions, but is mostly portable.
package update

import (
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kardianos/osext"
	"golang.org/x/mod/semver"
)

// OSsuffixMap is the OS to file suffix map for downloads.
var OSsuffixMap = map[string]string{ //nolint:gochecknoglobals
	"darwin":  ".dmg",
	"windows": ".exe.zip",
	"freebsd": ".txz",
	"linux":   "", // too many variants right now.
}

// Latest is where we find the latest release.
const Latest = "https://api.github.com/repos/%s/releases/latest"

// GitHub API and JSON unmarshal timeout.
const (
	timeout         = 10 * time.Second
	downloadTimeout = 5 * time.Minute
)

// Update contains running Version, Current version and Download URL for Current version.
// Outdate is true if the running version is older than the current version.
type Update struct {
	Outdate bool
	RelDate time.Time
	Version string
	Current string
	CurrURL string
}

// Check checks if the app this library lives in has an updated version on GitHub.
func Check(userRepo string, version string) (*Update, error) {
	release, err := GetRelease(fmt.Sprintf(Latest, userRepo))
	if err != nil {
		return nil, err
	}

	return FillUpdate(release, version), nil
}

// GetRelease returns a GitHub release. See Check for an example on how to use it.
func GetRelease(uri string) (*GitHubReleasesLatest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("requesting github: %w", err)
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("querying github: %w", err)
	}
	defer resp.Body.Close()

	var release GitHubReleasesLatest
	if err = json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decoding github response: %w", err)
	}

	return &release, nil
}

// FillUpdate compares a current version with the latest GitHub release.
func FillUpdate(release *GitHubReleasesLatest, version string) *Update {
	u := &Update{
		RelDate: release.PublishedAt,
		CurrURL: release.HTMLURL,
		Current: release.TagName,
		Version: "v" + strings.TrimPrefix(version, "v"),
		Outdate: semver.Compare("v"+strings.TrimPrefix(release.TagName, "v"),
			"v"+strings.TrimPrefix(version, "v")) > 0,
	}

	arch := runtime.GOARCH
	if arch == "arm" {
		arch = "armhf"
	} else if arch == "386" {
		arch = "i386"
	}

	suffix := OSsuffixMap[runtime.GOOS]
	if runtime.GOOS == "freebsd" || runtime.GOOS == "linux" {
		suffix = arch + suffix
	}

	for _, file := range release.Assets {
		if strings.HasSuffix(file.BrowserDownloadURL, suffix) {
			u.CurrURL = file.BrowserDownloadURL
			u.RelDate = file.UpdatedAt
		}
	}

	return u
}

// Command is the input data to perform an in-place update.
type Command struct {
	URL         string
	Path        string
	Args        []string
	*log.Logger // debug logs.
}

// This downloads the new file to a temp name in the same folder as the running file.
// Moves the running file to a backup name in the same folder.
// Moves the new file to the same location that the running file was at.
// Triggers another invocation of the app that sleeps 5 seconds then restarts.
// The running app must exit after this returns!
func Now(u *Command) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), downloadTimeout)
	defer cancel()

	backupFile, err := u.replaceFile(ctx)
	if err != nil {
		return backupFile, err
	}

	u.Printf("[Update] Triggering Restart: %s %s", u.Path, strings.Join(u.Args, " "))

	if err := exec.Command(u.Path, u.Args...).Start(); err != nil { //nolint:gosec
		return backupFile, fmt.Errorf("executing restart command %w", err)
	}

	return backupFile, nil
}

func (u *Command) replaceFile(ctx context.Context) (string, error) {
	tempFolderPath, err := osext.ExecutableFolder()
	if err != nil {
		return "", fmt.Errorf("getting appliction folder: %w", err)
	}

	tempFile, err := u.writeFile(ctx, tempFolderPath)
	if err != nil {
		return "", err
	}

	if runtime.GOOS != "windows" {
		_ = os.Chmod(tempFile, 0755)
	}

	backupFile := u.Path + ".upgrade.backup." + time.Now().Format("06-01-02T15:04:05")
	u.Printf("[Update] Renaming %s => %s", u.Path, backupFile)

	if err := os.Rename(u.Path, backupFile); err != nil {
		return backupFile, fmt.Errorf("renaming original file %w", err)
	}

	u.Printf("[Update] Renaming %s => %s", tempFile, u.Path)

	if err := os.Rename(tempFile, u.Path); err != nil {
		return backupFile, fmt.Errorf("renaming downloaded file %w", err)
	}

	return backupFile, nil
}

func (u *Command) writeFile(ctx context.Context, folderPath string) (string, error) {
	tempFile, err := os.CreateTemp(folderPath, filepath.Base(u.Path))
	if err != nil {
		return "", fmt.Errorf("creating temporary file: %w", err)
	}
	defer tempFile.Close()

	u.Printf("[Update] Primed Temp File: %s", tempFile.Name())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.URL, nil)
	if err != nil {
		return tempFile.Name(), fmt.Errorf("creating request: %w", err)
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return tempFile.Name(), fmt.Errorf("downloading release: %w", err)
	}
	defer resp.Body.Close()

	return tempFile.Name(), u.decompressFile(tempFile, resp.Body)
}

func (u *Command) decompressFile(tempFile *os.File, resp io.Reader) error {
	switch {
	case strings.HasSuffix(u.URL, ".zip"):
		body, err := ioutil.ReadAll(resp)
		if err != nil {
			return fmt.Errorf("reading file from URL: %w", err)
		}

		if err := u.writeZipFile(tempFile, body); err != nil {
			return err
		}
	case strings.HasSuffix(u.URL, ".gz"):
		if err := u.writeGZipFile(tempFile, resp); err != nil {
			return err
		}
	case strings.HasSuffix(u.URL, ".bz2"):
		if _, err := io.Copy(tempFile, bzip2.NewReader(resp)); err != nil {
			return fmt.Errorf("bzunzipping temporary file: %w", err)
		}
	default:
		if _, err := io.Copy(tempFile, resp); err != nil {
			return fmt.Errorf("writing temporary file: %w", err)
		}
	}

	return nil
}

func (u *Command) writeGZipFile(tempFile *os.File, resp io.Reader) error {
	gz, err := gzip.NewReader(resp)
	if err != nil {
		return fmt.Errorf("reading gzip file: %w", err)
	}
	defer gz.Close()

	if _, err := io.Copy(tempFile, gz); err != nil { //nolint:gosec
		return fmt.Errorf("gunzipping temporary file: %w", err)
	}

	return nil
}

func (u *Command) writeZipFile(tempFile *os.File, body []byte) error {
	var (
		ioReader io.Reader
		buff     = bytes.NewBuffer(body)
	)

	size, err := io.Copy(buff, ioReader)
	if err != nil {
		return fmt.Errorf("buffering zip file: %w", err)
	}

	// Open a zip archive for reading.
	zipReader, err := zip.NewReader(bytes.NewReader(buff.Bytes()), size)
	if err != nil {
		return fmt.Errorf("reading zip file: %w", err)
	}

	// Find the exe file and write that.
	for _, zipFile := range zipReader.File {
		if strings.HasSuffix(zipFile.Name, ".exe") {
			f, err := zipFile.Open()
			if err != nil {
				return fmt.Errorf("reading zipped file: %w", err)
			}
			defer f.Close()

			if _, err := io.Copy(tempFile, f); err != nil { //nolint:gosec
				return fmt.Errorf("unzipping temporary file: %w", err)
			}

			return nil
		}
	}

	return nil
}

// Restart is meant to be called from a special flag that reloads the app after an upgrade.
func Restart(u *Command) error {
	fmt.Println("Sleeping 5 seconds before restarting.")
	time.Sleep(5 * time.Second) //nolint:gomnd

	if err := exec.Command(u.Path, u.Args...).Start(); err != nil { //nolint:gosec
		return fmt.Errorf("executing command %w", err)
	}

	return nil
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
	Size               int       `json:"size"`
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
