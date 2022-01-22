package notifiarr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/plex"
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
		Studio                string  `json:"studio"`
		Type                  string  `json:"type"`
		GrandParentTitle      string  `json:"grandparentTitle"`
		GrandparentKey        string  `json:"grandparentKey"`
		ParentKey             string  `json:"parentKey"`
		ParentTitle           string  `json:"parentTitle"`
		ParentYear            int     `json:"parentYear"`
		ParentThumb           string  `json:"parentThumb"`
		GrandparentThumb      string  `json:"grandparentThumb"`
		GrandparentArt        string  `json:"grandparentArt"`
		GrandparentTheme      string  `json:"grandparentTheme"`
		ParentIndex           int64   `json:"parentIndex"`
		Index                 int64   `json:"index"`
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
	} `json:"Metadata"`
}

// PlexHandler handles an incoming webhook from Plex.
func (c *Config) PlexHandler(w http.ResponseWriter, r *http.Request) { //nolint:cyclop,varnamelen
	start := time.Now()

	if err := r.ParseMultipartForm(mnd.KB100); err != nil {
		c.Errorf("Parsing Multipart Form (plex): %v", err)
		c.Apps.Respond(w, http.StatusBadRequest, "form parse error")

		return
	}

	payload := r.Form.Get("payload")
	c.Debugf("Plex Webhook Payload: %s", payload)
	r.Header.Set("X-Request-Time", fmt.Sprintf("%dms", time.Since(start).Milliseconds()))

	var v plexIncomingWebhook

	switch err := json.Unmarshal([]byte(payload), &v); {
	case err != nil:
		c.Apps.Respond(w, http.StatusInternalServerError, "payload error")
		c.Errorf("Unmarshalling Plex payload: %v", err)
	case strings.EqualFold(v.Event, "admin.database.backup"):
		fallthrough
	case strings.EqualFold(v.Event, "device.new"):
		fallthrough
	case strings.EqualFold(v.Event, "media.rate"):
		fallthrough
	case strings.EqualFold(v.Event, "admin.database.corrupt"):
		c.Printf("Plex Incoming Webhook: %s, %s '%s' ~> %s (relaying to Notifiarr)",
			v.Server.Title, v.Account.Title, v.Event, v.Metadata.Title)
		c.sendPlexWebhook(&v)
		r.Header.Set("X-Request-Time", fmt.Sprintf("%dms", time.Since(start).Milliseconds()))
		c.Apps.Respond(w, http.StatusAccepted, "processing")
	case strings.EqualFold(v.Event, "media.resume") && c.plexTimer.Active(c.Plex.Cooldown.Duration):
		c.Printf("Plex Incoming Webhook Ignored (cooldown): %s, %s '%s' ~> %s",
			v.Server.Title, v.Account.Title, v.Event, v.Metadata.Title)
		c.Apps.Respond(w, http.StatusAlreadyReported, "ignored, cooldown")
	case strings.EqualFold(v.Event, "media.play"):
		fallthrough
	case strings.EqualFold(v.Event, "media.resume"):
		go c.collectSessions(EventHook, &v)
		c.Printf("Plex Incoming Webhook: %s, %s '%s' ~> %s (collecting sessions)",
			v.Server.Title, v.Account.Title, v.Event, v.Metadata.Title)
		r.Header.Set("X-Request-Time", fmt.Sprintf("%dms", time.Since(start).Milliseconds()))
		c.Apps.Respond(w, http.StatusAccepted, "processing")
	default:
		c.Apps.Respond(w, http.StatusAlreadyReported, "ignored, unsupported")
		c.Printf("Plex Incoming Webhook Ignored (unsupported): %s, %s '%s' ~> %s",
			v.Server.Title, v.Account.Title, v.Event, v.Metadata.Title)
	}
}

// sendPlexWebhook simply relays an incoming "admin" plex webhook to Notifiarr.com.
func (c *Config) sendPlexWebhook(hook *plexIncomingWebhook) {
	resp, err := c.SendData(PlexRoute.Path(EventHook), &Payload{Load: hook, Plex: &plex.Sessions{Name: c.Plex.Name}}, true)
	if err != nil {
		c.Errorf("Sending Plex Webhook to Notifiarr: %v", err)
		return
	}

	c.Printf("Plex ~> Notifiarr: %s '%s' ~> %s. %s", hook.Account.Title, hook.Event, hook.Metadata.Title, resp)
}
