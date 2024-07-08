package dashboard

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"golift.io/starr"
	"golift.io/starr/sonarr"
)

func (c *Cmd) getSonarrStates(ctx context.Context) []*State {
	states := []*State{}

	for instance, app := range c.Apps.Sonarr {
		if !app.Enabled() {
			continue
		}

		c.Debugf("Getting Sonarr State: %d:%s", instance+1, app.URL)

		state, err := c.getSonarrState(ctx, instance+1, app)
		if err != nil {
			state.Error = err.Error()
			c.Errorf("Getting Sonarr Queue from %d:%s: %v", instance+1, app.URL, err)
		}

		states = append(states, state)
	}

	return states
}

func (c *Cmd) getSonarrState(ctx context.Context, instance int, app *apps.SonarrConfig) (*State, error) {
	state := &State{Instance: instance, Next: []*Sortable{}, Name: app.Name}
	start := time.Now()

	allshows, err := app.GetAllSeriesContext(ctx)
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
	history, err := app.GetHistoryPage(&starr.PageReq{
		Page:     1,
		PageSize: showLatest + 5, //nolint:mnd // grab extra in case there's an error.
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
		if eps, err := app.GetSeriesEpisodes(&sonarr.GetEpisode{SeriesID: rec.SeriesID}); err == nil {
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
		eps, err := app.GetSeriesEpisodes(&sonarr.GetEpisode{SeriesID: item.id})
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
