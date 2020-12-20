//nolint:dupl
package dnclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"golift.io/starr"
	"golift.io/starr/readarr"
)

// ReadarrConfig represents the input data for a Readarr server.
type ReadarrConfig struct {
	*starr.Config
	*readarr.Readarr
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

// readarrHandlers is called once on startup to register the web API paths.
func (c *Client) readarrHandlers() {
	c.serveAPIpath(Readarr, "/add/{id:[0-9]+}", "POST", c.readarrAddBook)
	c.serveAPIpath(Readarr, "/check/{id:[0-9]+}/{grid:[0-9]+}", "GET", c.readarrCheckBook)
	c.serveAPIpath(Readarr, "/metadataProfiles/{id:[0-9]+}", "GET", c.readarrMetaProfiles)
	c.serveAPIpath(Readarr, "/qualityProfiles/{id:[0-9]+}", "GET", c.readarrProfiles)
	c.serveAPIpath(Readarr, "/rootFolder/{id:[0-9]+}", "GET", c.readarrRootFolders)
}

func (c *Client) readarrRootFolders(r *http.Request) (int, interface{}) {
	readar := getReadarr(r)

	// Get folder list from Readarr.
	folders, err := readar.GetRootFolders()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting folders: %w", err)
	}

	// Format folder list into a nice path=>freesSpace map.
	p := make(map[string]int64)
	for i := range folders {
		p[folders[i].Path] = folders[i].FreeSpace
	}

	return http.StatusOK, p
}

func (c *Client) readarrMetaProfiles(r *http.Request) (int, interface{}) {
	readar := getReadarr(r)

	// Get the metadata profiles from readarr.
	profiles, err := readar.GetMetadataProfiles()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting profiles: %w", err)
	}

	// Format profile ID=>Name into a nice map.
	p := make(map[int]string)
	for i := range profiles {
		p[profiles[i].ID] = profiles[i].Name
	}

	return http.StatusOK, p
}

func (c *Client) readarrProfiles(r *http.Request) (int, interface{}) {
	readar := getReadarr(r)

	// Get the profiles from readarr.
	profiles, err := readar.GetQualityProfiles()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting profiles: %w", err)
	}

	// Format profile ID=>Name into a nice map.
	p := make(map[int]string)
	for i := range profiles {
		p[profiles[i].ID] = profiles[i].Name
	}

	return http.StatusOK, p
}

func (c *Client) readarrCheckBook(r *http.Request) (int, interface{}) {
	readar := getReadarr(r)
	grid, _ := strconv.Atoi(mux.Vars(r)["grid"])

	// Check for existing book.
	if m, err := readar.GetBook(grid); err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking book: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, fmt.Errorf("%d: %w", grid, ErrExists)
	}

	return http.StatusOK, http.StatusText(http.StatusNotFound)
}

func (c *Client) readarrAddBook(r *http.Request) (int, interface{}) {
	readar := getReadarr(r)

	// Extract payload and check for TMDB ID.
	payload := &readarr.AddBookInput{}
	if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	} else if payload.ForeignBookID == 0 {
		return http.StatusUnprocessableEntity, fmt.Errorf("0: %w", ErrNoGRID)
	}

	// Check for existing book.
	if m, err := readar.GetBook(payload.ForeignBookID); err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking book: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, fmt.Errorf("%d: %w", payload.ForeignBookID, ErrExists)
	}

	// Add book using payload.
	book, err := readar.AddBook(payload)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding book: %w", err)
	}

	return http.StatusCreated, book
}
