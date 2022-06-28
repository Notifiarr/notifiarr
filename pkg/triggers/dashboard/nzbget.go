package dashboard

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/exp"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"golift.io/nzbget"
)

func (c *Cmd) getNZBGetStates() []*State {
	states := []*State{}

	for instance, app := range c.Apps.NZBGet {
		if app.URL == "" {
			continue
		}

		c.Debugf("Getting NZBGet State: %d:%s", instance+1, app.URL)

		state, err := c.getNZBGetState(instance+1, app)
		if err != nil {
			state.Error = err.Error()
			c.Errorf("Getting NZBGet Data from %d:%s: %v", instance+1, app.URL, err)
		}

		states = append(states, state)
	}

	return states
}

func (c *Cmd) getNZBGetState(instance int, n *apps.NZBGetConfig) (*State, error) {
	state := &State{Instance: instance, Name: n.Name}
	start := time.Now()

	queue, stat, hist, err := getNzbData(instance, n)
	if err != nil {
		exp.Apps.Add("NZBGet&&GET Errors", 1)
		return state, err
	}

	state.Elapsed.Duration = time.Since(start)
	state.Size = stat.DownloadedSizeMB * mnd.Megabyte
	state.Downloads = len(queue) + len(hist)
	state.Next = []*Sortable{}
	state.Latest = []*Sortable{}

	for idx, xfer := range queue {
		if xfer.Status == nzbget.GroupDOWNLOADING {
			state.Downloading++
		} else if xfer.Status == nzbget.GroupPAUSED {
			state.Paused++
		}

		if xfer.RemainingSizeMB > 0 || xfer.RemainingFileCount > 0 {
			state.Incomplete++
		}

		state.Next = append(state.Next, &Sortable{
			Date: start.Add(time.Duration(idx) * time.Minute), // hacky, but a reasonable guess?
			Name: xfer.NZBName,
		})
	}

	for _, xfer := range hist {
		state.Latest = append(state.Latest, &Sortable{
			Name: xfer.Name,
			Date: xfer.HistoryTime.Time,
		})

		if strings.HasPrefix(xfer.Status, "SUCCESS") {
			state.Downloaded++
		} else {
			state.Errors++
		}
	}

	sort.Sort(dateSorter(state.Next))
	sort.Sort(sort.Reverse(dateSorter(state.Latest)))
	state.Next.Shrink(showNext)
	state.Latest.Shrink(showLatest)

	return state, nil
}

func getNzbData(instance int, n *apps.NZBGetConfig) ([]*nzbget.Group, *nzbget.Status, []*nzbget.History, error) {
	queue, size, err := n.ListGroups()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("getting file groups (queue) from instance %d: %w", instance, err)
	}

	exp.Apps.Add("NZBGet&&GET Requests", 1)
	exp.Apps.Add("NZBGet&&Bytes Received", int64(size))

	stat, size, err := n.Status()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("getting status from instance %d: %w", instance, err)
	}

	exp.Apps.Add("NZBGet&&GET Requests", 1)
	exp.Apps.Add("NZBGet&&Bytes Received", int64(size))

	hist, size, err := n.History(true)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("getting status from instance %d: %w", instance, err)
	}

	exp.Apps.Add("NZBGet&&GET Requests", 1)
	exp.Apps.Add("NZBGet&&Bytes Received", int64(size))

	return queue, stat, hist, nil
}
