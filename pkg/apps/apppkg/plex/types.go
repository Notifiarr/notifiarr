//nolint:tagliatelle
package plex

import (
	"fmt"
	"time"
)

/* This file contains all the types for Plex Sessions API response. */

// Session is a Plex json struct.
type Session struct {
	User                 User   `json:"User"`
	Art                  string `json:"art"`
	AudienceRatingImg    string `json:"audienceRatingImage"`
	ContentRating        string `json:"contentRating"`
	GUID                 string `json:"guid"`
	GrandparentArt       string `json:"grandparentArt"`
	GrandparentGUID      string `json:"grandparentGuid"`
	GrandparentKey       string `json:"grandparentKey"`
	GrandparentRatingKey string `json:"grandparentRatingKey"`
	GrandparentTheme     string `json:"grandparentTheme"`
	GrandparentThumb     string `json:"grandparentThumb"`
	GrandparentTitle     string `json:"grandparentTitle"`
	Key                  string `json:"key"`
	LibrarySectionID     string `json:"librarySectionID"`
	LibrarySectionKey    string `json:"librarySectionKey"`
	LibrarySectionTitle  string `json:"librarySectionTitle"`
	OriginallyAvailable  string `json:"originallyAvailableAt"`
	ParentGUID           string `json:"parentGuid"`
	ParentKey            string `json:"parentKey"`
	ParentRatingKey      string `json:"parentRatingKey"`
	ParentThumb          string `json:"parentThumb"`
	ParentTitle          string `json:"parentTitle"`
	PrimaryExtraKey      string `json:"primaryExtraKey"`
	RatingImage          string `json:"ratingImage"`
	RatingKey            string `json:"ratingKey"`
	SessionKey           string `json:"sessionKey"`
	Studio               string `json:"studio"`
	Summary              string `json:"summary"`
	Thumb                string `json:"thumb"`
	Title                string `json:"title"`
	TitleSort            string `json:"titleSort"`
	Type                 string `json:"type"`
	Session              struct {
		ID        string `json:"id"`
		Location  string `json:"location"`
		Bandwidth int64  `json:"bandwidth"`
	} `json:"Session"`
	GuID             []*GUID   `json:"Guid,omitempty"`
	Media            []*Media  `json:"Media,omitempty"`
	ExternalRating   []*Rating `json:"Rating,omitempty"`
	Player           Player    `json:"Player"`
	TranscodeSession Transcode `json:"TranscodeSession"`
	/* Notifiarr does not need these. :shrug:
	Country  []*Country  `json:"Country"`
	Director []*Director `json:"Director"`
	Genre    []*Genre    `json:"Genre"`
	Producer []*Producer `json:"Producer"`
	Role     []*Role     `json:"Role"`
	Similar  []*Similar  `json:"Similar"`
	Writer   []*Writer   `json:"Writer"`
	*/
	Added          int64   `json:"addedAt"`
	AudienceRating float64 `json:"audienceRating"`
	Duration       float64 `json:"duration"`
	Index          int64   `json:"index"`
	LastViewed     int64   `json:"lastViewedAt"`
	ParentIndex    int64   `json:"parentIndex"`
	Rating         float64 `json:"rating"`
	Updated        int64   `json:"updatedAt"`
	ViewCount      int64   `json:"viewCount"`
	ViewOffset     float64 `json:"viewOffset"`
	Year           int     `json:"year"`
}

// User is part of a Plex Session.
type User struct {
	ID    string `json:"id"`
	Thumb string `json:"thumb"`
	Title string `json:"title"`
}

// Rating is part of Plex metadata.
type Rating struct {
	Image string      `json:"image"`
	Value interface{} `json:"value"`
	Type  string      `json:"type"`
}

// Player is part of a Plex Session.
type Player struct {
	StateTime   structDur `json:"stateTime"` // this is not a plex item. We calculate this.
	Address     string    `json:"address"`
	Device      string    `json:"device"`
	MachineID   string    `json:"machineIdentifier"`
	Model       string    `json:"model"`
	Platform    string    `json:"platform"`
	PlatformVer string    `json:"platformVersion"`
	Product     string    `json:"product"`
	Profile     string    `json:"profile"`
	PublicAddr  string    `json:"remotePublicAddress"`
	State       string    `json:"state"`
	Title       string    `json:"title"`
	Vendor      string    `json:"vendor"`
	Version     string    `json:"version"`
	UserID      int64     `json:"userID"`
	Relayed     bool      `json:"relayed"`
	Local       bool      `json:"local"`
	Secure      bool      `json:"secure"`
}

type structDur struct {
	time.Time
}

func (s *structDur) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`%.0f`, time.Since(s.Time).Seconds())), nil
}

// Country is part of a Plex Session.
type Country struct {
	Count  string `json:"count"`
	Filter string `json:"filter"`
	ID     string `json:"id"`
	Tag    string `json:"tag"`
}

// Director is part of a Plex Session.
type Director struct {
	Filter string `json:"filter"`
	ID     string `json:"id"`
	Tag    string `json:"tag"`
}

// Genre is part of a Plex Session.
type Genre struct {
	Count  string `json:"count"`
	Filter string `json:"filter"`
	Tag    string `json:"tag"`
	ID     int64  `json:"id"`
}

