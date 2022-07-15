package dashboard

import (
	"fmt"
	"sort"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"golift.io/cnfg"
)

func (c *Cmd) getDelugeStates() []*State {
	states := []*State{}

	for instance, app := range c.Apps.Deluge {
		if app.Timeout.Duration < 0 || app.URL == "" {
			continue
		}

		c.Debugf("Getting Deluge State: %d:%s", instance+1, app.URL)

		state, err := c.getDelugeState(instance+1, app)
		if err != nil {
			state.Error = err.Error()
			c.Errorf("Getting Deluge Data from %d:%s: %v", instance+1, app.URL, err)
		}

		states = append(states, state)
	}

	return states
}

func (c *Cmd) getDelugeState(instance int, app *apps.DelugeConfig) (*State, error) { //nolint:funlen,cyclop
	start := time.Now()
	xfers, err := app.GetXfersCompat()
	state := &State{
		Elapsed:  cnfg.Duration{Duration: time.Since(start)},
		Instance: instance,
		Name:     app.Name,
		Next:     []*Sortable{},
		Latest:   []*Sortable{},
	}

	if err != nil {
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
