//nolint:dupl
package dnclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"golift.io/starr/readarr"
)

// readarrHandlers is called once on startup to register the web API paths.
func (c *Client) readarrHandlers() {
	c.handleAPIpath(Readarr, "/add/{id:[0-9]+}", c.readarrAddBook, "POST")
	c.handleAPIpath(Readarr, "/check/{id:[0-9]+}/{grid:[0-9]+}", c.readarrCheckBook, "GET")
	c.handleAPIpath(Readarr, "/metadataProfiles/{id:[0-9]+}", c.readarrMetaProfiles, "GET")
	c.handleAPIpath(Readarr, "/qualityProfiles/{id:[0-9]+}", c.readarrProfiles, "GET")
	c.handleAPIpath(Readarr, "/rootFolder/{id:[0-9]+}", c.readarrRootFolders, "GET")
}

func (c *Client) readarrRootFolders(r *http.Request) (int, interface{}) {
	// Get folder list from Readarr.
	folders, err := getReadarr(r).GetRootFolders()
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
	// Get the metadata profiles from readarr.
	profiles, err := getReadarr(r).GetMetadataProfiles()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting profiles: %w", err)
	}

	// Format profile ID=>Name into a nice map.
	p := make(map[int64]string)
	for i := range profiles {
		p[profiles[i].ID] = profiles[i].Name
	}

	return http.StatusOK, p
}

func (c *Client) readarrProfiles(r *http.Request) (int, interface{}) {
	// Get the profiles from readarr.
	profiles, err := getReadarr(r).GetQualityProfiles()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting profiles: %w", err)
	}

	// Format profile ID=>Name into a nice map.
	p := make(map[int64]string)
	for i := range profiles {
		p[profiles[i].ID] = profiles[i].Name
	}

	return http.StatusOK, p
}

func (c *Client) readarrCheckBook(r *http.Request) (int, interface{}) {
	grid, _ := strconv.ParseInt(mux.Vars(r)["grid"], 10, 64)
	// Check for existing book.
	if m, err := getReadarr(r).GetBook(grid); err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking book: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, fmt.Errorf("%d: %w", grid, ErrExists)
	}

	return http.StatusOK, http.StatusText(http.StatusNotFound)
}

func (c *Client) readarrAddBook(r *http.Request) (int, interface{}) {
	payload := &readarr.AddBookInput{}
	// Extract payload and check for TMDB ID.
	if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	} else if payload.ForeignBookID == 0 {
		return http.StatusUnprocessableEntity, fmt.Errorf("0: %w", ErrNoGRID)
	}

	readar := getReadarr(r)
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
