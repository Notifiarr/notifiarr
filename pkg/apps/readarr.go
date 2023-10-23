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
	"golift.io/starr/debuglog"
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
	a.HandleAPIpath(starr.Readarr, "/notification", readarrGetNotifications, "GET")
	a.HandleAPIpath(starr.Readarr, "/notification", readarrUpdateNotification, "PUT")
	a.HandleAPIpath(starr.Readarr, "/notification", readarrAddNotification, "POST")
}

// ReadarrConfig represents the input data for a Readarr server.
type ReadarrConfig struct {
	ExtraConfig
	*starr.Config
	*readarr.Readarr `toml:"-" xml:"-" json:"-"`
	errorf           func(string, ...interface{}) `toml:"-" xml:"-" json:"-"`
}

func getReadarr(r *http.Request) *readarr.Readarr {
	app, _ := r.Context().Value(starr.Readarr).(*ReadarrConfig)
	return app.Readarr
}

// Enabled returns true if the Readarr instance is enabled and usable.
func (r *ReadarrConfig) Enabled() bool {
	return r != nil && r.Config != nil && r.URL != "" && r.APIKey != "" && r.Timeout.Duration >= 0
}

func (a *Apps) setupReadarr() error {
	for idx, app := range a.Readarr {
		if app.Config == nil || app.Config.URL == "" {
			return fmt.Errorf("%w: missing url: Readarr config %d", ErrInvalidApp, idx+1)
		} else if !strings.HasPrefix(app.Config.URL, "http://") && !strings.HasPrefix(app.Config.URL, "https://") {
			return fmt.Errorf("%w: URL must begin with http:// or https://: Readarr config %d", ErrInvalidApp, idx+1)
		}

		if a.Logger.DebugEnabled() {
			app.Config.Client = starr.ClientWithDebug(app.Timeout.Duration, app.ValidSSL, debuglog.Config{
				MaxBody: a.MaxBody,
				Debugf:  a.Debugf,
				Caller:  metricMakerCallback(string(starr.Readarr)),
				Redact:  []string{app.APIKey, app.Password, app.HTTPPass},
			})
		} else {
			app.Config.Client = starr.Client(app.Timeout.Duration, app.ValidSSL)
			app.Config.Client.Transport = NewMetricsRoundTripper(starr.Readarr.String(), app.Config.Client.Transport)
		}

		app.errorf = a.Errorf
		app.URL = strings.TrimRight(app.URL, "/")
		app.Readarr = readarr.New(app.Config)
	}

	return nil
}

// @Description  Adds a new Book to Readarr.
// @Summary      Add Readarr Book
// @Tags         Readarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Param        POST body readarr.AddBookInput true "new item content"
// @Accept       json
// @Success      201  {object} apps.Respond.apiResponse{message=readarr.Book} "created"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "bad json payload"
// @Failure      409  {object} apps.Respond.apiResponse{message=string} "item already exists"
// @Failure      422  {object} apps.Respond.apiResponse{message=string} "no valid editions provided"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error during check"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error during add"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/readarr/{instance}/add [post]
// @Security     ApiKeyAuth
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

// @Description  Fetches an Author from Readarr.
// @Summary      Get Readarr Author
// @Tags         Readarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Param        authorID  path   int64  true  "author ID"
// @Success      200  {object} apps.Respond.apiResponse{message=readarr.Author} "author content"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/readarr/{instance}/author/{authorID} [get]
// @Security     ApiKeyAuth
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
		// "tags":      book.Author.Tags,
	}
}

// Check for existing book.
// @Description  Checks if a book already exists in Readarr.
// @Summary      Check Readarr Book
// @Tags         Readarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Param        gridID    path   int64  true  "Good Reads ID"
// @Success      200  {object} apps.Respond.apiResponse{message=string} "not found"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      409  {object} apps.Respond.apiResponse{message=string} "already exists"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/readarr/{instance}/check/{gridID} [get]
// @Security     ApiKeyAuth
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

