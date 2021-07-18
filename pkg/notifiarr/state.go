package notifiarr

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"golift.io/starr/radarr"
)

/* This file sends state of affairs to notifiarr.com */
// That is, it collects library data and downloader data.

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

// State is partially filled out once for each app instance.
type State struct {
	// Shared
	Error    string      `json:"error"`
	Instance int         `json:"instance"`
	Missing  int64       `json:"missing,omitempty"`
	Size     int64       `json:"size"`
	Percent  float64     `json:"percent,omitempty"`
	Upcoming int64       `json:"upcoming,omitempty"`
	Next     []*Sortable `json:"next,omitempty"`
	Latest   []*Sortable `json:"latest,omitempty"`
	// Radarr
	Movies int64 `json:"movies,omitempty"`
	// Sonarr
	Shows    int64 `json:"shows,omitempty"`
	Episodes int64 `json:"episodes,omitempty"`
	// Readarr
	Authors int   `json:"authors,omitempty"`
	Books   int64 `json:"books,omitempty"`
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
}

type States struct {
	Lidarr  []*State `json:"lidarr"`
	Radarr  []*State `json:"radarr"`
	Readarr []*State `json:"readarr"`
	Sonarr  []*State `json:"sonarr"`
	Qbit    []*State `json:"qbit"`
	Deluge  []*State `json:"deluge"`
}

func (c *Config) GetState() {
	states := &States{
		Deluge:  c.getDelugeStates(),
		Lidarr:  c.getLidarrStates(),
		Qbit:    c.getQbitStates(),
		Radarr:  c.getRadarrStates(),
		Readarr: c.getReadarrStates(),
		Sonarr:  c.getSonarrStates(),
	}

	_, _, err := c.SendData(c.BaseURL+"/api/v1/user/dashboard", states, true)
	if err != nil {
		c.Errorf("Sending State Data: %v", err)
	}
}

func (c *Config) getDelugeStates() []*State {
	states := []*State{}

	for instance, d := range c.Apps.Deluge {
		if d.Deluge.URL == "" {
			continue
		}

		c.Debugf("Getting Deluge State: %d:%s", instance+1, d.Deluge.URL)

		state, err := c.getDelugeState(instance+1, d)
		if err != nil {
			state.Error = err.Error()
			c.Errorf("Getting Deluge Data from %d:%s: %v", instance+1, d.Deluge.URL, err)
		}

		states = append(states, state)
	}

	return states
}

func (c *Config) getLidarrStates() []*State {
	states := []*State{}

	for instance, r := range c.Apps.Lidarr {
		if r.URL == "" {
			continue
		}

		c.Debugf("Getting Lidarr State: %d:%s", instance+1, r.URL)

		state, err := c.getLidarrState(instance+1, r)
		if err != nil {
			state.Error = err.Error()
			c.Errorf("Getting Lidarr Queue from %d:%s: %v", instance+1, r.URL, err)
		}

		states = append(states, state)
	}

	return states
}

func (c *Config) getRadarrStates() []*State {
	states := []*State{}

	for instance, r := range c.Apps.Radarr {
		if r.URL == "" {
			continue
		}

		c.Debugf("Getting Radarr State: %d:%s", instance+1, r.URL)

		state, err := c.getRadarrState(instance+1, r)
		if err != nil {
			state.Error = err.Error()
			c.Errorf("Getting Radarr Queue from %d:%s: %v", instance+1, r.URL, err)
		}

		states = append(states, state)
	}

	return states
}

func (c *Config) getReadarrStates() []*State {
	states := []*State{}

	for instance, r := range c.Apps.Readarr {
		if r.URL == "" {
			continue
		}

		c.Debugf("Getting Readarr State: %d:%s", instance+1, r.URL)

		state, err := c.getReadarrState(instance+1, r)
		if err != nil {
			state.Error = err.Error()
			c.Errorf("Getting Readarr Queue from %d:%s: %v", instance+1, r.URL, err)
		}

		states = append(states, state)
	}

	return states
}

func (c *Config) getQbitStates() []*State {
	states := []*State{}

	for instance, q := range c.Apps.Qbit {
		if q.URL == "" {
			continue
		}

		c.Debugf("Getting Qbit State: %d:%s", instance+1, q.URL)

		state, err := c.getQbitState(instance+1, q)
		if err != nil {
			state.Error = err.Error()
			c.Errorf("Getting Qbit Data from %d:%s: %v", instance+1, q.URL, err)
		}

		states = append(states, state)
	}

	return states
}

func (c *Config) getSonarrStates() []*State {
	states := []*State{}

	for instance, s := range c.Apps.Sonarr {
		if s.URL == "" {
			continue
		}

		c.Debugf("Getting Sonarr State: %d:%s", instance+1, s.URL)

		state, err := c.getSonarrState(instance+1, s)
		if err != nil {
			state.Error = err.Error()
			c.Errorf("Getting Sonarr Queue from %d:%s: %v", instance+1, s.URL, err)
		}

		states = append(states, state)
	}

	return states
}

