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

// sendDownloadingQueues gathers the downloading queue items from cache and sends them.
func (c *cmd) sendDownloadingQueues(ctx context.Context, input *common.ActionInput) {
	lidarr := c.getDownloadingItemsLidrr(ctx)
	radarr := c.getDownloadingItemsRadarr(ctx)
	readarr := c.getDownloadingItemsReadarr(ctx)
	sonarr := c.getDownloadingItemsSonarr(ctx)

	if lidarr.Empty() && radarr.Empty() && readarr.Empty() && sonarr.Empty() {
		return
	}

	c.SendData(&website.Request{
		Route:      website.DownloadRoute,
		Event:      input.Type,
		LogPayload: true,
		ErrorsOnly: true,
		LogMsg: fmt.Sprintf("Downloading Items; Lidarr: %d, Radarr: %d, Readarr: %d, Sonarr: %d",
			lidarr.Len(), radarr.Len(), readarr.Len(), sonarr.Len()),
		Payload: &QueuesPaylod{
			Lidarr:  lidarr,
			Radarr:  radarr,
			Readarr: readarr,
			Sonarr:  sonarr,
		},
	})
}

func (c *cmd) getDownloadingItemsLidrr(_ context.Context) itemList {
	items := make(itemList)

	ci := website.GetClientInfo()
	if ci == nil {
		return items
	}

	for idx, app := range c.Apps.Lidarr {
		instance := idx + 1
		if !app.Enabled() || !ci.Actions.Apps.Lidarr.Finished(instance) {
			continue
		}

		cacheItem := data.GetWithID("lidarr", idx)
		if cacheItem == nil || cacheItem.Data == nil {
			continue
		}

		queue, _ := cacheItem.Data.(*lidarr.Queue)
		app := items[instance]

		for _, item := range queue.Records {
			if strings.ToLower(item.Status) == downloading {
				app.Queue = append(app.Queue, item)
			}
		}
	}

	return items
}

func (c *cmd) getDownloadingItemsRadarr(_ context.Context) itemList {
	items := make(itemList)

	ci := website.GetClientInfo()
	if ci == nil {
		return items
	}

	for idx, app := range c.Apps.Radarr {
		instance := idx + 1
		if !app.Enabled() || !ci.Actions.Apps.Radarr.Finished(instance) {
			continue
		}

		cacheItem := data.GetWithID("radarr", idx)
		if cacheItem == nil || cacheItem.Data == nil {
			continue
		}

		queue, _ := cacheItem.Data.(*radarr.Queue)
		app := items[instance]

		for _, item := range queue.Records {
			if strings.ToLower(item.Status) == downloading {
				app.Queue = append(app.Queue, item)
			}
		}
	}

	return items
}

func (c *cmd) getDownloadingItemsReadarr(_ context.Context) itemList {
	items := make(itemList)

	ci := website.GetClientInfo()
	if ci == nil {
		return items
	}

	for idx, app := range c.Apps.Readarr {
		instance := idx + 1
		if !app.Enabled() || !ci.Actions.Apps.Readarr.Finished(instance) {
			continue
		}

		cacheItem := data.GetWithID("readarr", idx)
		if cacheItem == nil || cacheItem.Data == nil {
			continue
		}

		queue, _ := cacheItem.Data.(*readarr.Queue)
		app := items[instance]

		for _, item := range queue.Records {
			if strings.ToLower(item.Status) == downloading {
				app.Queue = append(app.Queue, item)
			}
		}
	}

	return items
}

func (c *cmd) getDownloadingItemsSonarr(_ context.Context) itemList {
	items := make(itemList)

	ci := website.GetClientInfo()
	if ci == nil {
		return items
	}

	for idx, app := range c.Apps.Sonarr {
		instance := idx + 1
		if !app.Enabled() || !ci.Actions.Apps.Sonarr.Finished(instance) {
			continue
		}

		cacheItem := data.GetWithID("sonarr", idx)
		if cacheItem == nil || cacheItem.Data == nil {
			continue
		}

		queue, _ := cacheItem.Data.(*sonarr.Queue)
		app := items[instance]
		// repeatStomper is used to collapse duplicate download IDs.
		repeatStomper := make(map[string]*sonarr.QueueRecord)

		for _, item := range queue.Records {
			if strings.ToLower(item.Status) == downloading && repeatStomper[item.DownloadID] == nil {
				app.Queue = append(app.Queue, item)
				repeatStomper[item.DownloadID] = item
			}
		}
	}

	return items
}
