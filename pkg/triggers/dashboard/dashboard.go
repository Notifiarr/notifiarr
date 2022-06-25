package dashboard

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/exp"
	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/plexcron"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/cnfg"
	"golift.io/starr"
	"golift.io/starr/lidarr"
	"golift.io/starr/radarr"
	"golift.io/starr/readarr"
	"golift.io/starr/sonarr"
)

/* This file sends state of affairs to notifiarr.com */
// That is, it collects library data and downloader data.

const TrigDashboard common.TriggerName = "Initiating State Collection for Dashboard."

type Cmd struct {
	*common.Config
	PlexCron *plexcron.Action
}

// Action contains the exported methods for this package.
type Action struct {
	cmd *Cmd
}

// How many "upcoming" or "newest" items to send.
const (
	showNext   = 10
	showLatest = 5
)

// Sortable holds data about any Starr item. Kind of a generic data store.
type Sortable struct {
	id      int64
	Name    string    `json:"name"`
	Sub     string    `json:"subName,omitempty"`
	Date    time.Time `json:"date"`
	Season  int64     `json:"season,omitempty"`
	Episode int64     `json:"episode,omitempty"`
}

// SortableList allows sorting a list.
type SortableList []*Sortable

// State is partially filled out once for each app instance.
type State struct {
	// Shared
	Error    string        `json:"error"`
	Instance int           `json:"instance"`
	Missing  int64         `json:"missing,omitempty"`
	Size     int64         `json:"size"`
	Percent  float64       `json:"percent,omitempty"`
	Upcoming int64         `json:"upcoming,omitempty"`
	Next     SortableList  `json:"next,omitempty"`
	Latest   SortableList  `json:"latest,omitempty"`
	OnDisk   int64         `json:"onDisk,omitempty"`
	Elapsed  cnfg.Duration `json:"elapsed"` // How long it took.
	Name     string        `json:"name"`
	// Radarr
	Movies int64 `json:"movies,omitempty"`
	// Sonarr
	Shows    int64 `json:"shows,omitempty"`
	Episodes int64 `json:"episodes,omitempty"`
	// Readarr
	Authors  int   `json:"authors,omitempty"`
	Books    int64 `json:"books,omitempty"`
	Editions int   `json:"editions,omitempty"`
	// Lidarr
	Artists int   `json:"artists,omitempty"`
	Albums  int64 `json:"albums,omitempty"`
	Tracks  int64 `json:"tracks,omitempty"`
	// Downloader
	Downloads   int   `json:"downloads,omitempty"`
	Uploaded    int64 `json:"uploaded,omitempty"`
	Incomplete  int64 `json:"incomplete,omitempty"`
	Downloaded  int64 `json:"downloaded,omitempty"`
	Uploading   int64 `json:"uploading,omitempty"`
	Downloading int64 `json:"downloading,omitempty"`
	Seeding     int64 `json:"seeding,omitempty"`
	Paused      int64 `json:"paused,omitempty"`
	Errors      int64 `json:"errors,omitempty"`
	Month       int64 `json:"month,omitempty"`
	Week        int64 `json:"week,omitempty"`
}

// States is our compiled states for the dashboard.
type States struct {
	Lidarr  []*State       `json:"lidarr"`
	Radarr  []*State       `json:"radarr"`
	Readarr []*State       `json:"readarr"`
	Sonarr  []*State       `json:"sonarr"`
	Qbit    []*State       `json:"qbit"`
	Deluge  []*State       `json:"deluge"`
	SabNZB  []*State       `json:"sabnzbd"`
	Plex    *plex.Sessions `json:"plexSessions"`
}

// New configures the library.
func New(config *common.Config, plex *plexcron.Action) *Action {
	return &Action{
		cmd: &Cmd{
			Config:   config,
			PlexCron: plex,
		},
	}
}

// Create initializes the library.
func (a *Action) Create() {
	a.cmd.create()
}

