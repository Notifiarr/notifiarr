package starrqueue

import (
	"context"
	"fmt"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"golift.io/cache"
	"golift.io/starr/lidarr"
	"golift.io/starr/radarr"
	"golift.io/starr/readarr"
	"golift.io/starr/sonarr"
)

const maxQueuePayloadSize = 15

type lidarrRecord struct {
	*lidarr.QueueRecord
	Name            string `json:"name"`
	ArtistTitle     string `json:"artistTitle"`
	ForeignAlbumID  string `json:"foreignAlbumId"`
	ForeignArtistID string `json:"foreignArtistId"`
}

type radarrRecord struct {
	*radarr.QueueRecord
	Name           string `json:"name"`
	ForeignMovieID int64  `json:"foreignMovieId"`
}

type sonarrRecord struct {
	*sonarr.QueueRecord
	Name            string `json:"name"`
	ForeignSeriesID int64  `json:"foreignSeriesId"`
}

type readarrRecord struct {
	*readarr.QueueRecord
	Name            string `json:"name"`
	AuthorTitle     string `json:"authorTitle"`
	ForeignBookID   string `json:"foreignBookId"`
	ForeignAuthorID string `json:"foreignAuthorId"`
}

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
		ErrorsOnly: !c.DebugEnabled(),
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

func (c *cmd) getDownloadingItemsLidarr(ctx context.Context) itemList {
	items := make(itemList)

	info := clientinfo.Get()
	if info == nil {
		return items
	}

	for idx, app := range c.Apps.Lidarr {
		instance := idx + 1
		if !app.Enabled() || !info.Actions.Apps.Lidarr.Finished(instance) {
			continue
		}

		cacheItem := data.GetWithID("lidarr", idx)
		if cacheItem == nil || cacheItem.Data == nil {
			continue
		}

		queue, _ := cacheItem.Data.(*lidarr.Queue)
		lidarrQueue := c.rangeDownloadingItemsLidarr(ctx, idx, app, queue.Records)
		items[instance] = listItem{Name: app.Name, Queue: lidarrQueue, Total: queue.TotalRecords}
	}

	return items
}

func (c *cmd) rangeDownloadingItemsLidarr(
	ctx context.Context,
	idx int,
	app *apps.LidarrConfig,
	records []*lidarr.QueueRecord,
) []*lidarrRecord {
	lidarrQueue := []*lidarrRecord{}
	// repeatStomper is used to collapse duplicate download IDs.
	repeatStomper := make(map[string]struct{})

	for _, item := range records {
		if len(lidarrQueue) >= maxQueuePayloadSize {
			break
		}

		// Delay items have no download ID, so group (de-duplicate) them by size.
		if item.DownloadID == "" {
			item.DownloadID = fmt.Sprint(item.Size)
		}

		_, exists := repeatStomper[item.DownloadID]
		if s := strings.ToLower(item.Status); (s != downloading && s != delay) || exists {
			continue
		}

		// We have to connect back to the starr app and pull meta data for the active downloading item.
		// The data gets cached for a while so this extra api hit should only happen once for each item.
		cacheItem := data.GetWithID(fmt.Sprint("lidarrAlbum", item.AlbumID), idx)
		if cacheItem == nil || cacheItem.Data == nil {
			album, err := app.GetAlbumByIDContext(ctx, item.AlbumID)
			if err != nil {
				c.Errorf("Getting data for downloading item: %v", err)
				cacheItem = &cache.Item{Data: &lidarr.Album{Artist: &lidarr.Artist{}}} //nolint:wsl
			} else {
				data.SaveWithID(fmt.Sprint("lidarrAlbum", item.AlbumID), idx, album)
				cacheItem = &cache.Item{Data: album}
			}
		}

		album, _ := cacheItem.Data.(*lidarr.Album)
		repeatStomper[item.DownloadID] = struct{}{}
		lidarrQueue = append(lidarrQueue, &lidarrRecord{ //nolint:wsl
			QueueRecord:     item,
			Name:            album.Title,
			ArtistTitle:     album.Artist.ArtistName,
			ForeignAlbumID:  album.ForeignAlbumID,
			ForeignArtistID: album.Artist.ForeignArtistID,
		})
	}

	return lidarrQueue
}

