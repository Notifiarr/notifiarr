//nolint:tagliatelle
package plex

import (
	"fmt"
	"time"
)

/* This file contains all the types for Plex Sessions API response. */

// Session is a Plex json struct.
type Session struct {
	User                 User        `json:"User"`
	Player               Player      `json:"Player"`
	TranscodeSession     Transcode   `json:"TranscodeSession"`
	Added                int64       `json:"addedAt"`
	Art                  string      `json:"art"`
	AudienceRating       float64     `json:"audienceRating"`
	AudienceRatingImg    string      `json:"audienceRatingImage"`
	ContentRating        string      `json:"contentRating"`
	Duration             float64     `json:"duration"`
	GUID                 string      `json:"guid"`
	GrandparentArt       string      `json:"grandparentArt"`
	GrandparentGUID      string      `json:"grandparentGuid"`
	GrandparentKey       string      `json:"grandparentKey"`
	GrandparentRatingKey string      `json:"grandparentRatingKey"`
	GrandparentTheme     string      `json:"grandparentTheme"`
	GrandparentThumb     string      `json:"grandparentThumb"`
	GrandparentTitle     string      `json:"grandparentTitle"`
	Index                int         `json:"index"`
	Key                  string      `json:"key"`
	LastViewed           int64       `json:"lastViewedAt"`
	LibrarySectionID     string      `json:"librarySectionID"`
	LibrarySectionKey    string      `json:"librarySectionKey"`
	LibrarySectionTitle  string      `json:"librarySectionTitle"`
	OriginallyAvailable  string      `json:"originallyAvailableAt"`
	ParentGUID           string      `json:"parentGuid"`
	ParentIndex          int         `json:"parentIndex"`
	ParentKey            string      `json:"parentKey"`
	ParentRatingKey      string      `json:"parentRatingKey"`
	ParentThumb          string      `json:"parentThumb"`
	ParentTitle          string      `json:"parentTitle"`
	PrimaryExtraKey      string      `json:"primaryExtraKey"`
	ExternalRating       interface{} `json:"Rating,omitempty"` // bullshit.
	Rating               float64     `json:"rating"`
	RatingImage          string      `json:"ratingImage"`
	RatingKey            string      `json:"ratingKey"`
	SessionKey           string      `json:"sessionKey"`
	Studio               string      `json:"studio"`
	Summary              string      `json:"summary"`
	Thumb                string      `json:"thumb"`
	Title                string      `json:"title"`
	TitleSort            string      `json:"titleSort"`
	Type                 string      `json:"type"`
	Updated              int64       `json:"updatedAt"`
	ViewCount            int64       `json:"viewCount"`
	ViewOffset           float64     `json:"viewOffset"`
	Year                 int64       `json:"year"`
	Session              struct {
		Bandwidth int64  `json:"bandwidth"`
		ID        string `json:"id"`
		Location  string `json:"location"`
	} `json:"Session"`
	GuID  []*GUID  `json:"Guid,omitempty"`
	Media []*Media `json:"Media,omitempty"`
	/* Notifiarr does not need these. :shrug:
	Country  []*Country  `json:"Country"`
	Director []*Director `json:"Director"`
	Genre    []*Genre    `json:"Genre"`
	Producer []*Producer `json:"Producer"`
	Role     []*Role     `json:"Role"`
	Similar  []*Similar  `json:"Similar"`
	Writer   []*Writer   `json:"Writer"`
	*/
}

// User is part of a Plex Session.
type User struct {
	ID    string `json:"id"`
	Thumb string `json:"thumb"`
	Title string `json:"title"`
}

// Player is part of a Plex Session.
type Player struct {
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
	StateTime   structDur `json:"stateTime"` // this is not a plex item. We calculate this.
	Title       string    `json:"title"`
	UserID      int64     `json:"userID"`
	Vendor      string    `json:"vendor"`
	Version     string    `json:"version"`
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
	ID     string `json:"id"`
	Tag    string `json:"tag"`
}