func (c *Cmd) create() {
	var ticker *time.Ticker

	ci := c.ClientInfo

	if ci != nil && ci.Actions.Dashboard.Interval.Duration > 0 {
		ticker = time.NewTicker(ci.Actions.Dashboard.Interval.Duration)
		c.Printf("==> Dashboard State timer started, interval:%s, serial:%v",
			ci.Actions.Dashboard.Interval, c.Config.Serial)
	}

	c.Add(&common.Action{
		Name: TrigDashboard,
		Fn:   c.sendDashboardState,
		C:    make(chan website.EventType, 1),
		T:    ticker,
	})
}

// Send the current states for the dashboard to the website.
func (a *Action) Send(event website.EventType) {
	a.cmd.Exec(event, TrigDashboard)
}

func (c *Cmd) sendDashboardState(event website.EventType) {
	cmd := c.getStatesParallel
	if c.Serial {
		cmd = c.getStatesSerial
	}

	var (
		start  = time.Now()
		states = cmd()
		apps   = time.Since(start).Round(time.Millisecond)
	)

	c.SendData(&website.Request{
		Route:      website.DashRoute,
		Event:      event,
		LogPayload: true,
		LogMsg:     fmt.Sprintf("Dashboard State (elapsed: %v)", apps),
		Payload:    states,
	})
}

// getStatesSerial grabs data for each app serially.
func (c *Cmd) getStatesSerial() *States {
	sessions, _ := c.PlexCron.GetSessions(false)

	return &States{
		Deluge:  c.getDelugeStates(),
		Lidarr:  c.getLidarrStates(),
		Qbit:    c.getQbitStates(),
		Radarr:  c.getRadarrStates(),
		Readarr: c.getReadarrStates(),
		Sonarr:  c.getSonarrStates(),
		SabNZB:  c.getSabNZBStates(),
		Plex:    sessions,
	}
}

// getStatesParallel fires a routine for each app type and tries to get a lot of data fast!
func (c *Cmd) getStatesParallel() *States {
	states := &States{}

	var wg sync.WaitGroup

	wg.Add(8) //nolint:gomnd // we are polling 8 apps.

	go func() {
		defer c.CapturePanic()
		states.Deluge = c.getDelugeStates()
		wg.Done() //nolint:wsl
	}()
	go func() {
		defer c.CapturePanic()
		states.Lidarr = c.getLidarrStates()
		wg.Done() //nolint:wsl
	}()
	go func() {
		defer c.CapturePanic()
		states.Qbit = c.getQbitStates()
		wg.Done() //nolint:wsl
	}()
	go func() {
		defer c.CapturePanic()
		states.Radarr = c.getRadarrStates()
		wg.Done() //nolint:wsl
	}()
	go func() {
		defer c.CapturePanic()
		states.Readarr = c.getReadarrStates()
		wg.Done() //nolint:wsl
	}()
	go func() {
		defer c.CapturePanic()
		states.Sonarr = c.getSonarrStates()
		wg.Done() //nolint:wsl
	}()
	go func() {
		defer c.CapturePanic()
		states.SabNZB = c.getSabNZBStates()
		wg.Done() //nolint:wsl
	}()
	go func() {
		defer c.CapturePanic()
		states.Plex, _ = c.PlexCron.GetSessions(false)
		wg.Done() //nolint:wsl
	}()
	wg.Wait()

	return states
}

func (c *Cmd) getDelugeStates() []*State {
	states := []*State{}

	for instance, app := range c.Apps.Deluge {
		if app.Deluge.URL == "" {
			continue
		}

		c.Debugf("Getting Deluge State: %d:%s", instance+1, app.Deluge.URL)

		state, err := c.getDelugeState(instance+1, app)
		if err != nil {
			state.Error = err.Error()
			c.Errorf("Getting Deluge Data from %d:%s: %v", instance+1, app.Deluge.URL, err)
		}

		states = append(states, state)
	}

	return states
}

func (c *Cmd) getLidarrStates() []*State {
	states := []*State{}

	for instance, app := range c.Apps.Lidarr {
		if app.URL == "" {
			continue
		}

		c.Debugf("Getting Lidarr State: %d:%s", instance+1, app.URL)

		state, err := c.getLidarrState(instance+1, app)
		if err != nil {
			state.Error = err.Error()
			c.Errorf("Getting Lidarr Queue from %d:%s: %v", instance+1, app.URL, err)
		}

		states = append(states, state)
	}

	return states
}

