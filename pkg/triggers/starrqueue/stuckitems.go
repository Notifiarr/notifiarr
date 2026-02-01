package starrqueue

import (
	"context"
	"fmt"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"golift.io/starr"
	"golift.io/starr/lidarr"
	"golift.io/starr/radarr"
	"golift.io/starr/readarr"
	"golift.io/starr/sonarr"
)

// StuckItems sends the stuck queues items for all apps.
// Does not fetch fresh data first, uses cache.
func (a *Action) StuckItems(input *common.ActionInput) {
	id := logs.Log.Trace("", "start: Action.StuckItems", input.Type)
	defer logs.Log.Trace(id, "end: Action.StuckItems", input.Type)

	a.cmd.Exec(input, TrigStuckItems)
}

// sendStuckQueues gathers the stuck queue from cache and sends them.
func (c *cmd) sendStuckQueues(ctx context.Context, input *common.ActionInput) {
	id := logs.Log.Trace("", "start: sendStuckQueues", input.Type)
	defer logs.Log.Trace(id, "end: sendStuckQueues", input.Type)

	lidarr := c.getFinishedItemsLidarr(ctx)
	radarr := c.getFinishedItemsRadarr(ctx)
	readarr := c.getFinishedItemsReadarr(ctx)
	sonarr := c.getFinishedItemsSonarr(ctx)

	if lidarr.Empty() && radarr.Empty() && readarr.Empty() && sonarr.Empty() {
		mnd.Log.Debugf("[%s requested] No stuck items found.", input.Type)
		return
	}

	website.SendData(&website.Request{
		ReqID:      mnd.GetID(ctx),
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
	id := logs.Log.Trace("", "start: getFinishedItemsLidarr")
	defer logs.Log.Trace(id, "end: getFinishedItemsLidarr")

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
		// Pre-allocate with capacity to reduce allocations during append.
		appqueue := make([]*lidarrRecord, 0, len(queue.Records))
		// repeatStomper is used to collapse duplicate download IDs.
		repeatStomper := make(map[string]struct{}, len(queue.Records))

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			} else if _, exists := repeatStomper[item.DownloadID]; exists {
				continue
			}

			repeatStomper[item.DownloadID] = struct{}{}
			// Create minimal copy with only fields needed for stuck item detection.
			appqueue = append(appqueue, &lidarrRecord{QueueRecord: minimalLidarrRecord(item)}) //nolint:wsl
		}

		stuck[instance] = listItem{Name: app.Name, Queue: appqueue, Total: len(appqueue)}
		mnd.Log.Debugf("Checking Lidarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			instance, len(queue.Records), len(appqueue))
	}

	return stuck
}

func (c *cmd) getFinishedItemsRadarr(_ context.Context) itemList { //nolint:cyclop
	id := logs.Log.Trace("", "start: getFinishedItemsRadarr")
	defer logs.Log.Trace(id, "end: getFinishedItemsRadarr")

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
		// Pre-allocate with capacity to reduce allocations during append.
		appqueue := make([]*radarrRecord, 0, len(queue.Records))
		// repeatStomper is used to collapse duplicate download IDs.
		repeatStomper := make(map[string]struct{}, len(queue.Records))

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			} else if _, exists := repeatStomper[item.DownloadID]; exists {
				continue
			}

			repeatStomper[item.DownloadID] = struct{}{}
			// Create minimal copy with only fields needed for stuck item detection.
			appqueue = append(appqueue, &radarrRecord{QueueRecord: minimalRadarrRecord(item)}) //nolint:wsl
		}

		stuck[instance] = listItem{Name: app.Name, Queue: appqueue, Total: len(appqueue)}
		mnd.Log.Debugf("Checking Radarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			instance, len(queue.Records), len(appqueue))
	}

	return stuck
}

func (c *cmd) getFinishedItemsReadarr(_ context.Context) itemList { //nolint:cyclop
	id := logs.Log.Trace("", "start: getFinishedItemsReadarr")
	defer logs.Log.Trace(id, "end: getFinishedItemsReadarr")

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
		// Pre-allocate with capacity to reduce allocations during append.
		appqueue := make([]*readarrRecord, 0, len(queue.Records))
		// repeatStomper is used to collapse duplicate download IDs.
		repeatStomper := make(map[string]struct{}, len(queue.Records))

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			} else if _, exists := repeatStomper[item.DownloadID]; exists {
				continue
			}

			repeatStomper[item.DownloadID] = struct{}{}
			// Create minimal copy with only fields needed for stuck item detection.
			appqueue = append(appqueue, &readarrRecord{QueueRecord: minimalReadarrRecord(item)}) //nolint:wsl
		}

		stuck[instance] = listItem{Name: app.Name, Queue: appqueue, Total: len(appqueue)}
		mnd.Log.Debugf("Checking Readarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			instance, len(queue.Records), len(appqueue))
	}

	return stuck
}

