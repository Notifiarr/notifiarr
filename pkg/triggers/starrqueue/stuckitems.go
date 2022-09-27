package starrqueue

import (
	"context"
	"fmt"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/starr/lidarr"
	"golift.io/starr/radarr"
	"golift.io/starr/readarr"
	"golift.io/starr/sonarr"
)

// StuckItems sends the stuck queues items for all apps.
// Does not fetch fresh data first, uses cache.
func (a *Action) StuckItems(event website.EventType) {
	a.cmd.Exec(&common.ActionInput{Type: event}, TrigStuckItems)
}

// sendStuckQueues gathers the stuck queue from cache and sends them.
func (c *cmd) sendStuckQueues(ctx context.Context, input *common.ActionInput) {
	lidarr := c.getFinishedItemsLidarr(ctx)
	radarr := c.getFinishedItemsRadarr(ctx)
	readarr := c.getFinishedItemsReadarr(ctx)
	sonarr := c.getFinishedItemsSonarr(ctx)

	if lidarr.Empty() && radarr.Empty() && readarr.Empty() && sonarr.Empty() {
		c.Debugf("[%s requested] No stuck items found.", input.Type)
		return
	}

	c.SendData(&website.Request{
		Route:      website.StuckRoute,
		Event:      input.Type,
		LogPayload: true,
		LogMsg: fmt.Sprintf("Stuck Items; Lidarr: %d, Radarr: %d, Readarr: %d, Sonarr: %d",
			lidarr.Len(), radarr.Len(), readarr.Len(), sonarr.Len()),
		Payload: &QueuesPaylod{
			Lidarr:  lidarr,
			Radarr:  radarr,
			Readarr: readarr,
			Sonarr:  sonarr,
		},
	})
}

func (c *cmd) getFinishedItemsLidarr(_ context.Context) itemList { //nolint:cyclop
	stuck := make(itemList)

	for idx, app := range c.Apps.Lidarr {
		ci := website.GetClientInfo()
		if !app.Enabled() || ci == nil || !ci.Actions.Apps.Lidarr.Stuck(idx+1) {
			continue
		}

		item := data.GetWithID("lidarr", idx)
		if item == nil || item.Data == nil {
			continue
		}

		queue, _ := item.Data.(*lidarr.Queue)
		instance := idx + 1
		stuckapp := stuck[instance]

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			}

			stuckapp.Queue = append(stuckapp.Queue, item)
		}

		stuckapp.Name = c.Apps.Lidarr[idx].Name // this should be safe.
		stuck[instance] = stuckapp

		c.Debugf("Checking Lidarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			instance, len(queue.Records), len(stuck[instance].Queue))
	}

	return stuck
}

func (c *cmd) getFinishedItemsRadarr(_ context.Context) itemList { //nolint:cyclop
	stuck := make(itemList)

	for idx, app := range c.Apps.Radarr {
		ci := website.GetClientInfo()
		if !app.Enabled() || ci == nil || !ci.Actions.Apps.Radarr.Stuck(idx+1) {
			continue
		}

		item := data.GetWithID("radarr", idx)
		if item == nil || item.Data == nil {
			continue
		}

		queue, _ := item.Data.(*radarr.Queue)
		instance := idx + 1
		stuckapp := stuck[instance]

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			}

			stuckapp.Queue = append(stuckapp.Queue, item)
		}

		stuckapp.Name = c.Apps.Radarr[idx].Name // this should be safe.
		stuck[instance] = stuckapp

		c.Debugf("Checking Radarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			instance, len(queue.Records), len(stuck[instance].Queue))
	}

	return stuck
}

func (c *cmd) getFinishedItemsReadarr(_ context.Context) itemList { //nolint:cyclop
	stuck := make(itemList)

	for idx, app := range c.Apps.Readarr {
		ci := website.GetClientInfo()
		if !app.Enabled() || ci == nil || !ci.Actions.Apps.Readarr.Stuck(idx+1) {
			continue
		}

		item := data.GetWithID("readarr", idx)
		if item == nil || item.Data == nil {
			continue
		}

		queue, _ := item.Data.(*readarr.Queue)
		instance := idx + 1
		stuckapp := stuck[instance]

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			}

			stuckapp.Queue = append(stuckapp.Queue, item)
		}

		stuckapp.Name = c.Apps.Readarr[idx].Name // this should be safe.
		stuck[instance] = stuckapp

		c.Debugf("Checking Readarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			instance, len(queue.Records), len(stuck[instance].Queue))
	}

	return stuck
}

func (c *cmd) getFinishedItemsSonarr(_ context.Context) itemList { //nolint:cyclop
	stuck := make(itemList)

	for idx, app := range c.Apps.Sonarr {
		ci := website.GetClientInfo()
		if !app.Enabled() || ci == nil || !ci.Actions.Apps.Sonarr.Stuck(idx+1) {
			continue
		}

		cacheItem := data.GetWithID("sonarr", idx)
		if cacheItem == nil || cacheItem.Data == nil {
			continue
		}

		queue, _ := cacheItem.Data.(*sonarr.Queue)
		instance := idx + 1
		stuckapp := stuck[instance]
		// repeatStomper is used to collapse duplicate download IDs.
		repeatStomper := make(map[string]*sonarr.QueueRecord)

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			} else if repeatStomper[item.DownloadID] != nil {
				continue
			}

			repeatStomper[item.DownloadID] = item
			stuckapp.Queue = append(stuckapp.Queue, item)
		}

		stuckapp.Name = c.Apps.Sonarr[idx].Name // this should be safe.
		stuck[instance] = stuckapp

		c.Debugf("Checking Sonarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			instance, len(queue.Records), len(stuck[instance].Queue))
	}

	return stuck
}