func (c *Cmd) getRadarrStates() []*State {
	states := []*State{}

	for instance, app := range c.Apps.Radarr {
		if app.URL == "" {
			continue
		}

		c.Debugf("Getting Radarr State: %d:%s", instance+1, app.URL)

		state, err := c.getRadarrState(instance+1, app)
		if err != nil {
			state.Error = err.Error()
			c.Errorf("Getting Radarr Queue from %d:%s: %v", instance+1, app.URL, err)
		}

		states = append(states, state)
	}

	return states
}

func (c *Cmd) getReadarrStates() []*State {
	states := []*State{}

	for instance, app := range c.Apps.Readarr {
		if app.URL == "" {
			continue
		}

		c.Debugf("Getting Readarr State: %d:%s", instance+1, app.URL)

		state, err := c.getReadarrState(instance+1, app)
		if err != nil {
			state.Error = err.Error()
			c.Errorf("Getting Readarr Queue from %d:%s: %v", instance+1, app.URL, err)
		}

		states = append(states, state)
	}

	return states
}

func (c *Cmd) getQbitStates() []*State {
	states := []*State{}

	for instance, app := range c.Apps.Qbit {
		if app.URL == "" {
			continue
		}

		c.Debugf("Getting Qbit State: %d:%s", instance+1, app.URL)

		state, err := c.getQbitState(instance+1, app)
		if err != nil {
			state.Error = err.Error()
			c.Errorf("Getting Qbit Data from %d:%s: %v", instance+1, app.URL, err)
		}

		states = append(states, state)
	}

	return states
}

func (c *Cmd) getSonarrStates() []*State {
	states := []*State{}

	for instance, app := range c.Apps.Sonarr {
		if app.URL == "" {
			continue
		}

		c.Debugf("Getting Sonarr State: %d:%s", instance+1, app.URL)

		state, err := c.getSonarrState(instance+1, app)
		if err != nil {
			state.Error = err.Error()
			c.Errorf("Getting Sonarr Queue from %d:%s: %v", instance+1, app.URL, err)
		}

		states = append(states, state)
	}

	return states
}

func (c *Cmd) getDelugeState(instance int, app *apps.DelugeConfig) (*State, error) { //nolint:funlen,cyclop
	start := time.Now()
	size, xfers, err := app.GetXfersCompat()
	state := &State{
		Elapsed:  cnfg.Duration{Duration: time.Since(start)},
		Instance: instance,
		Name:     app.Name,
		Next:     []*Sortable{},
		Latest:   []*Sortable{},
	}

	exp.Apps.Add("Deluge&&GET Requests", 1)
	exp.Apps.Add("Deluge&&Bytes Received", size)

	if err != nil {
		exp.Apps.Add("Deluge&&GET Errors", 1)
		return state, fmt.Errorf("getting transfers from instance %d: %w", instance, err)
	}

	for _, xfer := range xfers {
		if eta, _ := xfer.Eta.Int64(); eta != 0 && xfer.FinishedTime == 0 {
			//			c.Error(xfer.FinishedTime, eta, xfer.Name)
			state.Next = append(state.Next, &Sortable{
				Name: xfer.Name,
				Date: time.Now().Add(time.Second * time.Duration(eta)),
			})
		} else if xfer.FinishedTime > 0 {
			seconds := time.Duration(xfer.FinishedTime) * time.Second
			state.Latest = append(state.Latest, &Sortable{
				Name: xfer.Name,
				Date: time.Now().Add(-seconds).Round(time.Second),
			})
		}

		state.Size += int64(xfer.TotalSize)
		state.Uploaded += int64(xfer.TotalUploaded)
		state.Downloaded += int64(xfer.AllTimeDownload)
		state.Downloads++

		if xfer.UploadPayloadRate > 0 {
			state.Uploading++
		}

		if xfer.DownloadPayloadRate > 0 {
			state.Downloading++
		}

		if !xfer.IsFinished {
			state.Incomplete++
		}

		if xfer.IsSeed {
			state.Seeding++
		}

		if xfer.Paused {
			state.Paused++
		}

		if xfer.Message != "OK" {
			state.Errors++
		}
	}

	sort.Sort(dateSorter(state.Next))
	sort.Sort(sort.Reverse(dateSorter(state.Latest)))
	state.Next.Shrink(showNext)
	state.Latest.Shrink(showLatest)

	return state, nil
}

