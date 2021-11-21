package apps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/gorilla/mux"
	"golift.io/cnfg"
	"golift.io/starr"
	"golift.io/starr/readarr"
)

// readarrHandlers is called once on startup to register the web API paths.
func (a *Apps) readarrHandlers() {
	a.HandleAPIpath(starr.Readarr, "/add", readarrAddBook, "POST")
	a.HandleAPIpath(starr.Readarr, "/author/{authorid:[0-9]+}", readarrGetAuthor, "GET")
	a.HandleAPIpath(starr.Readarr, "/check/{grid:[0-9]+}", readarrCheckBook, "GET")
	a.HandleAPIpath(starr.Readarr, "/get/{bookid:[0-9]+}", readarrGetBook, "GET")
	a.HandleAPIpath(starr.Readarr, "/metadataProfiles", readarrMetaProfiles, "GET")
	a.HandleAPIpath(starr.Readarr, "/qualityProfiles", readarrQualityProfiles, "GET")
	a.HandleAPIpath(starr.Readarr, "/qualityProfile", readarrGetQualityProfile, "GET")
	a.HandleAPIpath(starr.Readarr, "/qualityProfile", readarrAddQualityProfile, "POST")
	a.HandleAPIpath(starr.Readarr, "/qualityProfile/{profileID:[0-9]+}", readarrUpdateQualityProfile, "PUT")
	a.HandleAPIpath(starr.Readarr, "/rootFolder", readarrRootFolders, "GET")
	a.HandleAPIpath(starr.Readarr, "/search/{query}", readarrSearchBook, "GET")
	a.HandleAPIpath(starr.Readarr, "/update", readarrUpdateBook, "PUT")
	a.HandleAPIpath(starr.Readarr, "/tag", readarrGetTags, "GET")
	a.HandleAPIpath(starr.Readarr, "/tag/{tid:[0-9]+}/{label}", readarrUpdateTag, "PUT")
	a.HandleAPIpath(starr.Readarr, "/tag/{label}", readarrSetTag, "PUT")
	a.HandleAPIpath(starr.Readarr, "/updateauthor", readarrUpdateAuthor, "PUT")
	a.HandleAPIpath(starr.Readarr, "/command/search/{bookid:[0-9]+}", readarrTriggerSearchBook, "GET")
}

// ReadarrConfig represents the input data for a Readarr server.
type ReadarrConfig struct {
	Name      string        `toml:"name" xml:"name"`
	Interval  cnfg.Duration `toml:"interval" xml:"interval"`
	StuckItem bool          `toml:"stuck_items" xml:"stuck_items"`
	CheckQ    *uint         `toml:"check_q" xml:"check_q"`
	*starr.Config
	*readarr.Readarr
	Errorf func(string, ...interface{}) `toml:"-" xml:"-"`
}

func (a *Apps) setupReadarr(timeout time.Duration) error {
	for i := range a.Readarr {
		if a.Readarr[i].Config == nil || a.Readarr[i].Config.URL == "" {
			return fmt.Errorf("%w: missing url: Readarr config %d", ErrInvalidApp, i+1)
		}

		a.Readarr[i].Debugf = a.DebugLog.Printf
		a.Readarr[i].Errorf = a.ErrorLog.Printf
		a.Readarr[i].setup(timeout)
	}

	return nil
}

func (r *ReadarrConfig) setup(timeout time.Duration) {
	r.Readarr = readarr.New(r.Config)
	if r.Timeout.Duration == 0 {
		r.Timeout.Duration = timeout
	}

	// These things are not used in this package but this package configures them.
	if r.StuckItem && r.CheckQ == nil {
		i := uint(0)
		r.CheckQ = &i
	} else if r.CheckQ != nil {
		r.StuckItem = true
	}

	r.URL = strings.TrimRight(r.URL, "/")

	if u, err := r.GetURL(); err != nil {
		r.Errorf("Checking Readarr Path: %v", err)
	} else if u = strings.TrimRight(u, "/"); u != r.URL {
		r.Errorf("Readarr URL fixed: %s -> %s (continuing)", r.URL, u)
		r.URL = u
	}
}

func readarrAddBook(r *http.Request) (int, interface{}) {
	payload := &readarr.AddBookInput{}
	// Extract payload and check for GRID ID.
	switch err := json.NewDecoder(r.Body).Decode(payload); {
	case err != nil:
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	case len(payload.Editions) != 1:
		return http.StatusUnprocessableEntity,
			fmt.Errorf("invalid editions count; only 1 allowed: %d, %w", len(payload.Editions), ErrNoGRID)
	case payload.Editions[0].ForeignEditionID == "":
		return http.StatusUnprocessableEntity, fmt.Errorf("0: %w", ErrNoGRID)
	}

	app := getReadarr(r)
	// Check for existing book.
	m, err := app.GetBook(payload.Editions[0].ForeignEditionID)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking book: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, readarrData(m[0])
	}

	// Add book using payload.
	book, err := app.AddBook(payload)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding book: %w", err)
	}

	return http.StatusCreated, book
}

func readarrGetAuthor(r *http.Request) (int, interface{}) {
	authorID, _ := strconv.ParseInt(mux.Vars(r)["authorid"], mnd.Base10, mnd.Bits64)

	author, err := getReadarr(r).GetAuthorByID(authorID)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("getting author: %w", err)
	}

	return http.StatusOK, author
}

