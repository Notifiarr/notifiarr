package notifiarr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/exp"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/plex"
	"golift.io/datacounter"
)

// plexIncomingWebhook is the incoming webhook from a Plex Media Server.
//nolint:tagliatelle
type plexIncomingWebhook struct {
	Event   string  `json:"event"`
	User    bool    `json:"user"`
	Owner   bool    `json:"owner"`
	Rating  float64 `json:"rating"`
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
		LibrarySectionType   string `json:"librarySectionType"`
		RatingKey            string `json:"ratingKey"`
		ParentRatingKey      string `json:"parentRatingKey"`
		GrandparentRatingKey string `json:"grandparentRatingKey"`
		Key                  string `json:"key"`
		GUID                 string `json:"guid"`
		ParentGUID           string `json:"parentGuid"`
		GrandparentGUID      string `json:"grandparentGuid"`
		GuID                 []struct {
			ID string `json:"id"`
		} `json:"Guid"`
		Studio                string      `json:"studio"`
		Type                  string      `json:"type"`
		GrandParentTitle      string      `json:"grandparentTitle"`
		GrandparentKey        string      `json:"grandparentKey"`
		ParentKey             string      `json:"parentKey"`
		ParentTitle           string      `json:"parentTitle"`
		ParentYear            int         `json:"parentYear"`
		ParentThumb           string      `json:"parentThumb"`
		GrandparentThumb      string      `json:"grandparentThumb"`
		GrandparentArt        string      `json:"grandparentArt"`
		GrandparentTheme      string      `json:"grandparentTheme"`
		ParentIndex           int64       `json:"parentIndex"`
		Index                 int64       `json:"index"`
		Title                 string      `json:"title"`
		TitleSort             string      `json:"titleSort"`
		LibrarySectionTitle   string      `json:"librarySectionTitle"`
		LibrarySectionID      int         `json:"librarySectionID"`
		LibrarySectionKey     string      `json:"librarySectionKey"`
		ContentRating         string      `json:"contentRating"`
		Summary               string      `json:"summary"`
		Rating                float64     `json:"rating"`
		ExternalRating        interface{} `json:"Rating,omitempty"` // bullshit.
		AudienceRating        float64     `json:"audienceRating"`
		ViewOffset            int         `json:"viewOffset"`
		LastViewedAt          int         `json:"lastViewedAt"`
		Year                  int         `json:"year"`
		Tagline               string      `json:"tagline"`
		Thumb                 string      `json:"thumb"`
		Art                   string      `json:"art"`
		Duration              int         `json:"duration"`
		OriginallyAvailableAt string      `json:"originallyAvailableAt"`
		AddedAt               int         `json:"addedAt"`
		UpdatedAt             int         `json:"updatedAt"`
		AudienceRatingImage   string      `json:"audienceRatingImage"`
		PrimaryExtraKey       string      `json:"primaryExtraKey"`
		RatingImage           string      `json:"ratingImage"`
	} `json:"Metadata"`
}

// PlexHandler handles an incoming webhook from Plex.
func (c *Config) PlexHandler(w http.ResponseWriter, r *http.Request) { //nolint:cyclop,varnamelen,funlen
	exp.Apps.Add("Plex&&Incoming Webhooks", 1)

	start := time.Now()
	rcvd := datacounter.NewReaderCounter(r.Body)
	r.Body = &apps.FakeCloser{
		App:     "Plex",
		Rcvd:    rcvd.Count,   // This gets added...
		CloseFn: r.Body.Close, // when this gets called.
		Reader:  rcvd,
	}

	if err := r.ParseMultipartForm(mnd.Megabyte); err != nil {
		c.Errorf("Parsing Multipart Form (plex): %v", err)
		exp.Apps.Add("Plex&&Webhook Errors", 1)
		http.Error(w, "form parse error", http.StatusBadRequest)

		return
	}

	payload := r.Form.Get("payload")
	c.Debugf("Plex Webhook Payload: %s", payload)
	r.Header.Set("X-Request-Time", fmt.Sprintf("%dms", time.Since(start).Milliseconds()))
	exp.Apps.Add("Plex&&Bytes Received", int64(rcvd.Count()))

	var v plexIncomingWebhook

	switch err := json.Unmarshal([]byte(payload), &v); {
	case err != nil:
		exp.Apps.Add("Plex&&Webhook Errors", 1)
		http.Error(w, "payload error", http.StatusBadRequest)
		c.Errorf("Unmarshalling Plex payload: %v", err)
	case strings.EqualFold(v.Event, "admin.database.backup"):
		fallthrough
	case strings.EqualFold(v.Event, "device.new"):
		fallthrough
	case strings.EqualFold(v.Event, "media.rate"):
		fallthrough
	case strings.EqualFold(v.Event, "library.new"):
		fallthrough
	case strings.EqualFold(v.Event, "admin.database.corrupt"):
		c.Printf("Plex Incoming Webhook: %s, %s '%s' ~> %s (relaying to Notifiarr)",
			v.Server.Title, v.Account.Title, v.Event, v.Metadata.Title)
		c.QueueData(&SendRequest{
			Route:      PlexRoute,
			Event:      EventHook,
			LogPayload: true,
			LogMsg:     fmt.Sprintf("Plex Webhhok: %s '%s' ~> %s", v.Account.Title, v.Event, v.Metadata.Title),
			Payload:    &Payload{Load: &v, Plex: &plex.Sessions{Name: c.Plex.Name}},
		})
		r.Header.Set("X-Request-Time", fmt.Sprintf("%dms", time.Since(start).Milliseconds()))
		http.Error(w, "process", http.StatusAccepted)
	case strings.EqualFold(v.Event, "media.resume") && c.plexTimer.Active(c.Plex.Cooldown.Duration):
		c.Printf("Plex Incoming Webhook Ignored (cooldown): %s, %s '%s' ~> %s",
			v.Server.Title, v.Account.Title, v.Event, v.Metadata.Title)
		http.Error(w, "ignored, cooldown", http.StatusAlreadyReported)
	case strings.EqualFold(v.Event, "media.play"):
		fallthrough
	case strings.EqualFold(v.Event, "media.resume"):
		go c.collectSessions(EventHook, &v)
		c.Printf("Plex Incoming Webhook: %s, %s '%s' ~> %s (collecting sessions)",
			v.Server.Title, v.Account.Title, v.Event, v.Metadata.Title)
		r.Header.Set("X-Request-Time", fmt.Sprintf("%dms", time.Since(start).Milliseconds()))
		http.Error(w, "processing", http.StatusAccepted)
	default:
		http.Error(w, "ignored, unsupported", http.StatusAlreadyReported)
		c.Printf("Plex Incoming Webhook Ignored (unsupported): %s, %s '%s' ~> %s",
			v.Server.Title, v.Account.Title, v.Event, v.Metadata.Title)
	}
}
