package dashboard

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"golift.io/cnfg"
)

func (c *Cmd) getQbitStates() []*State {
	states := []*State{}

	for instance, app := range c.Apps.Qbit {
		if app.Timeout.Duration < 0 || app.URL == "" {
			continue
		}

		c.Debugf("Getting Qbit State: %d:%s", instance+1, app.URL)

		state, err := c.getQbitState(instance+1, app)
		if err != nil {
			state.Error = err.Error()
			c.Errorf("Getting Qbit Data from %d:%s: %v", instance+1, app.URL, err)
		}

		states = append(states, state)
	}

	return states
}

func (c *Cmd) getQbitState(instance int, app *apps.QbitConfig) (*State, error) { //nolint:cyclop
	start := time.Now()
	xfers, err := app.GetXfers()

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
		if xfer.Eta != 8640000 && xfer.Eta != 0 && xfer.AmountLeft > 0 {
			state.Next = append(state.Next, &Sortable{
				Name: xfer.Name,
				Date: time.Now().Add(time.Second * time.Duration(xfer.Eta)),
			})
		} else if xfer.AmountLeft == 0 {
			state.Latest = append(state.Latest, &Sortable{
				Name: xfer.Name,
				Date: time.Unix(int64(xfer.CompletionOn), 0).Round(time.Second),
			})
		}

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

	sort.Sort(dateSorter(state.Next))
	sort.Sort(sort.Reverse(dateSorter(state.Latest)))
	state.Next.Shrink(showNext)
	state.Latest.Shrink(showLatest)

	return state, nil
}