func (c *Cmd) getLidarrState(instance int, app *apps.LidarrConfig) (*State, error) {
	state := &State{Instance: instance, Next: []*Sortable{}, Name: app.Name}
	start := time.Now()

	albums, err := app.GetAlbum("") // all albums
	state.Elapsed.Duration = time.Since(start)

	if err != nil {
		return state, fmt.Errorf("getting albums from instance %d: %w", instance, err)
	}

	artistIDs := make(map[int64]struct{})

	for _, album := range albums {
		have := false
		state.Albums++

		if album.Statistics != nil {
			artistIDs[album.ArtistID] = struct{}{}
			state.Percent += album.Statistics.PercentOfTracks
			state.Size += int64(album.Statistics.SizeOnDisk)
			state.Tracks += int64(album.Statistics.TotalTrackCount)
			state.Missing += int64(album.Statistics.TrackCount - album.Statistics.TrackFileCount)
			have = album.Statistics.TrackCount-album.Statistics.TrackFileCount < 1
			state.OnDisk += int64(album.Statistics.TrackFileCount)
		}

		if album.ReleaseDate.After(time.Now()) && album.Monitored && !have {
			state.Next = append(state.Next, &Sortable{
				id:   album.ID,
				Name: album.Title,
				Date: album.ReleaseDate,
				Sub:  album.Artist.ArtistName,
			})
		}
	}

	if state.Tracks > 0 {
		state.Percent /= float64(state.Tracks)
	} else {
		state.Percent = 100
	}

	state.Artists = len(artistIDs)
	sort.Sort(dateSorter(state.Next))
	state.Next.Shrink(showNext)

	if state.Latest, err = c.getLidarrHistory(app); err != nil {
		return state, fmt.Errorf("instance %d: %w", instance, err)
	}

	return state, nil
}

// getLidarrHistory is not done.
func (c *Cmd) getLidarrHistory(app *apps.LidarrConfig) ([]*Sortable, error) {
	history, err := app.GetHistoryPage(&starr.Req{
		Page:     1,
		PageSize: showLatest + 20, //nolint:gomnd // grab extra in case some are tracks and not albums.
		SortDir:  starr.SortDescend,
		SortKey:  "date",
		Filter:   lidarr.FilterTrackFileImported,
	})
	if err != nil {
		return nil, fmt.Errorf("getting history: %w", err)
	}

	table := []*Sortable{}
	albumIDs := make(map[int64]*struct{})

	for _, rec := range history.Records {
		if len(table) >= showLatest {
			break
		} else if albumIDs[rec.AlbumID] != nil {
			continue // we already have this album
		}

		albumIDs[rec.AlbumID] = &struct{}{}

		// An error here gets swallowed.
		if album, err := app.GetAlbumByID(rec.AlbumID); err == nil {
			table = append(table, &Sortable{
				Name: album.Title,
				Sub:  album.Artist.ArtistName,
				Date: rec.Date,
			})
		}
	}

	return table, nil
}