// MediaStream is part of a Plex Session.
type MediaStream struct {
	AudioChannelLayout string  `json:"audioChannelLayout,omitempty"`
	BitDepth           int     `json:"bitDepth,omitempty"`
	Bitrate            float64 `json:"bitrate"`
	BitrateMode        string  `json:"bitrateMode,omitempty"`
	Channels           int     `json:"channels,omitempty"`
	ChromaLocation     string  `json:"chromaLocation,omitempty"`
	ChromaSubsampling  string  `json:"chromaSubsampling,omitempty"`
	Codec              string  `json:"codec"`
	CodedHeight        int64   `json:"codedHeight,omitempty"`
	CodedWidth         int64   `json:"codedWidth,omitempty"`
	ColorPrimaries     string  `json:"colorPrimaries,omitempty"`
	ColorTrc           string  `json:"colorTrc,omitempty"`
	Decision           string  `json:"decision"`
	Default            bool    `json:"default,omitempty"`
	DisplayTitle       string  `json:"displayTitle"`
	ExtDisplayTitle    string  `json:"extendedDisplayTitle"`
	FrameRate          float64 `json:"frameRate,omitempty"`
	HasScalingMatrix   bool    `json:"hasScalingMatrix,omitempty"`
	Height             int64   `json:"height,omitempty"`
	ID                 string  `json:"id"`
	Index              int     `json:"index"`
	Language           string  `json:"language,omitempty"`
	LanguageCode       string  `json:"languageCode,omitempty"`
	Level              int     `json:"level,omitempty"`
	Location           string  `json:"location"`
	Profile            string  `json:"profile"`
	RefFrames          int     `json:"refFrames,omitempty"`
	SamplingRate       int     `json:"samplingRate,omitempty"`
	ScanType           string  `json:"scanType,omitempty"`
	Selected           bool    `json:"selected,omitempty"`
	StreamType         int     `json:"streamType"`
	Width              int64   `json:"width,omitempty"`
	LanguageTag        string  `json:"languageTag,omitempty"`
}

// MediaPart is part of a Plex Session.
type MediaPart struct {
	AudioProfile    string         `json:"audioProfile"`
	Bitrate         float64        `json:"bitrate"`
	Container       string         `json:"container"`
	Decision        string         `json:"decision"`
	Duration        float64        `json:"duration"`
	File            string         `json:"file"`
	Height          int64          `json:"height"`
	ID              string         `json:"id"`
	Indexes         string         `json:"indexes"`
	Key             string         `json:"key"`
	Protocol        string         `json:"protocol"`
	Selected        bool           `json:"selected"`
	Size            int64          `json:"size"`
	StreamingOptmzd bool           `json:"optimizedForStreaming"`
	VideoProfile    string         `json:"videoProfile"`
	Width           int64          `json:"width"`
	Stream          []*MediaStream `json:"Stream"`
}

// Media is part of a Plex Session.
type Media struct {
	AspectRatio     string       `json:"aspectRatio"`
	AudioChannels   int          `json:"audioChannels"`
	AudioCodec      string       `json:"audioCodec"`
	AudioProfile    string       `json:"audioProfile"`
	Bitrate         float64      `json:"bitrate"`
	Container       string       `json:"container"`
	Duration        float64      `json:"duration"`
	Height          int64        `json:"height"`
	ID              string       `json:"id"`
	Protocol        string       `json:"protocol"`
	StreamingOptmzd bool         `json:"optimizedForStreaming"`
	VideoCodec      string       `json:"videoCodec"`
	VideoFrameRate  string       `json:"videoFrameRate"`
	VideoProfile    string       `json:"videoProfile"`
	VideoResolution string       `json:"videoResolution"`
	Width           int64        `json:"width"`
	Selected        bool         `json:"selected"`
	Part            []*MediaPart `json:"Part"`
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
	AudioChannels       int     `json:"audioChannels"`
	AudioCodec          string  `json:"audioCodec"`
	AudioDecision       string  `json:"audioDecision"`
	Container           string  `json:"container"`
	Context             string  `json:"context"`
	Duration            int64   `json:"duration"`
	Key                 string  `json:"key"`
	MaxOffsetAvailable  float64 `json:"maxOffsetAvailable"`
	MinOffsetAvailable  float64 `json:"minOffsetAvailable"`
	Progress            float64 `json:"progress"`
	Protocol            string  `json:"protocol"`
	Remaining           int64   `json:"remaining"`
	Size                int64   `json:"size"`
	SourceAudioCodec    string  `json:"sourceAudioCodec"`
	SourceVideoCodec    string  `json:"sourceVideoCodec"`
	Speed               float64 `json:"speed"`
	TimeStamp           float64 `json:"timeStamp"`
	VideoCodec          string  `json:"videoCodec"`
	VideoDecision       string  `json:"videoDecision"`
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