func (c *cmd) getFinishedItemsSonarr(_ context.Context) itemList { //nolint:cyclop
	id := logs.Log.Trace("", "start: getFinishedItemsSonarr")
	defer logs.Log.Trace(id, "end: getFinishedItemsSonarr")

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
		// Pre-allocate with capacity to reduce allocations during append.
		appqueue := make([]*sonarrRecord, 0, len(queue.Records))
		// repeatStomper is used to collapse duplicate download IDs.
		repeatStomper := make(map[string]struct{}, len(queue.Records))

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			} else if _, exists := repeatStomper[item.DownloadID]; exists {
				continue
			}

			repeatStomper[item.DownloadID] = struct{}{}
			// Create minimal copy with only fields needed for stuck item detection.
			appqueue = append(appqueue, &sonarrRecord{QueueRecord: minimalSonarrRecord(item)}) //nolint:wsl
		}

		stuck[instance] = listItem{Name: app.Name, Queue: appqueue, Total: len(appqueue)}
		mnd.Log.Debugf("Checking Sonarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			instance, len(queue.Records), len(appqueue))
	}

	return stuck
}

// minimalLidarrRecord creates a copy of the QueueRecord with only fields needed for stuck item detection.
// This reduces payload size by omitting progress info and metadata not relevant to stuck items.
func minimalLidarrRecord(record *lidarr.QueueRecord) *lidarr.QueueRecord {
	return &lidarr.QueueRecord{
		ID:                    record.ID,
		ArtistID:              record.ArtistID,
		AlbumID:               record.AlbumID,
		Title:                 record.Title,
		Status:                record.Status,
		TrackedDownloadStatus: record.TrackedDownloadStatus,
		StatusMessages:        record.StatusMessages,
		ErrorMessage:          record.ErrorMessage,
		DownloadID:            record.DownloadID,
		DownloadClient:        record.DownloadClient,
		// Omitted: Size, Sizeleft, Timeleft, EstimatedCompletionTime, Quality,
		// OutputPath, Indexer, Protocol, HasPostImportCategory, DownloadForced
	}
}

// minimalRadarrRecord creates a copy of the QueueRecord with only fields needed for stuck item detection.
// This reduces payload size by omitting progress info and metadata not relevant to stuck items.
func minimalRadarrRecord(record *radarr.QueueRecord) *radarr.QueueRecord {
	return &radarr.QueueRecord{
		ID:                    record.ID,
		MovieID:               record.MovieID,
		Title:                 record.Title,
		Status:                record.Status,
		TrackedDownloadStatus: record.TrackedDownloadStatus,
		TrackedDownloadState:  record.TrackedDownloadState,
		StatusMessages:        record.StatusMessages,
		ErrorMessage:          record.ErrorMessage,
		DownloadID:            record.DownloadID,
		DownloadClient:        record.DownloadClient,
		// Omitted: Size, Sizeleft, Timeleft, EstimatedCompletionTime, Quality,
		// CustomFormats, Languages, OutputPath, Indexer, Protocol, HasPostImportCategory
	}
}

// minimalReadarrRecord creates a copy of the QueueRecord with only fields needed for stuck item detection.
// This reduces payload size by omitting progress info and metadata not relevant to stuck items.
func minimalReadarrRecord(record *readarr.QueueRecord) *readarr.QueueRecord {
	return &readarr.QueueRecord{
		ID:                    record.ID,
		AuthorID:              record.AuthorID,
		BookID:                record.BookID,
		Title:                 record.Title,
		Status:                record.Status,
		TrackedDownloadStatus: record.TrackedDownloadStatus,
		TrackedDownloadState:  record.TrackedDownloadState,
		StatusMessages:        record.StatusMessages,
		ErrorMessage:          record.ErrorMessage,
		DownloadID:            record.DownloadID,
		DownloadClient:        record.DownloadClient,
		// Omitted: Size, Sizeleft, Timeleft, EstimatedCompletionTime, Quality,
		// OutputPath, Indexer, Protocol, HasPostImportCategory, DownloadForced
	}
}

// minimalSonarrRecord creates a copy of the QueueRecord with only fields needed for stuck item detection.
// This reduces payload size by omitting progress info and metadata not relevant to stuck items.
func minimalSonarrRecord(record *sonarr.QueueRecord) *sonarr.QueueRecord {
	return &sonarr.QueueRecord{
		ID:                    record.ID,
		SeriesID:              record.SeriesID,
		EpisodeID:             record.EpisodeID,
		Title:                 record.Title,
		Status:                record.Status,
		TrackedDownloadStatus: record.TrackedDownloadStatus,
		TrackedDownloadState:  record.TrackedDownloadState,
		StatusMessages:        record.StatusMessages,
		ErrorMessage:          record.ErrorMessage,
		DownloadID:            record.DownloadID,
		DownloadClient:        record.DownloadClient,
		// Omitted: Size, Sizeleft, Timeleft, EstimatedCompletionTime, Quality,
		// Language, OutputPath, Indexer, Protocol, HasPostImportCategory
	}
}

// Ensure starr import is used for StatusMessage type reference.
var _ *starr.StatusMessage