// @Description  Returns a book from Readarr.
// @Summary      Get Readarr Book
// @Tags         Readarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Param        bookID    path   int64  true  "Book ID"
// @Success      200  {object} apps.Respond.apiResponse{message=string} "not found"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/readarr/{instance}/get/{bookID} [get]
// @Security     ApiKeyAuth
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
// @Description  Fetches all Metadata Profiles from Readarr.
// @Summary      Get Readarr Metadata Profiles
// @Tags         Readarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Success      200  {object} apps.Respond.apiResponse{message=map[int64]string} "map of ID to name"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/readarr/{instance}/metadataProfiles [get]
// @Security     ApiKeyAuth
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
// @Description  Fetches all Quality Profiles from Readarr.
// @Summary      Get Readarr Quality Profiles
// @Tags         Readarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Success      200  {object} apps.Respond.apiResponse{message=map[int64]string} "map of ID to name"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/readarr/{instance}/qualityProfiles [get]
// @Security     ApiKeyAuth
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
// @Description  Fetches all Quality Profiles Data from Readarr.
// @Summary      Get Readarr Quality Profile Data
// @Tags         Readarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Success      200  {object} apps.Respond.apiResponse{message=[]readarr.QualityProfile} "all profiles"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/readarr/{instance}/qualityProfile [get]
// @Security     ApiKeyAuth
func readarrGetQualityProfile(req *http.Request) (int, interface{}) {
	profiles, err := getReadarr(req).GetQualityProfilesContext(req.Context())
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting profiles: %w", err)
	}

	return http.StatusOK, profiles
}

// @Description  Creates a new Readarr Quality Profile.
// @Summary      Add Readarr Quality Profile
// @Tags         Readarr
// @Produce      json
// @Accept       json
// @Param        instance  path   int64  true  "instance ID"
// @Param        POST body readarr.QualityProfile true "new item content"
// @Success      200  {object} apps.Respond.apiResponse{message=int64} "new profile ID"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "json input error"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/readarr/{instance}/qualityProfile [post]
// @Security     ApiKeyAuth
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

// @Description  Updates a Readarr Quality Profile.
// @Summary      Update Readarr Quality Profile
// @Tags         Readarr
// @Produce      json
// @Accept       json
// @Param        instance  path   int64  true  "instance ID"
// @Param        profileID  path   int64  true  "profile ID to update"
// @Param        PUT body readarr.QualityProfile true "updated item content"
// @Success      200  {object} string "ok"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "json input error"
// @Failure      422  {object} apps.Respond.apiResponse{message=string} "no profile ID"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/readarr/{instance}/qualityProfile/{profileID} [put]
// @Security     ApiKeyAuth
func readarrUpdateQualityProfile(req *http.Request) (int, interface{}) {
	var profile readarr.QualityProfile

	// Extract payload and check for TMDB ID.
	err := json.NewDecoder(req.Body).Decode(&profile)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	profile.ID, _ = strconv.ParseInt(mux.Vars(req)["profileID"], mnd.Base10, mnd.Bits64)
	if profile.ID == 0 {
		return http.StatusUnprocessableEntity, ErrNonZeroID
	}

	// Get the profiles from radarr.
	_, err = getReadarr(req).UpdateQualityProfileContext(req.Context(), &profile)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("updating profile: %w", err)
	}

	return http.StatusOK, "OK"
}

// Get folder list from Readarr.
// @Description  Returns all Readarr Root Folders paths and free space.
// @Summary      Retrieve Readarr Root Folders
// @Tags         Readarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Success      200  {object} apps.Respond.apiResponse{message=map[string]int64} "map of path->space free"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/readarr/{instance}/rootFolder [get]
// @Security     ApiKeyAuth
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

