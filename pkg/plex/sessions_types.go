package plex

/* This file contains all the types for Plex Sessions API response. */

type Session struct {
	User                User      `json:"User"`
	Player              Player    `json:"Player"`
	TranscodeSession    Transcode `json:"TranscodeSession"`
	Added               string    `json:"addedAt"`
	Art                 string    `json:"art"`
	AudienceRating      string    `json:"audienceRating"`
	AudienceRatingImg   string    `json:"audienceRatingImage"`
	Duration            float64   `json:"duration,string"`
	ViewOffset          float64   `json:"viewOffset,string"`
	GUID                string    `json:"guid"`
	Key                 string    `json:"key"`
	LastViewed          int64     `json:"lastViewedAt,string"`
	LibrarySectionID    string    `json:"librarySectionID"`
	LibrarySectionKey   string    `json:"librarySectionKey"`
	LibrarySectionTitle string    `json:"librarySectionTitle"`
	OriginallyAvailable string    `json:"originallyAvailableAt"`
	PrimaryExtraKey     string    `json:"primaryExtraKey"`
	Rating              string    `json:"rating"`
	RatingImage         string    `json:"ratingImage"`
	RatingKey           string    `json:"ratingKey"`
	SessionKey          string    `json:"sessionKey"`
	Studio              string    `json:"studio"`
	Summary             string    `json:"summary"`
	Thumb               string    `json:"thumb"`
	Title               string    `json:"title"`
	TitleSort           string    `json:"titleSort"`
	Type                string    `json:"type"`
	Updated             int64     `json:"updatedAt,string"`
	Year                string    `json:"year"`
	Session             struct {
		Bandwidth int64  `json:"bandwidth"`
		ID        string `json:"id"`
		Location  string `json:"location"`
	} `json:"Session"`
	Country  []*Country  `json:"Country"`
	Director []*Director `json:"Director"`
	Genre    []*Genre    `json:"Genre"`
	Media    []*Media    `json:"Media"`
	Producer []*Producer `json:"Producer"`
	Role     []*Role     `json:"Role"`
	Similar  []*Similar  `json:"Similar"`
	Writer   []*Writer   `json:"Writer"`
}

type User struct {
	ID    string `json:"id"`
	Thumb string `json:"thumb"`
	Title string `json:"title"`
}

type Player struct {
	Address     string `json:"address"`
	Device      string `json:"device"`
	MachineID   string `json:"machineIdentifier"`
	Model       string `json:"model"`
	Platform    string `json:"platform"`
	PlatformVer string `json:"platformVersion"`
	Product     string `json:"product"`
	Profile     string `json:"profile"`
	PublicAddr  string `json:"remotePublicAddress"`
	State       string `json:"state"`
	Title       string `json:"title"`
	UserID      int64  `json:"userID"`
	Vendor      string `json:"vendor"`
	Version     string `json:"version"`
	Relayed     bool   `json:"relayed"`
	Local       bool   `json:"local"`
	Secure      bool   `json:"secure"`
}

type Country struct {
	Count  string `json:"count"`
	Filter string `json:"filter"`
	ID     string `json:"id"`
	Tag    string `json:"tag"`
}

type Director struct {
	Filter string `json:"filter"`
	ID     string `json:"id"`
	Tag    string `json:"tag"`
}

type Genre struct {
	Count  string `json:"count"`
	Filter string `json:"filter"`
	ID     string `json:"id"`
	Tag    string `json:"tag"`
}

