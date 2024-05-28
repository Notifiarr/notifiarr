package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type UnstableFile struct {
	Time time.Time `json:"time"`
	File string    `json:"file"`
	Ver  string    `json:"version"`
	Rev  int       `json:"revision"`
	Size int64     `json:"size"`
}

// LatestUS is where we find the latest unstable.
const unstableURL = "https://unstable.golift.io"

// CheckUnstable checks if the provided app has an updated version on GitHub.
// Pass in revision only, no version.
func CheckUnstable(ctx context.Context, app string, revision string) (*Update, error) {
	uri := fmt.Sprintf("%s/%s/%s.%s.installer.exe", unstableURL, strings.ToLower(app), app, runtime.GOARCH)
	if runtime.GOOS == "linux" {
		uri = fmt.Sprintf("%s/%s/%s.%s.gz", unstableURL, strings.ToLower(app), app, runtime.GOARCH)
	} else if runtime.GOOS == "darwin" {
		uri = fmt.Sprintf("%s/%s/%s.dmg", unstableURL, strings.ToLower(app), app)
	}

	release, err := GetUnstable(ctx, uri)
	if err != nil {
		return nil, err
	}

	oldRev, _ := strconv.Atoi(revision)

	return &Update{
		RelDate: release.Time,
		CurrURL: release.File,
		Current: fmt.Sprint(release.Ver, "-", release.Rev),
		Version: revision, // on well.
		RelSize: release.Size,
		Outdate: release.Rev > oldRev,
	}, nil
}

// GetUnstable returns an unstable release. See CheckUnstable for an example on how to use it.
func GetUnstable(ctx context.Context, uri string) (*UnstableFile, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri+".txt", nil)
	if err != nil {
		return nil, fmt.Errorf("requesting unstable: %w", err)
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("querying unstable: %w", err)
	}
	defer resp.Body.Close()

	var release UnstableFile
	if err = json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decoding unstable response: %w", err)
	}

	release.Time, _ = time.Parse(time.RFC1123, resp.Header.Get("last-modified"))
	release.File = uri

	return &release, nil
}