func (c *cmd) getDownloadingItemsRadarr(ctx context.Context) itemList {
	items := make(itemList)

	info := clientinfo.Get()
	if info == nil {
		return items
	}

	for idx, app := range c.Apps.Radarr {
		instance := idx + 1
		if !app.Enabled() || !info.Actions.Apps.Radarr.Finished(instance) {
			continue
		}

		cacheItem := data.GetWithID("radarr", idx)
		if cacheItem == nil || cacheItem.Data == nil {
			continue
		}

		queue, _ := cacheItem.Data.(*radarr.Queue)
		radarrQueue := c.rangeDownloadingItemsRadarr(ctx, idx, app, queue.Records)
		items[instance] = listItem{Name: app.Name, Queue: radarrQueue, Total: queue.TotalRecords}
	}

	return items
}

func (c *cmd) rangeDownloadingItemsRadarr(
	ctx context.Context,
	idx int,
	app *apps.RadarrConfig,
	records []*radarr.QueueRecord,
) []*radarrRecord {
	radarrQueue := []*radarrRecord{}
	// repeatStomper is used to collapse duplicate download IDs.
	repeatStomper := make(map[string]struct{})

	for _, item := range records {
		if len(radarrQueue) >= maxQueuePayloadSize {
			break
		}

		// Delay items have no download ID, so group (de-duplicate) them by size.
		if item.DownloadID == "" {
			item.DownloadID = fmt.Sprint(item.Size)
		}

		_, exists := repeatStomper[item.DownloadID]
		if s := strings.ToLower(item.Status); (s != downloading && s != delay) || exists {
			continue
		}

		// We have to connect back to the starr app and pull meta data for the active downloading item.
		// The data gets cached for a while so this extra api hit should only happen once for each item.
		cacheItem := data.GetWithID(fmt.Sprint("radarrMovie", item.MovieID), idx)
		if cacheItem == nil || cacheItem.Data == nil {
			movie, err := app.GetMovieByIDContext(ctx, item.MovieID)
			if err != nil {
				c.Errorf("Getting data for downloading item: %v", err)
				cacheItem = &cache.Item{Data: &radarr.Movie{}} //nolint:wsl
			} else {
				data.SaveWithID(fmt.Sprint("radarrMovie", item.MovieID), idx, movie)
				cacheItem = &cache.Item{Data: movie}
			}
		}

		movie, _ := cacheItem.Data.(*radarr.Movie)
		repeatStomper[item.DownloadID] = struct{}{}
		radarrQueue = append(radarrQueue, &radarrRecord{ //nolint:wsl
			QueueRecord:    item,
			Name:           movie.Title,
			ForeignMovieID: movie.TmdbID,
		})
	}

	return radarrQueue
}

func (c *cmd) getDownloadingItemsReadarr(ctx context.Context) itemList {
	items := make(itemList)

	info := clientinfo.Get()
	if info == nil {
		return items
	}

	for idx, app := range c.Apps.Readarr {
		instance := idx + 1
		if !app.Enabled() || !info.Actions.Apps.Readarr.Finished(instance) {
			continue
		}

		cacheItem := data.GetWithID("readarr", idx)
		if cacheItem == nil || cacheItem.Data == nil {
			continue
		}

		queue, _ := cacheItem.Data.(*readarr.Queue)
		readarrQueue := c.rangeDownloadingItemsReadarr(ctx, idx, app, queue.Records)
		items[instance] = listItem{Name: app.Name, Queue: readarrQueue, Total: queue.TotalRecords}
	}

	return items
}