// MediaStream is part of a Plex Session.
type MediaStream struct {
	AudioChannelLayout string  `json:"audioChannelLayout,omitempty"`
	BitrateMode        string  `json:"bitrateMode,omitempty"`
	ChromaLocation     string  `json:"chromaLocation,omitempty"`
	ChromaSubsampling  string  `json:"chromaSubsampling,omitempty"`
	Codec              string  `json:"codec"`
	ColorPrimaries     string  `json:"colorPrimaries,omitempty"`
	ColorTrc           string  `json:"colorTrc,omitempty"`
	Decision           string  `json:"decision"`
	DisplayTitle       string  `json:"displayTitle"`
	ExtDisplayTitle    string  `json:"extendedDisplayTitle"`
	ID                 string  `json:"id"`
	Language           string  `json:"language,omitempty"`
	LanguageCode       string  `json:"languageCode,omitempty"`
	Location           string  `json:"location"`
	Profile            string  `json:"profile"`
	ScanType           string  `json:"scanType,omitempty"`
	LanguageTag        string  `json:"languageTag,omitempty"`
	BitDepth           int     `json:"bitDepth,omitempty"`
	Bitrate            float64 `json:"bitrate"`
	Channels           int     `json:"channels,omitempty"`
	CodedHeight        int64   `json:"codedHeight,omitempty"`
	CodedWidth         int64   `json:"codedWidth,omitempty"`
	FrameRate          float64 `json:"frameRate,omitempty"`
	Height             int64   `json:"height,omitempty"`
	Index              int     `json:"index"`
	Level              int     `json:"level,omitempty"`
	RefFrames          int     `json:"refFrames,omitempty"`
	SamplingRate       int     `json:"samplingRate,omitempty"`
	StreamType         int     `json:"streamType"`
	Width              int64   `json:"width,omitempty"`
	Default            bool    `json:"default,omitempty"`
	HasScalingMatrix   bool    `json:"hasScalingMatrix,omitempty"`
	Selected           bool    `json:"selected,omitempty"`
}

// MediaPart is part of a Plex Session.
type MediaPart struct {
	AudioProfile    string         `json:"audioProfile"`
	Container       string         `json:"container"`
	Decision        string         `json:"decision"`
	File            string         `json:"file"`
	ID              string         `json:"id"`
	Indexes         string         `json:"indexes"`
	Key             string         `json:"key"`
	Protocol        string         `json:"protocol"`
	VideoProfile    string         `json:"videoProfile"`
	Stream          []*MediaStream `json:"Stream"`
	Bitrate         float64        `json:"bitrate"`
	Duration        float64        `json:"duration"`
	Height          int64          `json:"height"`
	Size            int64          `json:"size"`
	Width           int64          `json:"width"`
	Selected        bool           `json:"selected"`
	StreamingOptmzd bool           `json:"optimizedForStreaming"`
}

// Media is part of a Plex Session.
type Media struct {
	AspectRatio     string       `json:"aspectRatio"`
	AudioCodec      string       `json:"audioCodec"`
	AudioProfile    string       `json:"audioProfile"`
	Container       string       `json:"container"`
	ID              string       `json:"id"`
	Protocol        string       `json:"protocol"`
	VideoCodec      string       `json:"videoCodec"`
	VideoFrameRate  string       `json:"videoFrameRate"`
	VideoProfile    string       `json:"videoProfile"`
	VideoResolution string       `json:"videoResolution"`
	Part            []*MediaPart `json:"Part"`
	AudioChannels   int          `json:"audioChannels"`
	Bitrate         float64      `json:"bitrate"`
	Duration        float64      `json:"duration"`
	Height          int64        `json:"height"`
	Width           int64        `json:"width"`
	StreamingOptmzd bool         `json:"optimizedForStreaming"`
	Selected        bool         `json:"selected"`
}

// Producer is part of a Plex Session.
type Producer struct {
	Count  string `json:"count"`
	Filter string `json:"filter"`
	ID     string `json:"id"`
	Tag    string `json:"tag"`
}

// Role is part of a Plex Session.
type Role struct {
	Count  string `json:"count,omitempty"`
	Filter string `json:"filter"`
	ID     string `json:"id"`
	Role   string `json:"role"`
	Tag    string `json:"tag"`
	Thumb  string `json:"thumb,omitempty"`
}

// Transcode is part of a Plex Session.
type Transcode struct {
	AudioCodec          string  `json:"audioCodec"`
	AudioDecision       string  `json:"audioDecision"`
	Container           string  `json:"container"`
	Context             string  `json:"context"`
	Key                 string  `json:"key"`
	Protocol            string  `json:"protocol"`
	SourceAudioCodec    string  `json:"sourceAudioCodec"`
	SourceVideoCodec    string  `json:"sourceVideoCodec"`
	VideoCodec          string  `json:"videoCodec"`
	VideoDecision       string  `json:"videoDecision"`
	AudioChannels       int     `json:"audioChannels"`
	Duration            int64   `json:"duration"`
	MaxOffsetAvailable  float64 `json:"maxOffsetAvailable"`
	MinOffsetAvailable  float64 `json:"minOffsetAvailable"`
	Progress            float64 `json:"progress"`
	Remaining           int64   `json:"remaining"`
	Size                int64   `json:"size"`
	Speed               float64 `json:"speed"`
	TimeStamp           float64 `json:"timeStamp"`
	Throttled           bool    `json:"throttled"`
	Complete            bool    `json:"complete"`
	XcodeHwFullPipeline bool    `json:"transcodeHwFullPipeline"`
	XcodeHwRequested    bool    `json:"transcodeHwRequested"`
}

// Writer is part of a Plex Session.
type Writer struct {
	Filter string `json:"filter"`
	ID     string `json:"id"`
	Tag    string `json:"tag"`
}

// Similar is part of a Plex Session.
type Similar struct {
	Filter string `json:"filter"`
	ID     string `json:"id"`
	Tag    string `json:"tag"`
	Count  string `json:"count,omitempty"`
}
