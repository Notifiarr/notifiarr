package dashboard

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/mrobinsn/go-rtorrent/rtorrent"
)

var ErrInvalidResponse = errors.New("invalid response")

func (c *Cmd) getRtorrentStates() []*State {
	states := []*State{}

	for instance, app := range c.Apps.Rtorrent {
		if !app.Enabled() {
			continue
		}

		c.Debugf("Getting rTorrent State: %d:%s", instance+1, app.URL)

		state, err := c.getRtorrentState(instance+1, app)
		if err != nil {
			state.Error = err.Error()
			c.Errorf("Getting rTorrent Data from %d:%s: %v", instance+1, app.URL, err)
		}

		states = append(states, state)
	}

	return states
}

func (c *Cmd) getRtorrentState(instance int, rTorrent *apps.RtorrentConfig) (*State, error) { //nolint:cyclop
	state := &State{Instance: instance, Name: rTorrent.Name}
	start := time.Now()

	data, err := getRtorrentData(rTorrent)
	if err != nil {
		return state, err
	}

	state.Elapsed.Duration = time.Since(start)
	state.Size = int64(data.DownTotal)
	state.Uploaded = int64(data.UpTotal)
	state.Downloads = len(data.Torrents)
	state.Next = []*Sortable{}
	state.Latest = []*Sortable{}

	for _, xfer := range data.Torrents {
		if !xfer.Active {
			state.Paused++
		} else if xfer.Active && xfer.Completed >= xfer.Size {
			state.Seeding++
		}

		if xfer.UpRate > 0 {
			state.Uploading++
		}

		if xfer.Message != "" {
			state.Errors++
		}

		if xfer.DownRate > 0 {
			state.Downloading++
		}

		if xfer.Completed < xfer.Size {
			state.Incomplete++
		}

		if !xfer.Finished.IsZero() {
			state.Downloaded++
			state.Latest = append(state.Latest, &Sortable{
				Name: xfer.Name,
				Date: xfer.Finished,
			})
		} else if !xfer.Started.IsZero() {
			state.Next = append(state.Next, &Sortable{
				Date: xfer.Started,
				Name: xfer.Name,
			})
		}
	}

	sort.Sort(sort.Reverse(dateSorter(state.Next)))
	sort.Sort(sort.Reverse(dateSorter(state.Latest)))
	state.Next.Shrink(showNext)
	state.Latest.Shrink(showLatest)

	return state, nil
}

type rTorrentData struct {
	DownTotal int
	UpTotal   int
	Torrents  []*RtorrentTorrent
}

func getRtorrentData(rTorrent *apps.RtorrentConfig) (*rTorrentData, error) {
	var (
		err    error
		output = &rTorrentData{}
	)

	output.Torrents, err = rTorrentTorrents(rTorrent)
	if err != nil {
		return nil, err
	}

	output.DownTotal, err = rTorrentDownTotal(rTorrent)
	if err != nil {
		return nil, err
	}

	output.UpTotal, err = rTorrentUpTotal(rTorrent)
	if err != nil {
		return nil, err
	}

	return output, nil
}

type RtorrentTorrent struct {
	Name      string
	Message   string
	Active    bool // inactive/active
	Size      int  // Total Size in bytes
	UpRate    int
	DownRate  int
	Completed int // Bytes Completed.
	Started   time.Time
	Finished  time.Time
}

func rTorrentTorrents(rTorrent *apps.RtorrentConfig) ([]*RtorrentTorrent, error) {
	args := []interface{}{
		"",
		string(rtorrent.ViewMain),
		rtorrent.DName.Query(),
		"d.message=",
		rtorrent.DIsActive.Query(),
		rtorrent.DSizeInBytes.Query(),
		rtorrent.DUpRate.Query(),
		rtorrent.DDownRate.Query(),
		rtorrent.DCompletedBytes.Query(),
		rtorrent.DFinishedTime.Query(),
		rtorrent.DStartedTime.Query(),
	}

	results, err := rTorrent.Call("d.multicall2", args...)
	if err != nil {
		return nil, fmt.Errorf("%w: d.multicall2 XMLRPC call failed", err)
	}

	var torrents []*RtorrentTorrent

	resInt, _ := results.([]interface{})
	for _, outerResult := range resInt {
		resOut, _ := outerResult.([]interface{})
		for _, innerResult := range resOut {
			torrentData, ok := innerResult.([]interface{})
			if !ok {
				return nil, fmt.Errorf("%w: data returned from query is unusable", ErrInvalidResponse)
			}

			//nolint:forcetypeassert // if these are bad it crashes here. :(
			torrents = append(torrents, &RtorrentTorrent{
				Name:      torrentData[0].(string),
				Message:   torrentData[1].(string),
				Active:    torrentData[2].(int) > 0,
				Size:      torrentData[3].(int),
				UpRate:    torrentData[4].(int),
				DownRate:  torrentData[5].(int),
				Completed: torrentData[6].(int),
				Finished:  time.Unix(int64(torrentData[7].(int)), 0),
				Started:   time.Unix(int64(torrentData[8].(int)), 0),
			})
		}
	}

	return torrents, nil
}

// rTorrentDownTotal returns the total downloaded metric reported by this RTorrent instance (bytes).
func rTorrentDownTotal(rTorrent *apps.RtorrentConfig) (int, error) {
	result, err := rTorrent.Call("throttle.global_down.total")
	if err != nil {
		return 0, fmt.Errorf("%w: throttle.global_down.total XMLRPC call failed", err)
	}

	if totals, ok := result.([]interface{}); ok {
		result = totals[0]
	}

	if total, ok := result.(int); ok {
		return total, nil
	}

	return 0, fmt.Errorf("%w: result isn't integer: %s", ErrInvalidResponse, result)
}

// rTorrentUpTotal returns the total uploaded metric reported by this RTorrent instance (bytes).
func rTorrentUpTotal(rTorrent *apps.RtorrentConfig) (int, error) {
	result, err := rTorrent.Call("throttle.global_up.total")
	if err != nil {
		return 0, fmt.Errorf("%w: throttle.global_up.total XMLRPC call failed", err)
	}

	if totals, ok := result.([]interface{}); ok {
		result = totals[0]
	}

	if total, ok := result.(int); ok {
		return total, nil
	}

	return 0, fmt.Errorf("%w: result isn't integer: %s", ErrInvalidResponse, result)
}
