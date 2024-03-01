package starrqueue

import (
	"context"
	"fmt"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
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
		ci := clientinfo.Get()
		if !app.Enabled() || ci == nil || !ci.Actions.Apps.Lidarr.Stuck(idx+1) {
			continue
		}

		item := data.GetWithID("lidarr", idx)
		if item == nil || item.Data == nil {
			continue
		}

		queue, _ := item.Data.(*lidarr.Queue)
		instance := idx + 1
		appqueue := []*lidarrRecord{}
		// repeatStomper is used to collapse duplicate download IDs.
		repeatStomper := make(map[string]struct{})

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			} else if _, exists := repeatStomper[item.DownloadID]; exists {
				continue
			}

			repeatStomper[item.DownloadID] = struct{}{}
			appqueue = append(appqueue, &lidarrRecord{QueueRecord: item}) //nolint:wsl
		}

		stuck[instance] = listItem{Name: app.Name, Queue: appqueue, Total: queue.TotalRecords}
		c.Debugf("Checking Lidarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			instance, len(queue.Records), len(appqueue))
	}

	return stuck
}

func (c *cmd) getFinishedItemsRadarr(_ context.Context) itemList { //nolint:cyclop
	stuck := make(itemList)

	for idx, app := range c.Apps.Radarr {
		ci := clientinfo.Get()
		if !app.Enabled() || ci == nil || !ci.Actions.Apps.Radarr.Stuck(idx+1) {
			continue
		}

		item := data.GetWithID("radarr", idx)
		if item == nil || item.Data == nil {
			continue
		}

		queue, _ := item.Data.(*radarr.Queue)
		instance := idx + 1
		appqueue := []*radarrRecord{}
		// repeatStomper is used to collapse duplicate download IDs.
		repeatStomper := make(map[string]struct{})

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			} else if _, exists := repeatStomper[item.DownloadID]; exists {
				continue
			}

			repeatStomper[item.DownloadID] = struct{}{}
			appqueue = append(appqueue, &radarrRecord{QueueRecord: item}) //nolint:wsl
		}

		stuck[instance] = listItem{Name: app.Name, Queue: appqueue, Total: queue.TotalRecords}
		c.Debugf("Checking Radarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			instance, len(queue.Records), len(appqueue))
	}

	return stuck
}

func (c *cmd) getFinishedItemsReadarr(_ context.Context) itemList { //nolint:cyclop
	stuck := make(itemList)

	for idx, app := range c.Apps.Readarr {
		ci := clientinfo.Get()
		if !app.Enabled() || ci == nil || !ci.Actions.Apps.Readarr.Stuck(idx+1) {
			continue
		}

		item := data.GetWithID("readarr", idx)
		if item == nil || item.Data == nil {
			continue
		}

		queue, _ := item.Data.(*readarr.Queue)
		instance := idx + 1
		appqueue := []*readarrRecord{}
		// repeatStomper is used to collapse duplicate download IDs.
		repeatStomper := make(map[string]struct{})

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			} else if _, exists := repeatStomper[item.DownloadID]; exists {
				continue
			}

			repeatStomper[item.DownloadID] = struct{}{}
			appqueue = append(appqueue, &readarrRecord{QueueRecord: item}) //nolint:wsl
		}

		stuck[instance] = listItem{Name: app.Name, Queue: appqueue, Total: queue.TotalRecords}
		c.Debugf("Checking Readarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			instance, len(queue.Records), len(appqueue))
	}

	return stuck
}

func (c *cmd) getFinishedItemsSonarr(_ context.Context) itemList { //nolint:cyclop
	stuck := make(itemList)

	for idx, app := range c.Apps.Sonarr {
		ci := clientinfo.Get()
		if !app.Enabled() || ci == nil || !ci.Actions.Apps.Sonarr.Stuck(idx+1) {
			continue
		}

		cacheItem := data.GetWithID("sonarr", idx)
		if cacheItem == nil || cacheItem.Data == nil {
			continue
		}

		queue, _ := cacheItem.Data.(*sonarr.Queue)
		instance := idx + 1
		appqueue := []*sonarrRecord{}
		// repeatStomper is used to collapse duplicate download IDs.
		repeatStomper := make(map[string]struct{})

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			} else if _, exists := repeatStomper[item.DownloadID]; exists {
				continue
			}

			repeatStomper[item.DownloadID] = struct{}{}
			appqueue = append(appqueue, &sonarrRecord{QueueRecord: item}) //nolint:wsl
		}

		stuck[instance] = listItem{Name: app.Name, Queue: appqueue, Total: queue.TotalRecords}
		c.Debugf("Checking Sonarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			instance, len(queue.Records), len(appqueue))
	}

	return stuck
}
