package plexcron

import (
	"context"
	"fmt"

	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

// typo killers.
const (
	mediaResume = "media.resume"
	mediaPlay   = "media.play"
	playing     = "playing"
	paused      = "paused"
)

// sendSessionNew is used when the end user does not have or use Plex webhooks.
// They can enable the plex session tracker to send notifications for new sessions.
// event is either media.play or media.resume.
func (c *cmd) sendSessionPlaying(ctx context.Context, session *plex.Session, sessions *plex.Sessions, event string) {
	if err := c.checkPlexAgent(ctx, session); err != nil {
		mnd.Log.Errorf("Failed Plex Request: %v", err)
		return
	}

	website.Site.SendData(&website.Request{
		Route: website.PlexRoute,
		Event: website.EventHook,
		Payload: &website.Payload{
			Snap: c.getMetaSnap(ctx),
			Plex: sessions,
			Load: convertSessionsToWebhook(session, event),
		},
		LogMsg: fmt.Sprintf("Plex New Session on %s {%s/%s} %s => %s: %s (%s)",
			c.Plex.Server.Name(), session.Session.ID, session.SessionKey, session.User.Title,
			session.Type, session.Title, session.Player.State),
		LogPayload: true,
	})
}

// convertSessionsToWebhook exists to shoehorn a "session" into a webhook.
// This is because "playing" and "resume" was originally written around Plex Webhooks.
// And then we decided to make this work without webhooks. Since the website already
// knows how to deal with the webhook, we are converting a session into that same payload.
func convertSessionsToWebhook(session *plex.Session, event string) *plex.IncomingWebhook {
	return &plex.IncomingWebhook{
		Event:  event,
		User:   true,
		Rating: session.Rating,
		Account: struct {
			ID    int    `json:"id"`
			Thumb string `json:"thumb"`
			Title string `json:"title"`
		}{Title: session.User.Title},
		Player: struct {
			Local         bool   `json:"local"`
			PublicAddress string `json:"publicAddress"`
			Title         string `json:"title"`
			UUID          string `json:"uuid"`
		}{Title: session.Player.Title},
		Metadata: plex.WebhookMetadata{
			LibrarySectionType:    "", // does not exist
			RatingKey:             session.RatingKey,
			ParentRatingKey:       session.ParentRatingKey,
			GrandparentRatingKey:  session.GrandparentRatingKey,
			Key:                   session.Key,
			GUID:                  session.GUID,
			ParentGUID:            session.ParentGUID,
			GrandparentGUID:       session.GrandparentGUID,
			GuID:                  session.GuID,
			Studio:                session.Studio,
			Type:                  session.Type,
			GrandParentTitle:      session.GrandparentTitle,
			GrandparentKey:        session.GrandparentKey,
			ParentKey:             session.ParentKey,
			ParentTitle:           session.ParentTitle,
			ParentYear:            0, // does not exist
			ParentThumb:           session.ParentThumb,
			GrandparentThumb:      session.GrandparentThumb,
			GrandparentArt:        session.GrandparentArt,
			GrandparentTheme:      session.GrandparentTheme,
			ParentIndex:           session.ParentIndex,
			Index:                 session.Index,
			Title:                 session.Title,
			TitleSort:             session.TitleSort,
			LibrarySectionTitle:   session.LibrarySectionTitle,
			LibrarySectionID:      session.LibrarySectionID,
			LibrarySectionKey:     session.LibrarySectionKey,
			ContentRating:         session.ContentRating,
			Summary:               session.Summary,
			Rating:                session.Rating,
			ExternalRating:        session.ExternalRating,
			AudienceRating:        session.AudienceRating,
			ViewOffset:            session.ViewOffset,
			LastViewedAt:          session.LastViewed,
			Year:                  session.Year,
			Tagline:               "", // does not exist
			Thumb:                 session.Thumb,
			Art:                   session.Art,
			Duration:              session.Duration,
			OriginallyAvailableAt: session.OriginallyAvailable,
			AddedAt:               session.Added,
			UpdatedAt:             session.Updated,
			AudienceRatingImage:   session.AudienceRatingImg,
			PrimaryExtraKey:       session.PrimaryExtraKey,
			RatingImage:           session.RatingImage,
		},
	}
}