func (c *Config) getDelugeState(instance int, d *apps.DelugeConfig) (*State, error) {
	state := &State{Instance: instance}

	xfers, err := d.GetXfersCompat()
	if err != nil {
		return state, fmt.Errorf("getting transfers from instance %d: %w", instance, err)
	}

	for _, xfer := range xfers {
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

	return state, nil
}

func (c *Config) getLidarrState(instance int, l *apps.LidarrConfig) (*State, error) {
	state := &State{Instance: instance, Next: []*Sortable{}}

	albums, err := l.GetAlbum("") // all albums
	if err != nil {
		return state, fmt.Errorf("getting albums from instance %d: %w", instance, err)
	}

	artistIDs := make(map[int64]struct{})

	for _, album := range albums {
		state.Albums++

		if album.Statistics != nil {
			artistIDs[album.ArtistID] = struct{}{}
			state.Percent += album.Statistics.PercentOfTracks
			state.Size += int64(album.Statistics.SizeOnDisk)
			state.Tracks += int64(album.Statistics.TotalTrackCount)
			state.Missing += int64(album.Statistics.TrackCount - album.Statistics.TrackFileCount)
		}
	}

	state.Percent /= float64(state.Tracks)
	state.Artists = len(artistIDs)

	return state, nil
}

func (c *Config) getQbitState(instance int, q *apps.QbitConfig) (*State, error) {
	state := &State{Instance: instance}

	xfers, err := q.GetXfers()
	if err != nil {
		return state, fmt.Errorf("getting transfers from instance %d: %w", instance, err)
	}

	for _, xfer := range xfers {
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

	return state, nil
}

func (c *Config) getRadarrState(instance int, r *apps.RadarrConfig) (*State, error) {
	state := &State{Instance: instance, Next: []*Sortable{}, Latest: []*Sortable{}}

	movies, err := r.GetMovie(0)
	if err != nil {
		return state, fmt.Errorf("getting movies from instance %d: %w", instance, err)
	}

	processRadarrState(state, movies)

	return sortRadarrLists(state), nil
}

func processRadarrState(state *State, movies []*radarr.Movie) {
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
		}
	}
}

func sortRadarrLists(state *State) *State {
	// Ascending: dates closer to now() at top
	sort.Sort(sort.Reverse(dateSorter(state.Latest)))
	sort.Sort(dateSorter(state.Next))

	if len(state.Next) > showNext {
		state.Next = state.Next[:showNext]
	}

	if len(state.Latest) > showLatest {
		state.Latest = state.Latest[:showLatest]
	}

	return state
}

func (c *Config) getReadarrState(instance int, r *apps.ReadarrConfig) (*State, error) {
	state := &State{Instance: instance, Next: []*Sortable{}}

	books, err := r.GetBook("") // all books
	if err != nil {
		return state, fmt.Errorf("getting books from instance %d: %w", instance, err)
	}

	authorIDs := make(map[int64]struct{})

	for _, book := range books {
		state.Books++

		if book.Statistics != nil {
			authorIDs[book.AuthorID] = struct{}{}
			// state.Percent += book.Statistics.PercentOfBooks
			// state.Editions += book.Statistics.TotalBookCount
			state.Size += int64(book.Statistics.SizeOnDisk)
			state.Missing += int64(book.Statistics.BookCount - book.Statistics.BookFileCount)
		}
	}

	// state.Percent /= float64(state.Editions)
	state.Authors = len(authorIDs)

	return state, nil
}

func (c *Config) getSonarrState(instance int, r *apps.SonarrConfig) (*State, error) {
	state := &State{Instance: instance, Next: []*Sortable{}}

	allshows, err := r.GetAllSeries()
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
		}

		if show.NextAiring.After(time.Now()) {
			state.Next = append(state.Next, &Sortable{
				id:   show.ID,
				Name: show.Title,
				Date: show.NextAiring,
			})
		}
	}

	state.Percent /= float64(state.Shows)

	if state.Next, err = c.getSonarrStateUpcoming(r, state.Next); err != nil {
		return state, fmt.Errorf("instance %d: %w", instance, err)
	}

	return state, nil
}

func (c *Config) getSonarrStateUpcoming(r *apps.SonarrConfig, next []*Sortable) ([]*Sortable, error) {
	sort.Sort(dateSorter(next))

	redo := []*Sortable{}

	for _, item := range next {
		eps, err := r.GetSeriesEpisodes(item.id)
		if err != nil {
			return nil, fmt.Errorf("getting series ID %d (%s): %w", item.id, item.Name, err)
		}

		for _, ep := range eps {
			if ep.AirDateUtc.Year() == item.Date.Year() && ep.AirDateUtc.YearDay() == item.Date.YearDay() &&
				ep.SeasonNumber != 0 && ep.EpisodeNumber != 0 {
				redo = append(redo, &Sortable{
					Name:    item.Name,
					Sub:     ep.Title,
					Date:    ep.AirDateUtc,
					Season:  ep.SeasonNumber,
					Episode: ep.EpisodeNumber,
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
