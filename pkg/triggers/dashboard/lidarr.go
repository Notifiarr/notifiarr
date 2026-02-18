package dashboard

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"sort"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"golift.io/starr"
	"golift.io/starr/lidarr"
)

const lidarrAlbumPageSize = 500

func (c *Cmd) getLidarrStates(ctx context.Context) []*State {
	states := []*State{}

	for instance, app := range c.Apps.Lidarr {
		if !app.Enabled() || !c.Enabled.Lidarr {
			continue
		}

		mnd.Log.Debugf(mnd.GetID(ctx), "Getting Lidarr State: %d:%s", instance+1, app.URL)

		state, err := c.getLidarrState(ctx, instance+1, &app)
		if err != nil {
			state.Error = err.Error()
			mnd.Log.Errorf(mnd.GetID(ctx), "Getting Lidarr Queue from %d:%s: %v", instance+1, app.URL, err)
		}

		states = append(states, state)
	}

	return states
}

func (c *Cmd) getLidarrState(ctx context.Context, instance int, app *apps.Lidarr) (*State, error) {
	state := &State{Instance: instance, Next: []*Sortable{}, Name: app.Name}
	start := time.Now()
	now := time.Now()
	artistIDs := make(map[int64]struct{})

	for page := 1; ; page++ {
		albums, totalRecords, err := c.getLidarrAlbumPageContext(ctx, app, page, lidarrAlbumPageSize)
		if err != nil {
			if page != 1 {
				return state, fmt.Errorf("getting albums page %d from instance %d: %w", page, instance, err)
			}

			// Older Lidarr versions may not support paginated album responses.
			albums, err = app.GetAlbumContext(ctx, "") // all albums
			if err != nil {
				return state, fmt.Errorf("getting albums from instance %d: %w", instance, err)
			}

			accumulateLidarrAlbumStats(state, artistIDs, albums, now)
			break
		}

		if len(albums) == 0 {
			break
		}

		accumulateLidarrAlbumStats(state, artistIDs, albums, now)

		if len(albums) < lidarrAlbumPageSize || (totalRecords > 0 && page*lidarrAlbumPageSize >= totalRecords) {
			break
		}
	}

	finalizeLidarrAlbumStats(state, artistIDs)
	state.Elapsed.Duration = time.Since(start)
	sort.Sort(dateSorter(state.Next))
	state.Next.Shrink(showNext)

	var err error
	if state.Latest, err = c.getLidarrHistory(ctx, app); err != nil {
		return state, fmt.Errorf("instance %d: %w", instance, err)
	}

	return state, nil
}

// getLidarrHistory is not done.
func (c *Cmd) getLidarrHistory(ctx context.Context, app *apps.Lidarr) ([]*Sortable, error) {
	history, err := app.GetHistoryPageContext(ctx, &starr.PageReq{
		Page:     1,
		PageSize: showLatest + 20, //nolint:mnd // grab extra in case some are tracks and not albums.
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
		if album, err := app.GetAlbumByIDContext(ctx, rec.AlbumID); err == nil {
			table = append(table, &Sortable{
				Name: album.Title,
				Sub:  album.Artist.ArtistName,
				Date: rec.Date,
			})
		}
	}

	return table, nil
}

type lidarrAlbumPage struct {
	TotalRecords int              `json:"totalRecords"`
	Records      []*lidarr.Album  `json:"records"`
}

func (c *Cmd) getLidarrAlbumPageContext( //nolint:unparam
	ctx context.Context,
	app *apps.Lidarr,
	page, pageSize int,
) ([]*lidarr.Album, int, error) {
	params := make(url.Values)
	params.Set("page", starr.Str(page))
	params.Set("pageSize", starr.Str(pageSize))
	params.Set("sortKey", "id")
	params.Set("sortDirection", string(starr.SortAscend))

	resp := &lidarrAlbumPage{}
	err := app.GetInto(ctx, starr.Request{
		URI:   path.Join(lidarr.APIver, "album"),
		Query: params,
	}, resp)

	return resp.Records, resp.TotalRecords, err
}

func applyLidarrAlbumStats(state *State, albums []*lidarr.Album, now func() time.Time) {
	artistIDs := make(map[int64]struct{})

	accumulateLidarrAlbumStats(state, artistIDs, albums, now())
	finalizeLidarrAlbumStats(state, artistIDs)
}

func accumulateLidarrAlbumStats(state *State, artistIDs map[int64]struct{}, albums []*lidarr.Album, now time.Time) {
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

		if album.ReleaseDate.After(now) && album.Monitored && !have {
			state.Next = append(state.Next, &Sortable{
				id:   album.ID,
				Name: album.Title,
				Date: album.ReleaseDate,
				Sub:  album.Artist.ArtistName,
			})
		}
	}
}

func finalizeLidarrAlbumStats(state *State, artistIDs map[int64]struct{}) {
	if state.Tracks > 0 {
		state.Percent /= float64(state.Tracks)
	} else {
		state.Percent = 100
	}

	state.Artists = len(artistIDs)
}
