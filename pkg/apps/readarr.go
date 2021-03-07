package apps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"golift.io/starr"
	"golift.io/starr/readarr"
)

// readarrHandlers is called once on startup to register the web API paths.
func (a *Apps) readarrHandlers() {
	a.HandleAPIpath(Readarr, "/add", readarrAddBook, "POST")
	a.HandleAPIpath(Readarr, "/author/{authorid:[0-9]+}", readarrGetAuthor, "GET")
	a.HandleAPIpath(Readarr, "/check/{grid:[0-9]+}", readarrCheckBook, "GET")
	a.HandleAPIpath(Readarr, "/get/{bookid:[0-9]+}", readarrGetBook, "GET")
	a.HandleAPIpath(Readarr, "/metadataProfiles", readarrMetaProfiles, "GET")
	a.HandleAPIpath(Readarr, "/qualityProfiles", readarrProfiles, "GET")
	a.HandleAPIpath(Readarr, "/rootFolder", readarrRootFolders, "GET")
	a.HandleAPIpath(Readarr, "/search/{query}", readarrSearchBook, "GET")
	a.HandleAPIpath(Readarr, "/update", readarrUpdateBook, "PUT")
	a.HandleAPIpath(Readarr, "/tag", readarrGetTags, "GET")
	a.HandleAPIpath(Readarr, "/tag/{tid:[0-9]+}/{label}", readarrUpdateTag, "PUT")
	a.HandleAPIpath(Readarr, "/tag/{label}", readarrSetTag, "PUT")
	a.HandleAPIpath(Readarr, "/updateauthor", readarrUpdateAuthor, "PUT")
}

// ReadarrConfig represents the input data for a Readarr server.
type ReadarrConfig struct {
	*starr.Config
	readarr *readarr.Readarr
}

func (r *ReadarrConfig) setup(timeout time.Duration) {
	r.readarr = readarr.New(r.Config)
	if r.Timeout.Duration == 0 {
		r.Timeout.Duration = timeout
	}
}

func readarrAddBook(r *http.Request) (int, interface{}) {
	payload := &readarr.AddBookInput{}
	// Extract payload and check for GRID ID.
	err := json.NewDecoder(r.Body).Decode(payload)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	} else if payload.ForeignBookID == 0 {
		return http.StatusUnprocessableEntity, fmt.Errorf("0: %w", ErrNoGRID)
	}

	app := getReadarr(r)
	// Check for existing book.
	m, err := app.GetBook(payload.ForeignBookID)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking book: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, fmt.Errorf("%d: %w", payload.ForeignBookID, ErrExists)
	}

	// Add book using payload.
	book, err := app.AddBook(payload)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding book: %w", err)
	}

	return http.StatusCreated, book
}

func readarrGetAuthor(r *http.Request) (int, interface{}) {
	authorID, _ := strconv.ParseInt(mux.Vars(r)["authorid"], 10, 64)

	author, err := getReadarr(r).GetAuthorByID(authorID)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("getting author: %w", err)
	}

	return http.StatusOK, author
}

// Check for existing book.
func readarrCheckBook(r *http.Request) (int, interface{}) {
	grid, _ := strconv.ParseInt(mux.Vars(r)["grid"], 10, 64)

	m, err := getReadarr(r).GetBook(grid)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking book: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, fmt.Errorf("%d: %w", grid, ErrExists)
	}

	return http.StatusOK, http.StatusText(http.StatusNotFound)
}

func readarrGetBook(r *http.Request) (int, interface{}) {
	bookID, _ := strconv.ParseInt(mux.Vars(r)["bookid"], 10, 64)

	book, err := getReadarr(r).GetBookByID(bookID)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking book: %w", err)
	}

	return http.StatusOK, book
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

// Get the profiles from readarr.
func readarrProfiles(r *http.Request) (int, interface{}) {
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
	books, err := getReadarr(r).GetBook(0)
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