type MediaStream struct {
	Bitrate         float64 `json:"bitrate,string"`
	Codec           string  `json:"codec"`
	ColorPrimaries  string  `json:"colorPrimaries,omitempty"`
	ColorTrc        string  `json:"colorTrc,omitempty"`
	Decision        string  `json:"decision"`
	Default         string  `json:"default,omitempty"`
	DisplayTitle    string  `json:"displayTitle"`
	ExtDisplayTitle string  `json:"extendedDisplayTitle"`
	FrameRate       float64 `json:"frameRate,omitempty,string"`
	Height          int64   `json:"height,omitempty,string"`
	ID              string  `json:"id"`
	Location        string  `json:"location"`
	StreamType      string  `json:"streamType"`
	Width           int64   `json:"width,omitempty,string"`
	BitrateMode     string  `json:"bitrateMode,omitempty"`
	Channels        string  `json:"channels,omitempty"`
	Language        string  `json:"language,omitempty"`
	LanguageCode    string  `json:"languageCode,omitempty"`
	Selected        string  `json:"selected,omitempty"`
}

type MediaPart struct {
	Stream          []*MediaStream `json:"Stream"`
	Bitrate         int64          `json:"bitrate,string"`
	Container       string         `json:"container"`
	Decision        string         `json:"decision"`
	Duration        float64        `json:"duration,string"`
	Height          int64          `json:"height,string"`
	ID              string         `json:"id"`
	Indexes         string         `json:"indexes"`
	StreamingOptmzd string         `json:"optimizedForStreaming"`
	Protocol        string         `json:"protocol"`
	VideoProfile    string         `json:"videoProfile"`
	Width           int64          `json:"width,string"`
	Selected        bool           `json:"selected"`
}

type Media struct {
	Part            []*MediaPart `json:"Part"`
	AudioChannels   int          `json:"audioChannels,string"`
	AudioCodec      string       `json:"audioCodec"`
	Bitrate         int64        `json:"bitrate,string"`
	Container       string       `json:"container"`
	Duration        float64      `json:"duration,string"`
	Height          int64        `json:"height,string"`
	ID              string       `json:"id"`
	StreamingOptmzd string       `json:"optimizedForStreaming"`
	Protocol        string       `json:"protocol"`
	VideoCodec      string       `json:"videoCodec"`
	VideoFrameRate  string       `json:"videoFrameRate"`
	VideoProfile    string       `json:"videoProfile"`
	VideoResolution string       `json:"videoResolution"`
	Width           int64        `json:"width,string"`
	Selected        bool         `json:"selected"`
}

type Producer struct {
	Filter string `json:"filter"`
	ID     string `json:"id"`
	Tag    string `json:"tag"`
}

type Role struct {
	Count  string `json:"count,omitempty"`
	Filter string `json:"filter"`
	ID     string `json:"id"`
	Role   string `json:"role"`
	Tag    string `json:"tag"`
	Thumb  string `json:"thumb,omitempty"`
}

type Transcode struct {
	AudioChannels       int    `json:"audioChannels"`
	AudioCodec          string `json:"audioCodec"`
	AudioDecision       string `json:"audioDecision"`
	Container           string `json:"container"`
	Context             string `json:"context"`
	Duration            int64  `json:"duration"`
	Key                 string `json:"key"`
	MaxOffsetAvailable  string `json:"maxOffsetAvailable"`
	MinOffsetAvailable  string `json:"minOffsetAvailable"`
	Progress            string `json:"progress"`
	Protocol            string `json:"protocol"`
	Remaining           int64  `json:"remaining"`
	Size                int64  `json:"size"`
	SourceAudioCodec    string `json:"sourceAudioCodec"`
	SourceVideoCodec    string `json:"sourceVideoCodec"`
	Speed               string `json:"speed"`
	TimeStamp           string `json:"timeStamp"`
	VideoCodec          string `json:"videoCodec"`
	VideoDecision       string `json:"videoDecision"`
	Throttled           bool   `json:"throttled"`
	Complete            bool   `json:"complete"`
	XcodeHwFullPipeline bool   `json:"transcodeHwFullPipeline"`
	XcodeHwRequested    bool   `json:"transcodeHwRequested"`
}

type Writer struct {
	Filter string `json:"filter"`
	ID     string `json:"id"`
	Tag    string `json:"tag"`
}

type Similar struct {
	Filter string `json:"filter"`
	ID     string `json:"id"`
	Tag    string `json:"tag"`
	Count  string `json:"count,omitempty"`
}
