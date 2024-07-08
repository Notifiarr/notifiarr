package dashboard

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"golift.io/starr"
	"golift.io/starr/lidarr"
)

func (c *Cmd) getLidarrStates(ctx context.Context) []*State {
	states := []*State{}

	for instance, app := range c.Apps.Lidarr {
		if !app.Enabled() {
			continue
		}

		c.Debugf("Getting Lidarr State: %d:%s", instance+1, app.URL)

		state, err := c.getLidarrState(ctx, instance+1, app)
		if err != nil {
			state.Error = err.Error()
			c.Errorf("Getting Lidarr Queue from %d:%s: %v", instance+1, app.URL, err)
		}

		states = append(states, state)
	}

	return states
}

func (c *Cmd) getLidarrState(ctx context.Context, instance int, app *apps.LidarrConfig) (*State, error) {
	state := &State{Instance: instance, Next: []*Sortable{}, Name: app.Name}
	start := time.Now()

	albums, err := app.GetAlbumContext(ctx, "") // all albums
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

	if state.Latest, err = c.getLidarrHistory(ctx, app); err != nil {
		return state, fmt.Errorf("instance %d: %w", instance, err)
	}

	return state, nil
}

// getLidarrHistory is not done.
func (c *Cmd) getLidarrHistory(ctx context.Context, app *apps.LidarrConfig) ([]*Sortable, error) {
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