// @Description  Searches all Book Titles for the search term provided.
// @Summary      Search for Readarr Books
// @Tags         Readarr
// @Produce      json
// @Param        query     path   string  true  "title search string"
// @Param        instance  path   int64   true  "instance ID"
// @Success      200  {object} apps.Respond.apiResponse{message=[]apps.readarrSearchBook.bookData}  "minimal book data"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/readarr/{instance}/search/{query} [get]
// @Security     ApiKeyAuth
func readarrSearchBook(req *http.Request) (int, interface{}) {
	books, err := getReadarr(req).GetBookContext(req.Context(), "")
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("getting books: %w", err)
	}

	type bookData struct {
		// Book ID
		ID int64 `json:"id"`
		// Book Title
		Title string `json:"title"`
		// Release Date
		Release time.Time `json:"release"`
		// Author ID
		AuthorID int64 `json:"authorId"`
		// Author Name
		Author string `json:"author"`
		// Book Overview
		Overview string `json:"overview"`
		// Rating Value
		Ratings float64 `json:"ratings"`
		// Book Page count
		Pages int `json:"pages"`
		// Exists on disk or not?
		Exists bool `json:"exists"`
		// Number of files on disk for this book.
		Files int `json:"files"`
	}

	query := strings.TrimSpace(strings.ToLower(mux.Vars(req)["query"])) // in
	returnBooks := make([]*bookData, 0)                                 // out

	for _, book := range books {
		if bookSearch(query, book.Title, book.Editions) {
			item := &bookData{
				ID:       book.ID,
				Title:    book.Title,
				Release:  book.ReleaseDate,
				Author:   book.AuthorTitle,
				AuthorID: book.AuthorID,
				Overview: book.Overview,
				Ratings:  book.Ratings.Value,
				Pages:    book.PageCount,
			}

			if book.Statistics != nil {
				item.Files = book.Statistics.BookFileCount
				item.Exists = book.Statistics.SizeOnDisk > 0
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

// @Description  Returns all Readarr Tags
// @Summary      Retrieve Readarr Tags
// @Tags         Readarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Success      200  {object} apps.Respond.apiResponse{message=[]starr.Tag} "tags"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/readarr/{instance}/tag [get]
// @Security     ApiKeyAuth
func readarrGetTags(req *http.Request) (int, interface{}) {
	tags, err := getReadarr(req).GetTagsContext(req.Context())
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("getting tags: %w", err)
	}

	return http.StatusOK, tags
}

// @Description  Updates the label for a an existing tag.
// @Summary      Update Readarr Tag Label
// @Tags         Readarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Param        tagID     path   int64  true  "tag ID to update"
// @Param        label     path   string  true  "new label"
// @Success      200  {object} apps.Respond.apiResponse{message=int64}  "tag ID"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/readarr/{instance}/tag/{tagID}/{label} [put]
// @Security     ApiKeyAuth
func readarrUpdateTag(req *http.Request) (int, interface{}) {
	id, _ := strconv.Atoi(mux.Vars(req)["tid"])

	tag, err := getReadarr(req).UpdateTagContext(req.Context(), &starr.Tag{ID: id, Label: mux.Vars(req)["label"]})
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating tag: %w", err)
	}

	return http.StatusOK, tag.ID
}

// @Description  Creates a new tag with the provided label.
// @Summary      Create Readarr Tag
// @Tags         Readarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Param        label     path   string true  "new tag's label"
// @Success      200  {object} apps.Respond.apiResponse{message=int64}  "tag ID"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/readarr/{instance}/tag/{label} [put]
// @Security     ApiKeyAuth
func readarrSetTag(req *http.Request) (int, interface{}) {
	tag, err := getReadarr(req).AddTagContext(req.Context(), &starr.Tag{Label: mux.Vars(req)["label"]})
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("setting tag: %w", err)
	}

	return http.StatusOK, tag.ID
}

