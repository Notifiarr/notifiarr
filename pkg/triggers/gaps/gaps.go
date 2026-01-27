package gaps

import (
	"context"
	"fmt"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"golift.io/cnfg"
	"golift.io/starr/radarr"
)

/* Gaps allows filling gaps in Radarr collections. */

const TrigCollectionGaps common.TriggerName = "Sending Radarr Collection Gaps."

const randomMilliseconds = 5000

// Action contains the exported methods for this package.
type Action struct {
	cmd *cmd
}

type cmd struct {
	*common.Config
}

// New configures the library.
func New(config *common.Config) *Action {
	return &Action{cmd: &cmd{Config: config}}
}

// Create initializes the library.
func (a *Action) Create() {
	a.cmd.create()
}

func (c *cmd) create() {
	var dur time.Duration

	ci := clientinfo.Get()
	if ci != nil && ci.Actions.Gaps.Interval.Duration > 0 && len(ci.Actions.Gaps.Instances) > 0 {
		randomTime := time.Duration(c.Config.Rand().Intn(randomMilliseconds)) * time.Millisecond
		dur = ci.Actions.Gaps.Interval.Duration + randomTime
		mnd.Log.Printf("==> Collection Gaps Timer Enabled, interval:%s", ci.Actions.Gaps.Interval)
	}

	c.Add(&common.Action{
		Key:  "TrigCollectionGaps",
		Name: TrigCollectionGaps,
		Fn:   c.sendGaps,
		C:    make(chan *common.ActionInput, 1),
		D:    cnfg.Duration{Duration: dur},
	})
}

// Send radarr collection gaps to the website.
func (a *Action) Send(event website.EventType) {
	a.cmd.Exec(&common.ActionInput{Type: event}, TrigCollectionGaps)
}

func (c *cmd) sendGaps(ctx context.Context, input *common.ActionInput) {
	info := clientinfo.Get()
	if info == nil || len(info.Actions.Gaps.Instances) == 0 || len(c.Apps.Radarr) == 0 {
		mnd.Log.Errorf("[%s requested] Cannot send Radarr Collection Gaps: instances or configured Radarrs (%d) are zero.",
			input.Type, len(c.Apps.Radarr))
		return
	}

	for idx, app := range c.Apps.Radarr {
		instance := idx + 1
		if !app.Enabled() || !info.Actions.Gaps.Instances.Has(instance) {
			continue
		}

		type radarrGapsPayload struct {
			Instance int             `json:"instance"`
			Name     string          `json:"name"`
			Movies   []*radarr.Movie `json:"movies"`
		}

		movies, err := app.GetMovieContext(ctx, &radarr.GetMovie{ExcludeLocalCovers: true})
		if err != nil {
			mnd.Log.Errorf("[%s requested] Radarr Collection Gaps (%d:%s) failed: getting movies: %v",
				input.Type, instance, app.URL, err)
			continue
		}

		// Filter to only movies with collections and strip unnecessary data.
		// This dramatically reduces payload size while maintaining backward compatibility.
		collectionMovies := filterAndStripMoviesForGaps(movies)

		website.SendData(&website.Request{
			Route:      website.GapsRoute,
			Event:      input.Type,
			LogPayload: true,
			LogMsg:     fmt.Sprintf("Radarr Collection Gaps (%d:%s)", instance, app.URL),
			Payload:    &radarrGapsPayload{Movies: collectionMovies, Name: app.Name, Instance: instance},
		})
	}
}

// filterAndStripMoviesForGaps filters movies to only those with collections
// and strips unnecessary fields to minimize payload size.
// This maintains the same []*radarr.Movie type for backward compatibility.
func filterAndStripMoviesForGaps(movies []*radarr.Movie) []*radarr.Movie {
	result := make([]*radarr.Movie, 0, len(movies)/4) // estimate ~25% have collections

	for _, movie := range movies {
		// Skip movies without collections - they're irrelevant for gap detection.
		if movie.Collection == nil || movie.Collection.TmdbID == 0 {
			continue
		}

		// Create a minimal copy with only the fields needed for gap detection.
		// This keeps the same type but dramatically reduces JSON size.
		minimal := &radarr.Movie{
			ID:        movie.ID,
			Title:     movie.Title,
			TmdbID:    movie.TmdbID,
			Year:      movie.Year,
			Monitored: movie.Monitored,
			HasFile:   movie.HasFile,
			Collection: &radarr.Collection{
				Name:   movie.Collection.Name,
				TmdbID: movie.Collection.TmdbID,
				// Omit Images - not needed for gap detection
			},
			// Include quality profile so server knows what profile to use for adding missing movies
			QualityProfileID: movie.QualityProfileID,
			// Include tags for filtering/organization
			Tags: movie.Tags,
			// Include path for root folder detection
			Path: movie.Path,
		}

		result = append(result, minimal)
	}

	return result
}