func (c *Cmd) getQbitState(instance int, app *apps.QbitConfig) (*State, error) { //nolint:cyclop,funlen
	start := time.Now()
	size, xfers, err := app.GetXfers()

	state := &State{
		Elapsed:  cnfg.Duration{Duration: time.Since(start)},
		Instance: instance,
		Name:     app.Name,
		Next:     []*Sortable{},
		Latest:   []*Sortable{},
	}

	exp.Apps.Add("Qbit&&GET Requests", 1)
	exp.Apps.Add("Qbit&&Bytes Received", size)

	if err != nil {
		exp.Apps.Add("Qbit&&GET Errors", 1)
		return state, fmt.Errorf("getting transfers from instance %d: %w", instance, err)
	}

	for _, xfer := range xfers {
		if xfer.Eta != 8640000 && xfer.Eta != 0 && xfer.AmountLeft > 0 {
			state.Next = append(state.Next, &Sortable{
				Name: xfer.Name,
				Date: time.Now().Add(time.Second * time.Duration(xfer.Eta)),
			})
		} else if xfer.AmountLeft == 0 {
			state.Latest = append(state.Latest, &Sortable{
				Name: xfer.Name,
				Date: time.Unix(int64(xfer.CompletionOn), 0).Round(time.Second),
			})
		}

		state.Size += xfer.Size
		state.Uploaded += xfer.Uploaded
		state.Downloaded += int64(xfer.Downloaded)
		state.Downloads++

		switch strings.ToLower(strings.TrimSpace(xfer.State)) {
		case "stalledup", "moving", "forcedup":
			state.Seeding++
		case "downloading", "forceddl":
			state.Downloading++
		case "uploading":
			state.Uploading++
		case "pausedup", "pauseddl":
			state.Paused++
		case "queuedup", "checkingup", "allocating", "metadl", "queueddl", "stalleddl", "checkingdl", "checkingresumedata":
			state.Incomplete++
		case "unknown", "missingfiles", "error":
			state.Errors++
		default:
			state.Errors++
		}
	}

	sort.Sort(dateSorter(state.Next))
	sort.Sort(sort.Reverse(dateSorter(state.Latest)))
	state.Next.Shrink(showNext)
	state.Latest.Shrink(showLatest)

	return state, nil
}

func (c *Cmd) getRadarrState(instance int, r *apps.RadarrConfig) (*State, error) {
	state := &State{Instance: instance, Next: []*Sortable{}, Latest: []*Sortable{}, Name: r.Name}
	start := time.Now()

	movies, err := r.GetMovie(0)
	state.Elapsed.Duration = time.Since(start)

	if err != nil {
		return state, fmt.Errorf("getting movies from instance %d: %w", instance, err)
	}

	processRadarrState(state, movies)
	sort.Sort(sort.Reverse(dateSorter(state.Latest)))
	sort.Sort(dateSorter(state.Next))
	state.Latest.Shrink(showLatest)
	state.Next.Shrink(showNext)

	return state, nil
}

func processRadarrState(state *State, movies []*radarr.Movie) { //nolint:cyclop
	for _, movie := range movies {
		state.Movies++
		state.Size += movie.SizeOnDisk

		if !movie.HasFile && movie.IsAvailable {
			state.Missing++
		}

		if !movie.HasFile && !movie.IsAvailable {
			state.Upcoming++
		}

		date := movie.DigitalRelease
		if date.IsZero() || movie.PhysicalRelease.After(time.Now()) {
			date = movie.PhysicalRelease
		}

		if date.After(time.Now()) && !movie.HasFile {
			state.Next = append(state.Next, &Sortable{Name: movie.Title, Date: date})
		}

		if movie.MovieFile != nil {
			state.Latest = append(state.Latest, &Sortable{Name: movie.Title, Date: movie.MovieFile.DateAdded})
			state.OnDisk++
		}
	}
}

func (c *Cmd) getReadarrState(instance int, app *apps.ReadarrConfig) (*State, error) {
	state := &State{Instance: instance, Next: []*Sortable{}, Name: app.Name}
	start := time.Now()

	books, err := app.GetBook("") // all books
	state.Elapsed.Duration = time.Since(start)

	if err != nil {
		return state, fmt.Errorf("getting books from instance %d: %w", instance, err)
	}

	authorIDs := make(map[int64]struct{})

	for _, book := range books {
		have := false
		state.Books++

		if book.Statistics != nil {
			authorIDs[book.AuthorID] = struct{}{}
			state.Percent += book.Statistics.PercentOfBooks
			state.Size += int64(book.Statistics.SizeOnDisk)
			state.Editions += book.Statistics.TotalBookCount
			state.Missing += int64(book.Statistics.BookCount - book.Statistics.BookFileCount)
			have = book.Statistics.BookCount-book.Statistics.BookFileCount < 1
			state.OnDisk += int64(book.Statistics.BookFileCount)
		}

		author := "unknown author"
		if book.Author != nil {
			author = book.Author.AuthorName
		}

		if book.ReleaseDate.After(time.Now()) && book.Monitored && !have {
			state.Next = append(state.Next, &Sortable{
				id:   book.ID,
				Name: book.Title,
				Date: book.ReleaseDate,
				Sub:  author,
			})
		}
	}

	if state.Editions > 0 {
		state.Percent /= float64(state.Editions)
	} else {
		state.Percent = 100
	}

	state.Authors = len(authorIDs)
	sort.Sort(dateSorter(state.Next))
	state.Next.Shrink(showNext)

	if state.Latest, err = c.getReadarrHistory(app); err != nil {
		return state, fmt.Errorf("instance %d: %w", instance, err)
	}

	return state, nil
}