// @Description  Updates a Book in Readarr.
// @Summary      Update Readarr Book
// @Tags         Readarr
// @Produce      json
// @Accept       json
// @Param        instance  path  int64  true  "instance ID"
// @Param        moveFiles query int64  true  "move files? true/false"
// @Param        PUT body readarr.Book  true  "book content"
// @Success      200  {object} apps.Respond.apiResponse{message=string} "ok"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "bad json input"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/readarr/{instance}/update [put]
// @Security     ApiKeyAuth
func readarrUpdateBook(req *http.Request) (int, interface{}) {
	var book readarr.Book

	err := json.NewDecoder(req.Body).Decode(&book)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	moveFiles := mux.Vars(req)["moveFiles"] == fmt.Sprint(true)

	err = getReadarr(req).UpdateBookContext(req.Context(), book.ID, &book, moveFiles)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating book: %w", err)
	}

	return http.StatusOK, "readarr seems to have worked"
}

// @Description  Updates an Author in Readarr.
// @Summary      Update Readarr Author
// @Tags         Readarr
// @Produce      json
// @Accept       json
// @Param        instance  path  int64  true  "instance ID"
// @Param        PUT body readarr.Author  true  "author content"
// @Success      200  {object} apps.Respond.apiResponse{message=string} "ok"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "bad json input"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/readarr/{instance}/updateauthor [put]
// @Security     ApiKeyAuth
func readarrUpdateAuthor(req *http.Request) (int, interface{}) {
	var author readarr.Author

	err := json.NewDecoder(req.Body).Decode(&author)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	_, err = getReadarr(req).UpdateAuthorContext(req.Context(), &author, true)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating author: %w", err)
	}

	return http.StatusOK, "readarr seems to have worked"
}

// @Description  Returns Readarr Notifications with a name that matches 'notifiar'.
// @Summary      Retrieve Readarr Notifications
// @Tags         Readarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Success      200  {object} apps.Respond.apiResponse{message=[]readarr.NotificationOutput} "notifications"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/readarr/{instance}/notifications [get]
// @Security     ApiKeyAuth
func readarrGetNotifications(req *http.Request) (int, interface{}) {
	notifs, err := getReadarr(req).GetNotificationsContext(req.Context())
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("getting notifications: %w", err)
	}

	output := []*readarr.NotificationOutput{}

	for _, notif := range notifs {
		if strings.Contains(strings.ToLower(notif.Name), "notifiar") {
			output = append(output, notif)
		}
	}

	return http.StatusOK, output
}

// @Description  Updates a Notification in Readarr.
// @Summary      Update Readarr Notification
// @Tags         Readarr
// @Produce      json
// @Accept       json
// @Param        instance  path   int64  true  "instance ID"
// @Param        PUT body readarr.NotificationInput  true  "notification content"
// @Success      200  {object} apps.Respond.apiResponse{message=string} "ok"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "bad json input"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/readarr/{instance}/notification [put]
// @Security     ApiKeyAuth
func readarrUpdateNotification(req *http.Request) (int, interface{}) {
	var notif readarr.NotificationInput

	err := json.NewDecoder(req.Body).Decode(&notif)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	_, err = getReadarr(req).UpdateNotificationContext(req.Context(), &notif)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating notification: %w", err)
	}

	return http.StatusOK, mnd.Success
}

// @Description  Creates a new Readarr Notification.
// @Summary      Add Readarr Notification
// @Tags         Readarr
// @Produce      json
// @Accept       json
// @Param        instance  path   int64  true  "instance ID"
// @Param        POST body readarr.NotificationInput true "new item content"
// @Success      200  {object} apps.Respond.apiResponse{message=int64} "new notification ID"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "json input error"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/readarr/{instance}/notification [post]
// @Security     ApiKeyAuth
func readarrAddNotification(req *http.Request) (int, interface{}) {
	var notif readarr.NotificationInput

	err := json.NewDecoder(req.Body).Decode(&notif)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	id, err := getReadarr(req).AddNotificationContext(req.Context(), &notif)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("adding notification: %w", err)
	}

	return http.StatusOK, id
}