func (c *cmd) rangeDownloadingItemsReadarr(
	ctx context.Context,
	idx int,
	app *apps.ReadarrConfig,
	records []*readarr.QueueRecord,
) []*readarrRecord {
	readarrQueue := []*readarrRecord{}
	// repeatStomper is used to collapse duplicate download IDs.
	repeatStomper := make(map[string]struct{})

	for _, item := range records {
		if len(readarrQueue) >= maxQueuePayloadSize {
			break
		}

		// Delay items have no download ID, so group (de-duplicate) them by size.
		if item.DownloadID == "" {
			item.DownloadID = fmt.Sprint(item.Size)
		}

		_, exists := repeatStomper[item.DownloadID]
		if s := strings.ToLower(item.Status); (s != downloading && s != delay) || exists {
			continue
		}

		// We have to connect back to the starr app and pull meta data for the active downloading item.
		// The data gets cached for a while so this extra api hit should only happen once for each item.
		cacheItem := data.GetWithID(fmt.Sprint("readarrBook", item.BookID), idx)
		if cacheItem == nil || cacheItem.Data == nil {
			book, err := app.GetBookByIDContext(ctx, item.BookID)
			if err != nil {
				c.Errorf("Getting data for downloading item: %v", err)
				cacheItem = &cache.Item{Data: &readarr.Book{Author: &readarr.Author{}}} //nolint:wsl
			} else {
				data.SaveWithID(fmt.Sprint("readarrBook", item.BookID), idx, book)
				cacheItem = &cache.Item{Data: book}
			}
		}

		book, _ := cacheItem.Data.(*readarr.Book)
		repeatStomper[item.DownloadID] = struct{}{}
		readarrQueue = append(readarrQueue, &readarrRecord{ //nolint:wsl
			QueueRecord:     item,
			Name:            book.Title,
			AuthorTitle:     book.AuthorTitle,
			ForeignBookID:   book.ForeignBookID,
			ForeignAuthorID: book.Author.ForeignAuthorID,
		})
	}

	return readarrQueue
}

func (c *cmd) getDownloadingItemsSonarr(ctx context.Context) itemList {
	items := make(itemList)

	info := clientinfo.Get()
	if info == nil {
		return items
	}

	for idx, app := range c.Apps.Sonarr {
		instance := idx + 1
		if !app.Enabled() || !info.Actions.Apps.Sonarr.Finished(instance) {
			continue
		}

		cacheItem := data.GetWithID("sonarr", idx)
		if cacheItem == nil || cacheItem.Data == nil {
			continue
		}

		queue, _ := cacheItem.Data.(*sonarr.Queue)
		sonarrQueue := c.rangeDownloadingItemsSonarr(ctx, idx, app, queue.Records)
		items[instance] = listItem{Name: app.Name, Queue: sonarrQueue, Total: queue.TotalRecords}
	}

	return items
}

func (c *cmd) rangeDownloadingItemsSonarr(
	ctx context.Context,
	idx int,
	app *apps.SonarrConfig,
	records []*sonarr.QueueRecord,
) []*sonarrRecord {
	sonarrQueue := []*sonarrRecord{}
	// repeatStomper is used to collapse duplicate download IDs.
	repeatStomper := make(map[string]struct{})

	for _, item := range records {
		if len(sonarrQueue) >= maxQueuePayloadSize {
			break
		}

		// Delay items have no download ID, so group (de-duplicate) them by size.
		if item.DownloadID == "" {
			item.DownloadID = fmt.Sprint(item.Size)
		}

		_, exists := repeatStomper[item.DownloadID]
		if s := strings.ToLower(item.Status); (s != downloading && s != delay) || exists {
			continue
		}

		// We have to connect back to the starr app and pull meta data for the active downloading item.
		// The data gets cached for a while so this extra api hit should only happen once for each item.
		cacheItem := data.GetWithID(fmt.Sprint("sonarrSeries", item.SeriesID), idx)
		if cacheItem == nil || cacheItem.Data == nil {
			series, err := app.GetSeriesByIDContext(ctx, item.SeriesID)
			if err != nil {
				c.Errorf("Getting data for downloading item: %v", err)
				cacheItem = &cache.Item{Data: &sonarr.Series{}} //nolint:wsl
			} else {
				data.SaveWithID(fmt.Sprint("sonarrSeries", item.SeriesID), idx, series)
				cacheItem = &cache.Item{Data: series}
			}
		}

		series, _ := cacheItem.Data.(*sonarr.Series)
		repeatStomper[item.DownloadID] = struct{}{}
		sonarrQueue = append(sonarrQueue, &sonarrRecord{ //nolint:wsl
			QueueRecord:     item,
			Name:            series.Title,
			ForeignSeriesID: series.TvdbID,
		})
	}

	return sonarrQueue
}
