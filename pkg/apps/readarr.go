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
	starrConfig
	*starr.Config
	*readarr.Readarr `toml:"-" xml:"-" json:"-"`
	errorf           func(string, ...interface{}) `toml:"-" xml:"-" json:"-"`
}

func (a *Apps) setupReadarr(timeout time.Duration) error {
	for idx := range a.Readarr {
		if a.Readarr[idx].Config == nil || a.Readarr[idx].Config.URL == "" {
			return fmt.Errorf("%w: missing url: Readarr config %d", ErrInvalidApp, idx+1)
		}

		a.Readarr[idx].Debugf = a.Debugf
		a.Readarr[idx].errorf = a.Errorf
		a.Readarr[idx].setup(timeout)
	}

	return nil
}

func (r *ReadarrConfig) setup(timeout time.Duration) {
	r.Readarr = readarr.New(r.Config)
	r.Readarr.APIer = &starrAPI{api: r.Readarr.APIer, app: starr.Readarr.String()}

	if r.Timeout.Duration == 0 {
		r.Timeout.Duration = timeout
	}

	r.URL = strings.TrimRight(r.URL, "/")
}

func readarrAddBook(req *http.Request) (int, interface{}) {
	payload := &readarr.AddBookInput{}
	// Extract payload and check for GRID ID.
	switch err := json.NewDecoder(req.Body).Decode(payload); {
	case err != nil:
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	case len(payload.Editions) != 1:
		return http.StatusUnprocessableEntity,
			fmt.Errorf("invalid editions count; only 1 allowed: %d, %w", len(payload.Editions), ErrNoGRID)
	case payload.Editions[0].ForeignEditionID == "":
		return http.StatusUnprocessableEntity, fmt.Errorf("0: %w", ErrNoGRID)
	}

	// Check for existing book.
	m, err := getReadarr(req).GetBookContext(req.Context(), payload.Editions[0].ForeignEditionID)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking book: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, readarrData(m[0])
	}

	// Add book using payload.
	book, err := getReadarr(req).AddBookContext(req.Context(), payload)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding book: %w", err)
	}

	return http.StatusCreated, book
}

func readarrGetAuthor(req *http.Request) (int, interface{}) {
	authorID, _ := strconv.ParseInt(mux.Vars(req)["authorid"], mnd.Base10, mnd.Bits64)

	author, err := getReadarr(req).GetAuthorByIDContext(req.Context(), authorID)
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
		"tags":      book.Author.Tags,
	}
}

// Check for existing book.
func readarrCheckBook(req *http.Request) (int, interface{}) {
	grid := mux.Vars(req)["grid"]

	m, err := getReadarr(req).GetBookContext(req.Context(), grid)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking book: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, readarrData(m[0])
	}

	return http.StatusOK, http.StatusText(http.StatusNotFound)
}

func readarrGetBook(req *http.Request) (int, interface{}) {
	bookID, _ := strconv.ParseInt(mux.Vars(req)["bookid"], mnd.Base10, mnd.Bits64)

	book, err := getReadarr(req).GetBookByIDContext(req.Context(), bookID)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking book: %w", err)
	}

	return http.StatusOK, book
}

func readarrTriggerSearchBook(req *http.Request) (int, interface{}) {
	bookID, _ := strconv.ParseInt(mux.Vars(req)["bookid"], mnd.Base10, mnd.Bits64)

	output, err := getReadarr(req).SendCommandContext(req.Context(), &readarr.CommandRequest{
		Name:    "BookSearch",
		BookIDs: []int64{bookID},
	})
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking book: %w", err)
	}

	return http.StatusOK, output.Status
}

// Get the metadata profiles from readarr.
func readarrMetaProfiles(req *http.Request) (int, interface{}) {
	profiles, err := getReadarr(req).GetMetadataProfilesContext(req.Context())
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
func readarrQualityProfiles(req *http.Request) (int, interface{}) {
	profiles, err := getReadarr(req).GetQualityProfilesContext(req.Context())
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
func readarrGetQualityProfile(req *http.Request) (int, interface{}) {
	profiles, err := getReadarr(req).GetQualityProfilesContext(req.Context())
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting profiles: %w", err)
	}

	return http.StatusOK, profiles
}

func readarrAddQualityProfile(req *http.Request) (int, interface{}) {
	var profile readarr.QualityProfile

	// Extract payload and check for TMDB ID.
	err := json.NewDecoder(req.Body).Decode(&profile)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	// Get the profiles from radarr.
	id, err := getReadarr(req).AddQualityProfileContext(req.Context(), &profile)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding profile: %w", err)
	}

	return http.StatusOK, id
}

func readarrUpdateQualityProfile(req *http.Request) (int, interface{}) {
	var profile readarr.QualityProfile

	// Extract payload and check for TMDB ID.
	err := json.NewDecoder(req.Body).Decode(&profile)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	profile.ID, _ = strconv.ParseInt(mux.Vars(req)["profileID"], mnd.Base10, mnd.Bits64)
	if profile.ID == 0 {
		return http.StatusBadRequest, ErrNonZeroID
	}

	// Get the profiles from radarr.
	err = getReadarr(req).UpdateQualityProfileContext(req.Context(), &profile)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("updating profile: %w", err)
	}

	return http.StatusOK, "OK"
}

// Get folder list from Readarr.
func readarrRootFolders(req *http.Request) (int, interface{}) {
	folders, err := getReadarr(req).GetRootFoldersContext(req.Context())
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

func readarrSearchBook(req *http.Request) (int, interface{}) {
	books, err := getReadarr(req).GetBookContext(req.Context(), "")
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("getting books: %w", err)
	}

	query := strings.TrimSpace(strings.ToLower(mux.Vars(req)["query"])) // in
	returnBooks := make([]map[string]interface{}, 0)                    // out

	for _, book := range books {
		if bookSearch(query, book.Title, book.Editions) {
			item := map[string]interface{}{
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
				item["files"] = book.Statistics.BookFileCount
				item["exists"] = book.Statistics.SizeOnDisk > 0
			}

			returnBooks = append(returnBooks, item)
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

func readarrGetTags(req *http.Request) (int, interface{}) {
	tags, err := getReadarr(req).GetTagsContext(req.Context())
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("getting tags: %w", err)
	}

	return http.StatusOK, tags
}

func readarrUpdateTag(req *http.Request) (int, interface{}) {
	id, _ := strconv.Atoi(mux.Vars(req)["tid"])

	tag, err := getReadarr(req).UpdateTagContext(req.Context(), &starr.Tag{ID: id, Label: mux.Vars(req)["label"]})
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating tag: %w", err)
	}

	return http.StatusOK, tag.ID
}

func readarrSetTag(req *http.Request) (int, interface{}) {
	tag, err := getReadarr(req).AddTagContext(req.Context(), &starr.Tag{Label: mux.Vars(req)["label"]})
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("setting tag: %w", err)
	}

	return http.StatusOK, tag.ID
}

func readarrUpdateBook(req *http.Request) (int, interface{}) {
	var book readarr.Book

	err := json.NewDecoder(req.Body).Decode(&book)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	err = getReadarr(req).UpdateBookContext(req.Context(), book.ID, &book)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating book: %w", err)
	}

	return http.StatusOK, "readarr seems to have worked"
}

func readarrUpdateAuthor(req *http.Request) (int, interface{}) {
	var author readarr.Author

	err := json.NewDecoder(req.Body).Decode(&author)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	err = getReadarr(req).UpdateAuthorContext(req.Context(), author.ID, &author)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating author: %w", err)
	}

	return http.StatusOK, "readarr seems to have worked"
}