// getReadarrHistory is not done.
func (c *Cmd) getReadarrHistory(app *apps.ReadarrConfig) ([]*Sortable, error) {
	history, err := app.GetHistoryPage(&starr.Req{
		Page:     1,
		PageSize: showLatest,
		SortDir:  starr.SortDescend,
		SortKey:  "date",
		Filter:   readarr.FilterBookFileImported,
	})
	if err != nil {
		return nil, fmt.Errorf("getting history: %w", err)
	}

	table := []*Sortable{}

	for idx := 0; idx < len(history.Records) && len(table) < showLatest; idx++ {
		// An error here gets swallowed.
		if book, err := app.GetBookByID(history.Records[idx].BookID); err == nil {
			table = append(table, &Sortable{
				Name: book.Title,
				Sub:  book.Author.AuthorName,
				Date: history.Records[idx].Date,
			})
		}
	}

	return table, nil
}

func (c *Cmd) getSonarrState(instance int, app *apps.SonarrConfig) (*State, error) {
	state := &State{Instance: instance, Next: []*Sortable{}, Name: app.Name}
	start := time.Now()

	allshows, err := app.GetAllSeries()
	state.Elapsed.Duration = time.Since(start)

	if err != nil {
		return state, fmt.Errorf("getting series from instance %d: %w", instance, err)
	}

	for _, show := range allshows {
		state.Shows++
		if show.Statistics != nil {
			state.Percent += show.Statistics.PercentOfEpisodes
			state.Size += show.Statistics.SizeOnDisk
			state.Episodes += int64(show.Statistics.TotalEpisodeCount)
			state.Missing += int64(show.Statistics.EpisodeCount - show.Statistics.EpisodeFileCount)
			state.OnDisk += int64(show.Statistics.EpisodeFileCount)
		}

		if show.NextAiring.After(time.Now()) {
			state.Next = append(state.Next, &Sortable{
				id:   show.ID,
				Name: show.Title,
				Date: show.NextAiring,
			})
		}
	}

	if state.Shows > 0 {
		state.Percent /= float64(state.Shows)
	} else {
		state.Percent = 100
	}

	if state.Next, err = c.getSonarrStateUpcoming(app, state.Next); err != nil {
		return state, fmt.Errorf("instance %d: %w", instance, err)
	}

	if state.Latest, err = c.getSonarrHistory(app); err != nil {
		return state, fmt.Errorf("instance %d: %w", instance, err)
	}

	return state, nil
}

func (c *Cmd) getSonarrHistory(app *apps.SonarrConfig) ([]*Sortable, error) {
	history, err := app.GetHistoryPage(&starr.Req{
		Page:     1,
		PageSize: showLatest + 5, //nolint:gomnd // grab extra in case there's an error.
		SortDir:  starr.SortDescend,
		SortKey:  "date",
		Filter:   sonarr.FilterDownloadFolderImported,
	})
	if err != nil {
		return nil, fmt.Errorf("getting history: %w", err)
	}

	table := []*Sortable{}

	for _, rec := range history.Records {
		if len(table) >= showLatest {
			break
		}

		series, err := app.GetSeriesByID(rec.SeriesID)
		if err != nil {
			continue
		}

		// An error here gets swallowed.
		if eps, err := app.GetSeriesEpisodes(rec.SeriesID); err == nil {
			for _, episode := range eps {
				if episode.ID == rec.EpisodeID {
					table = append(table, &Sortable{
						Name:    series.Title,
						Sub:     episode.Title,
						Date:    rec.Date,
						Season:  episode.SeasonNumber,
						Episode: episode.EpisodeNumber,
					})
				}
			}
		}
	}

	return table, nil
}

