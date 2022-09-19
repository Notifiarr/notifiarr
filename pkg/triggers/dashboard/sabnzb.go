package dashboard

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
)

func (c *Cmd) getSabNZBStates(ctx context.Context) []*State {
	states := []*State{}

	for instance, app := range c.Apps.SabNZB {
		if !app.Enabled() {
			continue
		}

		c.Debugf("Getting SabNZB State: %d:%s", instance+1, app.URL)

		state, err := c.getSabNZBState(ctx, instance+1, app)
		if err != nil {
			state.Error = err.Error()
			c.Errorf("Getting SabNZB Data from %d:%s: %v", instance+1, app.URL, err)
		}

		states = append(states, state)
	}

	return states
}

func (c *Cmd) getSabNZBState(ctx context.Context, instance int, s *apps.SabNZBConfig) (*State, error) {
	state := &State{Instance: instance, Name: s.Name}
	start := time.Now()
	queue, err := s.GetQueue(ctx)
	hist, err2 := s.GetHistory(ctx)
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
