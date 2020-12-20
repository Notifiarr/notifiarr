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

func (c *Client) readarrMethods(r *mux.Router) {
	for _, r := range c.Config.Readarr {
		r.Readarr = readarr.New(r.Config)
	}

	r.Handle("/api/readarr/add/{id:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.readarrAddBook))).Methods("POST")
	r.Handle("/api/readarr/check/{id:[0-9]+}/{grid:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.readarrCheckBook))).Methods("GET")
	r.Handle("/api/readarr/metadataProfiles/{id:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.readarrMetaProfiles))).Methods("GET")
	r.Handle("/api/readarr/qualityProfiles/{id:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.readarrProfiles))).Methods("GET")
	r.Handle("/api/readarr/rootFolder/{id:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.readarrRootFolders))).Methods("GET")
}

func (c *Config) fixReadarrConfig() {
	for i := range c.Readarr {
		if c.Readarr[i].Timeout.Duration == 0 {
			c.Readarr[i].Timeout.Duration = c.Timeout.Duration
		}
	}
}

// ReadarrConfig represents the input data for a Readarr server.
type ReadarrConfig struct {
	*starr.Config
	*readarr.Readarr
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

func (c *Client) logReadarr() {
	if count := len(c.Readarr); count == 1 {
		c.Printf(" => Readarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Readarr[0].URL, c.Readarr[0].APIKey != "", c.Readarr[0].Timeout, c.Readarr[0].ValidSSL)
	} else {
		c.Print(" => Readarr Config:", count, "servers")

		for _, f := range c.Readarr {
			c.Printf(" =>    Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}
}

// getReadarr finds a Readarr based on the passed-in ID.
// Every Readarr handler calls this.
func (c *Client) getReadarr(id string) *ReadarrConfig {
	j, _ := strconv.Atoi(id)

	for i, app := range c.Readarr {
		if i != j-1 { // discordnotifier wants 1-indexes
			continue
		}

		return app
	}

	return nil
}

func (c *Client) readarrRootFolders(r *http.Request) (int, interface{}) {
	// Make sure the provided readarr id exists.
	readar := c.getReadarr(mux.Vars(r)["id"])
	if readar == nil {
		return http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", mux.Vars(r)["id"], ErrNoReadarr)
	}

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
	// Make sure the provided readarr id exists.
	readar := c.getReadarr(mux.Vars(r)["id"])
	if readar == nil {
		return http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", mux.Vars(r)["id"], ErrNoReadarr)
	}

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
	// Make sure the provided readarr id exists.
	readar := c.getReadarr(mux.Vars(r)["id"])
	if readar == nil {
		return http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", mux.Vars(r)["id"], ErrNoReadarr)
	}

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
	readar := c.getReadarr(mux.Vars(r)["id"])
	if readar == nil {
		return http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", mux.Vars(r)["id"], ErrNoReadarr)
	}

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
	// Make sure the provided readarr id exists.
	readar := c.getReadarr(mux.Vars(r)["id"])
	if readar == nil {
		return http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", mux.Vars(r)["id"], ErrNoReadarr)
	}

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
