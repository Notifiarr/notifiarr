package plex

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
