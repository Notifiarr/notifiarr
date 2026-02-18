package dashboard

import (
	"testing"
	"time"

	"golift.io/starr/lidarr"
)

func TestApplyLidarrAlbumStatsAggregatesCounts(t *testing.T) {
	now := time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC)
	state := &State{Next: []*Sortable{}}

	albums := []*lidarr.Album{
		{
			ID:         1,
			ArtistID:   100,
			Title:      "Future Missing",
			Monitored:  true,
			ReleaseDate: now.Add(24 * time.Hour),
			Artist:     &lidarr.Artist{ArtistName: "Artist A"},
			Statistics: &lidarr.Statistics{
				PercentOfTracks: 50,
				SizeOnDisk:      100,
				TotalTrackCount: 10,
				TrackCount:      10,
				TrackFileCount:  8,
			},
		},
		{
			ID:         2,
			ArtistID:   100,
			Title:      "Future Complete",
			Monitored:  true,
			ReleaseDate: now.Add(48 * time.Hour),
			Artist:     &lidarr.Artist{ArtistName: "Artist A"},
			Statistics: &lidarr.Statistics{
				PercentOfTracks: 100,
				SizeOnDisk:      200,
				TotalTrackCount: 5,
				TrackCount:      5,
				TrackFileCount:  5,
			},
		},
		{
			ID:          3,
			ArtistID:    200,
			Title:       "Future No Stats",
			Monitored:   true,
			ReleaseDate: now.Add(72 * time.Hour),
			Artist:      &lidarr.Artist{ArtistName: "Artist B"},
		},
	}

	applyLidarrAlbumStats(state, albums, func() time.Time { return now })

	if state.Albums != 3 {
		t.Fatalf("expected Albums=3, got %d", state.Albums)
	}

	if state.Artists != 1 {
		t.Fatalf("expected Artists=1 (only stats-backed artists), got %d", state.Artists)
	}

	if state.Tracks != 15 {
		t.Fatalf("expected Tracks=15, got %d", state.Tracks)
	}

	if state.Missing != 2 {
		t.Fatalf("expected Missing=2, got %d", state.Missing)
	}

	if state.OnDisk != 13 {
		t.Fatalf("expected OnDisk=13, got %d", state.OnDisk)
	}

	if state.Size != 300 {
		t.Fatalf("expected Size=300, got %d", state.Size)
	}

	if state.Percent != 10 {
		t.Fatalf("expected Percent=10, got %v", state.Percent)
	}

	if len(state.Next) != 2 {
		t.Fatalf("expected 2 upcoming albums, got %d", len(state.Next))
	}
}

func TestApplyLidarrAlbumStatsSetsPercentTo100WhenNoTracks(t *testing.T) {
	now := time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC)
	state := &State{Next: []*Sortable{}}

	albums := []*lidarr.Album{
		{
			ID:          1,
			ArtistID:    100,
			Title:       "No Stats",
			Monitored:   true,
			ReleaseDate: now.Add(24 * time.Hour),
			Artist:      &lidarr.Artist{ArtistName: "Artist A"},
		},
	}

	applyLidarrAlbumStats(state, albums, func() time.Time { return now })

	if state.Tracks != 0 {
		t.Fatalf("expected Tracks=0, got %d", state.Tracks)
	}

	if state.Percent != 100 {
		t.Fatalf("expected Percent=100 when no tracks are present, got %v", state.Percent)
	}
}

func TestAccumulateLidarrAlbumStatsAcrossPages(t *testing.T) {
	now := time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC)
	state := &State{Next: []*Sortable{}}
	artistIDs := map[int64]struct{}{}

	accumulateLidarrAlbumStats(state, artistIDs, []*lidarr.Album{
		{
			ID:          1,
			ArtistID:    100,
			Title:       "Page One",
			Monitored:   true,
			ReleaseDate: now.Add(24 * time.Hour),
			Artist:      &lidarr.Artist{ArtistName: "Artist A"},
			Statistics: &lidarr.Statistics{
				PercentOfTracks: 50,
				SizeOnDisk:      100,
				TotalTrackCount: 10,
				TrackCount:      10,
				TrackFileCount:  8,
			},
		},
	}, now)

	accumulateLidarrAlbumStats(state, artistIDs, []*lidarr.Album{
		{
			ID:          2,
			ArtistID:    200,
			Title:       "Page Two",
			Monitored:   true,
			ReleaseDate: now.Add(48 * time.Hour),
			Artist:      &lidarr.Artist{ArtistName: "Artist B"},
			Statistics: &lidarr.Statistics{
				PercentOfTracks: 100,
				SizeOnDisk:      200,
				TotalTrackCount: 5,
				TrackCount:      5,
				TrackFileCount:  5,
			},
		},
	}, now)

	finalizeLidarrAlbumStats(state, artistIDs)

	if state.Albums != 2 {
		t.Fatalf("expected Albums=2, got %d", state.Albums)
	}

	if state.Artists != 2 {
		t.Fatalf("expected Artists=2, got %d", state.Artists)
	}

	if state.Percent != 10 {
		t.Fatalf("expected Percent=10, got %v", state.Percent)
	}
}
