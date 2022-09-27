package dashboard

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"golift.io/starr"
	"golift.io/starr/readarr"
)

func (c *Cmd) getReadarrStates(ctx context.Context) []*State {
	states := []*State{}

	for instance, app := range c.Apps.Readarr {
		if !app.Enabled() {
			continue
		}

		c.Debugf("Getting Readarr State: %d:%s", instance+1, app.URL)

		state, err := c.getReadarrState(ctx, instance+1, app)
		if err != nil {
			state.Error = err.Error()
			c.Errorf("Getting Readarr Queue from %d:%s: %v", instance+1, app.URL, err)
		}

		states = append(states, state)
	}

	return states
}

func (c *Cmd) getReadarrState(ctx context.Context, instance int, app *apps.ReadarrConfig) (*State, error) {
	state := &State{Instance: instance, Next: []*Sortable{}, Name: app.Name}
	start := time.Now()

	books, err := app.GetBookContext(ctx, "") // all books
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

		if book.ReleaseDate.After(time.Now()) && book.Monitored && !have {
			state.Next = append(state.Next, &Sortable{
				id:   book.ID,
				Name: book.Title,
				Date: book.ReleaseDate,
				Sub:  book.AuthorTitle,
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

	if state.Latest, err = c.getReadarrHistory(ctx, app); err != nil {
		return state, fmt.Errorf("instance %d: %w", instance, err)
	}

	return state, nil
}

// getReadarrHistory is not done.
func (c *Cmd) getReadarrHistory(ctx context.Context, app *apps.ReadarrConfig) ([]*Sortable, error) {
	history, err := app.GetHistoryPageContext(ctx, &starr.PageReq{
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
		if book, err := app.GetBookByIDContext(ctx, history.Records[idx].BookID); err == nil {
			table = append(table, &Sortable{
				Name: book.Title,
				Sub:  book.AuthorTitle,
				Date: history.Records[idx].Date,
			})
		}
	}

	return table, nil
}
