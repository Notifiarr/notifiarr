package plex

// IncomingWebhook is the incoming webhook from a Plex Media Server.
//nolint:tagliatelle
type IncomingWebhook struct {
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
