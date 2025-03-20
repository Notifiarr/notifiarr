//nolint:tagliatelle
package plex

// IncomingWebhook is the incoming webhook from a Plex Media Server.
type IncomingWebhook struct {
	Server struct {
		Title string `json:"title"`
		UUID  string `json:"uuid"`
	} `json:"Server"`
	Event  string `json:"event"`
	Player struct {
		PublicAddress string `json:"publicAddress"`
		Title         string `json:"title"`
		UUID          string `json:"uuid"`
		Local         bool   `json:"local"`
	} `json:"Player"`
	Account struct {
		Thumb string `json:"thumb"`
		Title string `json:"title"`
		ID    int    `json:"id"`
	} `json:"Account"`
	Metadata WebhookMetadata `json:"Metadata"`
	Rating   float64         `json:"rating"`
	User     bool            `json:"user"`
	Owner    bool            `json:"owner"`
}

// WebhookMetadata is part of an IncomingWebhook.
type WebhookMetadata struct {
	LibrarySectionID      interface{} `json:"librarySectionID"` // int (plex) or string (session tracker)
	ExternalRating        interface{} `json:"Rating,omitempty"` // bullshit.
	LibrarySectionType    string      `json:"librarySectionType"`
	RatingKey             string      `json:"ratingKey"`
	ParentRatingKey       string      `json:"parentRatingKey"`
	GrandparentRatingKey  string      `json:"grandparentRatingKey"`
	Key                   string      `json:"key"`
	GUID                  string      `json:"guid"`
	ParentGUID            string      `json:"parentGuid"`
	GrandparentGUID       string      `json:"grandparentGuid"`
	Studio                string      `json:"studio"`
	Type                  string      `json:"type"`
	GrandParentTitle      string      `json:"grandparentTitle"`
	GrandparentKey        string      `json:"grandparentKey"`
	ParentKey             string      `json:"parentKey"`
	ParentTitle           string      `json:"parentTitle"`
	ParentThumb           string      `json:"parentThumb"`
	GrandparentThumb      string      `json:"grandparentThumb"`
	GrandparentArt        string      `json:"grandparentArt"`
	GrandparentTheme      string      `json:"grandparentTheme"`
	Title                 string      `json:"title"`
	TitleSort             string      `json:"titleSort"`
	LibrarySectionTitle   string      `json:"librarySectionTitle"`
	LibrarySectionKey     string      `json:"librarySectionKey"`
	ContentRating         string      `json:"contentRating"`
	Summary               string      `json:"summary"`
	Tagline               string      `json:"tagline"`
	Thumb                 string      `json:"thumb"`
	Art                   string      `json:"art"`
	OriginallyAvailableAt string      `json:"originallyAvailableAt"`
	AudienceRatingImage   string      `json:"audienceRatingImage"`
	PrimaryExtraKey       string      `json:"primaryExtraKey"`
	RatingImage           string      `json:"ratingImage"`
	GuID                  []*GUID     `json:"Guid"`
	ParentYear            int         `json:"parentYear"`
	ParentIndex           int64       `json:"parentIndex"`
	Index                 int64       `json:"index"`
	Rating                float64     `json:"rating"`
	AudienceRating        float64     `json:"audienceRating"`
	ViewOffset            float64     `json:"viewOffset"`
	LastViewedAt          int64       `json:"lastViewedAt"`
	Year                  int         `json:"year"`
	Duration              float64     `json:"duration"`
	AddedAt               int64       `json:"addedAt"`
	UpdatedAt             int64       `json:"updatedAt"`
}
