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

const maxQueuePayloadSize = 50

// sendDownloadingQueues gathers the downloading queue items from cache and sends them.
func (c *cmd) sendDownloadingQueues(ctx context.Context, input *common.ActionInput) {
	lidarr := c.getDownloadingItemsLidarr(ctx)
	radarr := c.getDownloadingItemsRadarr(ctx)
	readarr := c.getDownloadingItemsReadarr(ctx)
	sonarr := c.getDownloadingItemsSonarr(ctx)

	if lidarr.Empty() && radarr.Empty() && readarr.Empty() && sonarr.Empty() {
		c.Debugf("[%s requested] No Downloading Items found; Lidarr: %d, Radarr: %d, Readarr: %d, Sonarr: %d",
			input.Type, lidarr.Len(), radarr.Len(), readarr.Len(), sonarr.Len())

		if c.empty {
			return
		}

		c.empty = true
	} else {
		c.empty = false
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

func (c *cmd) getDownloadingItemsLidarr(_ context.Context) itemList { //nolint:cyclop
	items := make(itemList)

	ci := clientinfo.Get()
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
		appList := listItem{}
		// repeatStomper is used to collapse duplicate download IDs.
		repeatStomper := make(map[string]*lidarr.QueueRecord)

		for _, item := range queue.Records {
			// Delay items have no download ID, so group them by size.
			if item.DownloadID == "" {
				item.DownloadID = fmt.Sprint(item.Size)
			}

			if s := strings.ToLower(item.Status); (s == downloading || s == delay) && repeatStomper[item.DownloadID] == nil {
				appList.Queue = append(appList.Queue, item)
				repeatStomper[item.DownloadID] = item
				appList.Name = app.Name
			}
		}

		appList.Total = len(appList.Queue)
		appList.Queue = truncateQueue(appList.Queue)
		items[instance] = appList
	}

	return items
}

func (c *cmd) getDownloadingItemsRadarr(_ context.Context) itemList { //nolint:cyclop
	items := make(itemList)

	ci := clientinfo.Get()
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
		appList := listItem{}
		// repeatStomper is used to collapse duplicate download IDs.
		repeatStomper := make(map[string]*radarr.QueueRecord)

		for _, item := range queue.Records {
			// Delay items have no download ID, so group them by size.
			if item.DownloadID == "" {
				item.DownloadID = fmt.Sprint(item.Size)
			}

			if s := strings.ToLower(item.Status); (s == downloading || s == delay) && repeatStomper[item.DownloadID] == nil {
				appList.Queue = append(appList.Queue, item)
				repeatStomper[item.DownloadID] = item
				appList.Name = app.Name
			}
		}

		appList.Total = len(appList.Queue)
		appList.Queue = truncateQueue(appList.Queue)
		items[instance] = appList
	}

	return items
}

func (c *cmd) getDownloadingItemsReadarr(_ context.Context) itemList { //nolint:cyclop
	items := make(itemList)

	ci := clientinfo.Get()
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
		appList := listItem{}
		// repeatStomper is used to collapse duplicate download IDs.
		repeatStomper := make(map[string]*readarr.QueueRecord)

		for _, item := range queue.Records {
			// Delay items have no download ID, so group them by size.
			if item.DownloadID == "" {
				item.DownloadID = fmt.Sprint(item.Size)
			}

			if s := strings.ToLower(item.Status); (s == downloading || s == delay) && repeatStomper[item.DownloadID] == nil {
				appList.Queue = append(appList.Queue, item)
				repeatStomper[item.DownloadID] = item
				appList.Name = app.Name
			}
		}

		appList.Total = len(appList.Queue)
		appList.Queue = truncateQueue(appList.Queue)
		items[instance] = appList
	}

	return items
}

func (c *cmd) getDownloadingItemsSonarr(_ context.Context) itemList { //nolint:cyclop
	items := make(itemList)

	ci := clientinfo.Get()
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
		appList := listItem{}
		// repeatStomper is used to collapse duplicate download IDs.
		repeatStomper := make(map[string]*sonarr.QueueRecord)

		for _, item := range queue.Records {
			// Delay items have no download ID, so group them by size.
			if item.DownloadID == "" {
				item.DownloadID = fmt.Sprint(item.Size)
			}

			if s := strings.ToLower(item.Status); (s == downloading || s == delay) && repeatStomper[item.DownloadID] == nil {
				appList.Queue = append(appList.Queue, item)
				repeatStomper[item.DownloadID] = item
				appList.Name = app.Name
			}
		}

		appList.Total = len(appList.Queue)
		appList.Queue = truncateQueue(appList.Queue)
		items[instance] = appList
	}

	return items
}

func truncateQueue(queue []any) []any {
	if len(queue) <= maxQueuePayloadSize {
		return queue
	}

	return queue[:maxQueuePayloadSize]
}