func (c *Cmd) getSonarrStateUpcoming(app *apps.SonarrConfig, next []*Sortable) ([]*Sortable, error) {
	sort.Sort(dateSorter(next))

	redo := []*Sortable{}

	for _, item := range next {
		eps, err := app.GetSeriesEpisodes(item.id)
		if err != nil {
			return nil, fmt.Errorf("getting series ID %d (%s): %w", item.id, item.Name, err)
		}

		for _, episode := range eps {
			if episode.AirDateUtc.Year() == item.Date.Year() &&
				episode.AirDateUtc.YearDay() == item.Date.YearDay() &&
				episode.SeasonNumber != 0 && episode.EpisodeNumber != 0 {
				redo = append(redo, &Sortable{
					Name:    item.Name,
					Sub:     episode.Title,
					Date:    episode.AirDateUtc,
					Season:  episode.SeasonNumber,
					Episode: episode.EpisodeNumber,
				})

				break
			}
		}

		if len(redo) >= showNext {
			break
		}
	}

	return redo, nil
}

func (c *Cmd) getSabNZBStates() []*State {
	states := []*State{}

	for instance, app := range c.Apps.SabNZB {
		if app.URL == "" {
			continue
		}

		c.Debugf("Getting SabNZB State: %d:%s", instance+1, app.URL)

		state, err := c.getSabNZBState(instance+1, app)
		if err != nil {
			state.Error = err.Error()
			c.Errorf("Getting SabNZB Data from %d:%s: %v", instance+1, app.URL, err)
		}

		states = append(states, state)
	}

	return states
}

func (c *Cmd) getSabNZBState(instance int, s *apps.SabNZBConfig) (*State, error) {
	state := &State{Instance: instance, Name: s.Name}
	start := time.Now()
	queue, err := s.GetQueue()
	hist, err2 := s.GetHistory()
	state.Elapsed.Duration = time.Since(start)

	if err != nil {
		return state, fmt.Errorf("getting queue from instance %d: %w", instance, err)
	} else if err2 != nil {
		return state, fmt.Errorf("getting history from instance %d: %w", instance, err2)
	}

	state.Size = hist.TotalSize.Bytes
	state.Month = hist.MonthSize.Bytes
	state.Week = hist.WeekSize.Bytes

	state.Downloads = len(queue.Slots) + hist.Noofslots
	state.Next = []*Sortable{}
	state.Latest = []*Sortable{}

	for _, xfer := range queue.Slots {
		if strings.EqualFold(xfer.Status, "Downloading") {
			state.Downloading++
		} else if strings.EqualFold(xfer.Status, "Paused") {
			state.Paused++
		}

		if xfer.Mbleft > 0 {
			state.Incomplete++
		}

		state.Next = append(state.Next, &Sortable{
			Date: xfer.Eta.Round(time.Second).UTC(),
			Name: xfer.Filename,
		})
	}

	for _, xfer := range hist.Slots {
		state.Latest = append(state.Latest, &Sortable{
			Name: xfer.Name,
			Date: time.Unix(xfer.Completed, 0).Round(time.Second).UTC(),
		})

		if xfer.FailMessage != "" {
			state.Errors++
		} else {
			state.Downloaded++
		}
	}

	sort.Sort(dateSorter(state.Next))
	sort.Sort(sort.Reverse(dateSorter(state.Latest)))
	state.Next.Shrink(showNext)
	state.Latest.Shrink(showLatest)

	return state, nil
}

type dateSorter []*Sortable

func (s dateSorter) Len() int {
	return len(s)
}

func (s dateSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s dateSorter) Less(i, j int) bool {
	return s[i].Date.Before(s[j].Date)
}

// Shrink a sortable list.
func (s *SortableList) Shrink(size int) {
	if s == nil {
		return
	}

	if len(*s) > size {
		*s = (*s)[:size]
	}
}
