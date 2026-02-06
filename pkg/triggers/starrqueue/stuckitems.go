package starrqueue

import (
	"context"
	"fmt"
	"strings"

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
		mnd.Log.Debugf("[%s requested] No stuck items found.", input.Type)
		return
	}

	website.SendData(&website.Request{
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
			// Create minimal copy with only fields needed for stuck item detection.
			appqueue = append(appqueue, &lidarrRecord{QueueRecord: minimalLidarrRecord(item)}) //nolint:wsl
		}

		stuck[instance] = listItem{Name: app.Name, Queue: appqueue, Total: queue.TotalRecords}
		mnd.Log.Debugf("Checking Lidarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
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
			// Create minimal copy with only fields needed for stuck item detection.
			appqueue = append(appqueue, &radarrRecord{QueueRecord: minimalRadarrRecord(item)}) //nolint:wsl
		}

		stuck[instance] = listItem{Name: app.Name, Queue: appqueue, Total: queue.TotalRecords}
		mnd.Log.Debugf("Checking Radarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
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
			// Create minimal copy with only fields needed for stuck item detection.
			appqueue = append(appqueue, &readarrRecord{QueueRecord: minimalReadarrRecord(item)}) //nolint:wsl
		}

		stuck[instance] = listItem{Name: app.Name, Queue: appqueue, Total: queue.TotalRecords}
		mnd.Log.Debugf("Checking Readarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
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
			// Create minimal copy with only fields needed for stuck item detection.
			appqueue = append(appqueue, &sonarrRecord{QueueRecord: minimalSonarrRecord(item)}) //nolint:wsl
		}

		stuck[instance] = listItem{Name: app.Name, Queue: appqueue, Total: queue.TotalRecords}
		mnd.Log.Debugf("Checking Sonarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			instance, len(queue.Records), len(appqueue))
	}

	return stuck
}

// minimalLidarrRecord creates a copy of the QueueRecord with only fields needed for stuck item detection.
// This reduces payload size by omitting progress info and metadata not relevant to stuck items.
func minimalLidarrRecord(r *lidarr.QueueRecord) *lidarr.QueueRecord {
	return &lidarr.QueueRecord{
		ID:                    r.ID,
		ArtistID:              r.ArtistID,
		AlbumID:               r.AlbumID,
		Title:                 r.Title,
		Size:                  r.Size,
		Sizeleft:              r.Sizeleft,
		Status:                r.Status,
		TrackedDownloadStatus: r.TrackedDownloadStatus,
		StatusMessages:        r.StatusMessages,
		ErrorMessage:          r.ErrorMessage,
		DownloadID:            r.DownloadID,
		Protocol:              r.Protocol,
		DownloadClient:        r.DownloadClient,
		Indexer:               r.Indexer,
		HasPostImportCategory: r.HasPostImportCategory,
		// Omitted: Timeleft, EstimatedCompletionTime, Quality,
		// OutputPath, DownloadForced
	}
}

// minimalRadarrRecord creates a copy of the QueueRecord with only fields needed for stuck item detection.
// This reduces payload size by omitting progress info and metadata not relevant to stuck items.
func minimalRadarrRecord(r *radarr.QueueRecord) *radarr.QueueRecord {
	return &radarr.QueueRecord{
		ID:                    r.ID,
		MovieID:               r.MovieID,
		Title:                 r.Title,
		Size:                  r.Size,
		Sizeleft:              r.Sizeleft,
		Status:                r.Status,
		TrackedDownloadStatus: r.TrackedDownloadStatus,
		TrackedDownloadState:  r.TrackedDownloadState,
		StatusMessages:        r.StatusMessages,
		ErrorMessage:          r.ErrorMessage,
		DownloadID:            r.DownloadID,
		Protocol:              r.Protocol,
		DownloadClient:        r.DownloadClient,
		Indexer:               r.Indexer,
		HasPostImportCategory: r.HasPostImportCategory,
		// Omitted: Timeleft, EstimatedCompletionTime, Quality,
		// CustomFormats, Languages, OutputPath
	}
}

// minimalReadarrRecord creates a copy of the QueueRecord with only fields needed for stuck item detection.
// This reduces payload size by omitting progress info and metadata not relevant to stuck items.
func minimalReadarrRecord(r *readarr.QueueRecord) *readarr.QueueRecord {
	return &readarr.QueueRecord{
		ID:                    r.ID,
		AuthorID:              r.AuthorID,
		BookID:                r.BookID,
		Title:                 r.Title,
		Size:                  r.Size,
		Sizeleft:              r.Sizeleft,
		Status:                r.Status,
		TrackedDownloadStatus: r.TrackedDownloadStatus,
		TrackedDownloadState:  r.TrackedDownloadState,
		StatusMessages:        r.StatusMessages,
		ErrorMessage:          r.ErrorMessage,
		DownloadID:            r.DownloadID,
		Protocol:              r.Protocol,
		DownloadClient:        r.DownloadClient,
		Indexer:               r.Indexer,
		HasPostImportCategory: r.HasPostImportCategory,
		// Omitted: Timeleft, EstimatedCompletionTime, Quality,
		// OutputPath, DownloadForced
	}
}

// minimalSonarrRecord creates a copy of the QueueRecord with only fields needed for stuck item detection.
// This reduces payload size by omitting progress info and metadata not relevant to stuck items.
func minimalSonarrRecord(r *sonarr.QueueRecord) *sonarr.QueueRecord {
	return &sonarr.QueueRecord{
		ID:                    r.ID,
		SeriesID:              r.SeriesID,
		EpisodeID:             r.EpisodeID,
		Title:                 r.Title,
		Size:                  r.Size,
		Sizeleft:              r.Sizeleft,
		Status:                r.Status,
		TrackedDownloadStatus: r.TrackedDownloadStatus,
		TrackedDownloadState:  r.TrackedDownloadState,
		StatusMessages:        r.StatusMessages,
		ErrorMessage:          r.ErrorMessage,
		DownloadID:            r.DownloadID,
		Protocol:              r.Protocol,
		DownloadClient:        r.DownloadClient,
		Indexer:               r.Indexer,
		HasPostImportCategory: r.HasPostImportCategory,
		// Omitted: Timeleft, EstimatedCompletionTime, Quality,
		// Language, OutputPath
	}
}

// Ensure starr import is used for StatusMessage type reference.
var _ *starr.StatusMessage
