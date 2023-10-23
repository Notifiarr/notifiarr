package dashboard

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"golift.io/starr/radarr"
)

func (c *Cmd) getRadarrStates(ctx context.Context) []*State {
	states := []*State{}

	for instance, app := range c.Apps.Radarr {
		if !app.Enabled() {
			continue
		}

		c.Debugf("Getting Radarr State: %d:%s", instance+1, app.URL)

		state, err := c.getRadarrState(ctx, instance+1, app)
		if err != nil {
			state.Error = err.Error()
			c.Errorf("Getting Radarr Queue from %d:%s: %v", instance+1, app.URL, err)
		}

		states = append(states, state)
	}

	return states
}

func (c *Cmd) getRadarrState(ctx context.Context, instance int, r *apps.RadarrConfig) (*State, error) {
	state := &State{Instance: instance, Next: []*Sortable{}, Latest: []*Sortable{}, Name: r.Name}
	start := time.Now()

	movies, err := r.GetMovieContext(ctx, &radarr.GetMovie{ExcludeLocalCovers: true})
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
