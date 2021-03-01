package plex

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/Go-Lift-TV/discordnotifier-client/pkg/snapshot"
)

// hookMeta is the payload send to Notifiarr.
type hookMeta struct {
	*Sessions          `json:"plex"`
	*snapshot.Snapshot `json:"snapshot"`
	*Webhook           `json:"payload"`
}

// HandleHook is kicked off by the webserver in go routine. This runs
// after Plex drops off a webhook telling us someone did something.
// This gathers cpu/ram, and waits 10 seconds, then grabs plex sessions.
// It's all POSTed to notifiarr.
func (s *Server) SendMeta(hook *Webhook, apikey string) (b []byte, err error) {
	wg, hm := (&hookMeta{nil, &snapshot.Snapshot{}, hook}).get()

	// Wait 10 seconds for the server to get the session ready
	time.Sleep(plexWaitTime)

	if hm.Sessions, err = s.GetSessions(); err != nil {
		return nil, err
	}

	wg.Wait()

	b, _ = json.Marshal(&hm)
	// log.Println(string(data))

	return snapshot.SendJSON(snapshot.NotifiarrTestURL, apikey, b)
}

func (hm *hookMeta) get() (*sync.WaitGroup, *hookMeta) {
	ctx, cancel := context.WithTimeout(context.Background(), plexWaitTime)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(3) //nolint: gomnd,wsl

	go func() {
		_ = hm.GetCPUSample(ctx, true)
		wg.Done() //nolint:wsl
	}()
	go func() {
		_ = hm.GetMemoryUsage(ctx, true)
		wg.Done() //nolint:wsl
	}()
	go func() {
		_ = hm.GetLocalData(ctx, false)
		wg.Done() //nolint:wsl
	}()

	return &wg, hm
}

// Webhook is the incoming webhook from a Plex Media Server.
type Webhook struct {
	Event   string `json:"event"`
	User    bool   `json:"user"`
	Owner   bool   `json:"owner"`
	Account struct {
		ID    int    `json:"id"`
		Thumb string `json:"thumb"`
		Title string `json:"title"`
	} `json:"Account"`
	Server struct {
		Title string `json:"title"`
		UUID  string `json:"uuid"`
	} `json:"Server"`
	Player struct {
		Local         bool   `json:"local"`
		PublicAddress string `json:"publicAddress"`
		Title         string `json:"title"`
		UUID          string `json:"uuid"`
	} `json:"Player"`
	Metadata struct {
		LibrarySectionType    string  `json:"librarySectionType"`
		RatingKey             string  `json:"ratingKey"`
		Key                   string  `json:"key"`
		GUID                  string  `json:"guid"`
		Studio                string  `json:"studio"`
		Type                  string  `json:"type"`
		Title                 string  `json:"title"`
		TitleSort             string  `json:"titleSort"`
		LibrarySectionTitle   string  `json:"librarySectionTitle"`
		LibrarySectionID      int     `json:"librarySectionID"`
		LibrarySectionKey     string  `json:"librarySectionKey"`
		ContentRating         string  `json:"contentRating"`
		Summary               string  `json:"summary"`
		Rating                float64 `json:"rating"`
		AudienceRating        float64 `json:"audienceRating"`
		ViewOffset            int     `json:"viewOffset"`
		LastViewedAt          int     `json:"lastViewedAt"`
		Year                  int     `json:"year"`
		Tagline               string  `json:"tagline"`
		Thumb                 string  `json:"thumb"`
		Art                   string  `json:"art"`
		Duration              int     `json:"duration"`
		OriginallyAvailableAt string  `json:"originallyAvailableAt"`
		AddedAt               int     `json:"addedAt"`
		UpdatedAt             int     `json:"updatedAt"`
		AudienceRatingImage   string  `json:"audienceRatingImage"`
		PrimaryExtraKey       string  `json:"primaryExtraKey"`
		RatingImage           string  `json:"ratingImage"`
		Genre                 []struct {
			ID     int    `json:"id"`
			Filter string `json:"filter"`
			Tag    string `json:"tag"`
			Count  int    `json:"count"`
		} `json:"Genre"`
		Director []struct {
			ID     int    `json:"id"`
			Filter string `json:"filter"`
			Tag    string `json:"tag"`
		} `json:"Director"`
		Writer []struct {
			ID     int    `json:"id"`
			Filter string `json:"filter"`
			Tag    string `json:"tag"`
			Count  int    `json:"count"`
		} `json:"Writer"`
		Producer []struct {
			ID     int    `json:"id"`
			Filter string `json:"filter"`
			Tag    string `json:"tag"`
			Count  int    `json:"count,omitempty"`
		} `json:"Producer"`
		Country []struct {
			ID     int    `json:"id"`
			Filter string `json:"filter"`
			Tag    string `json:"tag"`
			Count  int    `json:"count"`
		} `json:"Country"`
		Role []struct {
			ID     int    `json:"id"`
			Filter string `json:"filter"`
			Tag    string `json:"tag"`
			Count  int    `json:"count,omitempty"`
			Role   string `json:"role"`
			Thumb  string `json:"thumb,omitempty"`
		} `json:"Role"`
		Similar []struct {
			ID     int    `json:"id"`
			Filter string `json:"filter"`
			Tag    string `json:"tag"`
			Count  int    `json:"count"`
		} `json:"Similar"`
	} `json:"Metadata"`
}
