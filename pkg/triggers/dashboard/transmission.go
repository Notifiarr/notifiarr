package dashboard

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	xmission "github.com/hekmon/transmissionrpc/v3"
	"golift.io/cnfg"
)

func (c *Cmd) getTransmissionStates(ctx context.Context) []*State {
	states := []*State{}

	for instance, app := range c.Apps.Transmission {
		if !app.Enabled() {
			continue
		}

		c.Debugf("Getting Transmission State: %d:%s", instance+1, app.URL)

		state, err := c.getTransmissionState(ctx, instance+1, app)
		if err != nil {
			state.Error = err.Error()
			c.Errorf("Getting Transmission Data from %d:%s: %v", instance+1, app.URL, err)
		}

		states = append(states, state)
	}

	return states
}

//nolint:cyclop,funlen
func (c *Cmd) getTransmissionState(ctx context.Context, instance int, app *apps.XmissionConfig) (*State, error) {
	start := time.Now()
	xfers, err := app.TorrentGetAll(ctx)

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
		if *xfer.ETA != 8640000 && *xfer.ETA != 0 && *xfer.LeftUntilDone > 0 {
			state.Next = append(state.Next, &Sortable{
				Name: *xfer.Name,
				Date: time.Now().Add(time.Second * time.Duration(*xfer.ETA)),
			})
		} else if *xfer.LeftUntilDone == 0 {
			state.Latest = append(state.Latest, &Sortable{
				Name: *xfer.Name,
				Date: xfer.DoneDate.Round(time.Second),
			})
		}

		state.Size += int64(xfer.TotalSize.Byte())
		state.Uploaded += *xfer.UploadedEver
		state.Downloaded += *xfer.DownloadedEver
		state.Downloads++

		if *xfer.PercentDone < 1 {
			state.Incomplete++
		}

		if *xfer.RateUpload > 0 {
			state.Uploading++
		}

		if xfer.ErrorString != nil && *xfer.ErrorString != "" {
			state.Errors++
		}

		switch *xfer.Status {
		case xmission.TorrentStatusSeed:
			state.Seeding++
		case xmission.TorrentStatusDownload:
			state.Downloading++
		case xmission.TorrentStatusStopped, xmission.TorrentStatusCheckWait,
			xmission.TorrentStatusIsolated, xmission.TorrentStatusDownloadWait,
			xmission.TorrentStatusCheck, xmission.TorrentStatusSeedWait:
			state.Paused++
		}
	}

	sort.Sort(dateSorter(state.Next))
	sort.Sort(sort.Reverse(dateSorter(state.Latest)))
	state.Next.Shrink(showNext)
	state.Latest.Shrink(showLatest)

	return state, nil
}