func readarrData(book *readarr.Book) map[string]interface{} {
	hasFile := false
	if book.Statistics != nil {
		hasFile = book.Statistics.SizeOnDisk > 0
	}

	return map[string]interface{}{
		"id":        book.ID,
		"hasFile":   hasFile,
		"monitored": book.Monitored,
	}
}

// Check for existing book.
func readarrCheckBook(r *http.Request) (int, interface{}) {
	grid := mux.Vars(r)["grid"]

	m, err := getReadarr(r).GetBook(grid)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking book: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, readarrData(m[0])
	}

	return http.StatusOK, http.StatusText(http.StatusNotFound)
}

func readarrGetBook(r *http.Request) (int, interface{}) {
	bookID, _ := strconv.ParseInt(mux.Vars(r)["bookid"], mnd.Base10, mnd.Bits64)

	book, err := getReadarr(r).GetBookByID(bookID)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking book: %w", err)
	}

	return http.StatusOK, book
}

func readarrTriggerSearchBook(r *http.Request) (int, interface{}) {
	bookID, _ := strconv.ParseInt(mux.Vars(r)["bookid"], mnd.Base10, mnd.Bits64)

	output, err := getReadarr(r).SendCommand(&readarr.CommandRequest{
		Name:    "BookSearch",
		BookIDs: []int64{bookID},
	})
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking book: %w", err)
	}

	return http.StatusOK, output.Status
}

// Get the metadata profiles from readarr.
func readarrMetaProfiles(r *http.Request) (int, interface{}) {
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

// Get the quality profiles from readarr.
func readarrQualityProfiles(r *http.Request) (int, interface{}) {
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

// Get the all quality profiles data from readarr.
func readarrGetQualityProfile(r *http.Request) (int, interface{}) {
	profiles, err := getReadarr(r).GetQualityProfiles()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting profiles: %w", err)
	}

	return http.StatusOK, profiles
}

func readarrAddQualityProfile(r *http.Request) (int, interface{}) {
	var profile readarr.QualityProfile

	// Extract payload and check for TMDB ID.
	err := json.NewDecoder(r.Body).Decode(&profile)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	// Get the profiles from radarr.
	id, err := getReadarr(r).AddQualityProfile(&profile)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding profile: %w", err)
	}

	return http.StatusOK, id
}

func readarrUpdateQualityProfile(r *http.Request) (int, interface{}) {
	var profile readarr.QualityProfile

	// Extract payload and check for TMDB ID.
	err := json.NewDecoder(r.Body).Decode(&profile)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	profile.ID, _ = strconv.ParseInt(mux.Vars(r)["profileID"], mnd.Base10, mnd.Bits64)
	if profile.ID == 0 {
		return http.StatusBadRequest, ErrNonZeroID
	}

	// Get the profiles from radarr.
	err = getReadarr(r).UpdateQualityProfile(&profile)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("updating profile: %w", err)
	}

	return http.StatusOK, "OK"
}

// Get folder list from Readarr.
func readarrRootFolders(r *http.Request) (int, interface{}) {
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

func readarrSearchBook(r *http.Request) (int, interface{}) {
	books, err := getReadarr(r).GetBook("")
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("getting books: %w", err)
	}

	query := strings.TrimSpace(strings.ToLower(mux.Vars(r)["query"])) // in
	returnBooks := make([]map[string]interface{}, 0)                  // out

	for _, book := range books {
		if bookSearch(query, book.Title, book.Editions) {
			b := map[string]interface{}{
				"id":       book.ID,
				"title":    book.Title,
				"release":  book.ReleaseDate,
				"author":   book.Author.AuthorName,
				"authorId": book.Author.ID,
				"overview": book.Overview,
				"ratings":  book.Ratings.Value,
				"pages":    book.PageCount,
				"exists":   false,
				"files":    0,
			}

			if book.Statistics != nil {
				b["files"] = book.Statistics.BookFileCount
				b["exists"] = book.Statistics.SizeOnDisk > 0
			}

			returnBooks = append(returnBooks, b)
		}
	}

	return http.StatusOK, returnBooks
}

func bookSearch(query, title string, editions []*readarr.Edition) bool {
	if strings.Contains(strings.ToLower(title), query) {
		return true
	}

	for _, t := range editions {
		if strings.Contains(strings.ToLower(t.Title), query) {
			return true
		}
	}

	return false
}

func readarrGetTags(r *http.Request) (int, interface{}) {
	tags, err := getReadarr(r).GetTags()
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("getting tags: %w", err)
	}

	return http.StatusOK, tags
}

func readarrUpdateTag(r *http.Request) (int, interface{}) {
	id, _ := strconv.Atoi(mux.Vars(r)["tid"])

	tagID, err := getReadarr(r).UpdateTag(id, mux.Vars(r)["label"])
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating tag: %w", err)
	}

	return http.StatusOK, tagID
}

func readarrSetTag(r *http.Request) (int, interface{}) {
	tagID, err := getReadarr(r).AddTag(mux.Vars(r)["label"])
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("setting tag: %w", err)
	}

	return http.StatusOK, tagID
}

func readarrUpdateBook(r *http.Request) (int, interface{}) {
	var book readarr.Book

	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	err = getReadarr(r).UpdateBook(book.ID, &book)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating book: %w", err)
	}

	return http.StatusOK, "readarr seems to have worked"
}

func readarrUpdateAuthor(r *http.Request) (int, interface{}) {
	var author readarr.Author

	err := json.NewDecoder(r.Body).Decode(&author)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	err = getReadarr(r).UpdateAuthor(author.ID, &author)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating author: %w", err)
	}

	return http.StatusOK, "readarr seems to have worked"
}
