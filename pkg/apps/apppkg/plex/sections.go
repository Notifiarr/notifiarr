//nolint:tagliatelle
package plex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
)

// LibrarySection is data about a library's section, ie movies, tv, music.
type LibrarySection struct {
	TrashSize        int    `json:"trashSize,omitempty"` // added later, not part of payload.
	AllowSync        bool   `json:"allowSync"`
	Art              string `json:"art"`
	Composite        string `json:"composite"`
	Filters          bool   `json:"filters"`
	Refreshing       bool   `json:"refreshing"`
	Thumb            string `json:"thumb"`
	Key              string `json:"key"` // this is the ID.
	Type             string `json:"type"`
	Title            string `json:"title"`
	Agent            string `json:"agent"`
	Scanner          string `json:"scanner"`
	Language         string `json:"language"`
	UUID             string `json:"uuid"`
	UpdatedAt        int    `json:"updatedAt"`
	CreatedAt        int    `json:"createdAt"`
	ScannedAt        int    `json:"scannedAt"`
	Content          bool   `json:"content"`
	Directory        bool   `json:"directory"`
	ContentChangedAt int    `json:"contentChangedAt"`
	Hidden           int    `json:"hidden"`
	Location         []struct {
		ID   int64  `json:"id"`
		Path string `json:"path"`
	} `json:"Location"`
}

// SectionDirectory contains the directory of sections.
type SectionDirectory struct {
	Size      int               `json:"size"`
	AllowSync bool              `json:"allowSync"`
	Title1    string            `json:"title1"`
	Directory []*LibrarySection `json:"Directory"`
}

// MediaSection is a plex response struct.
type MediaSection struct {
	Size                int    `json:"size"`
	AllowSync           bool   `json:"allowSync"`
	Identifier          string `json:"identifier"`
	LibrarySectionID    int    `json:"librarySectionID"`
	LibrarySectionTitle string `json:"librarySectionTitle"`
	LibrarySectionUUID  string `json:"librarySectionUUID"`
	MediaTagPrefix      string `json:"mediaTagPrefix"`
	MediaTagVersion     int    `json:"mediaTagVersion"`
	Metadata            []struct {
		RatingKey             string  `json:"ratingKey"`
		Key                   string  `json:"key"`
		ParentRatingKey       string  `json:"parentRatingKey,omitempty"`
		GrandparentRatingKey  string  `json:"grandparentRatingKey,omitempty"`
		GUID                  string  `json:"guid"`
		ParentGUID            string  `json:"parentGuid,omitempty"`
		GrandparentGUID       string  `json:"grandparentGuid,omitempty"`
		Type                  string  `json:"type"`
		Title                 string  `json:"title"`
		GrandparentKey        string  `json:"grandparentKey,omitempty"`
		ParentKey             string  `json:"parentKey,omitempty"`
		LibrarySectionTitle   string  `json:"librarySectionTitle"`
		LibrarySectionID      int     `json:"librarySectionID"`
		LibrarySectionKey     string  `json:"librarySectionKey"`
		GrandparentTitle      string  `json:"grandparentTitle,omitempty"`
		ParentTitle           string  `json:"parentTitle,omitempty"`
		ContentRating         string  `json:"contentRating"`
		Summary               string  `json:"summary"`
		Index                 int     `json:"index,omitempty"`
		ParentIndex           int     `json:"parentIndex,omitempty"`
		Rating                float64 `json:"rating,omitempty"`
		Year                  int     `json:"year,omitempty"`
		Thumb                 string  `json:"thumb"`
		Art                   string  `json:"art"`
		ParentThumb           string  `json:"parentThumb,omitempty"`
		GrandparentThumb      string  `json:"grandparentThumb,omitempty"`
		GrandparentArt        string  `json:"grandparentArt,omitempty"`
		GrandparentTheme      string  `json:"grandparentTheme,omitempty"`
		Duration              int     `json:"duration"`
		OriginallyAvailableAt string  `json:"originallyAvailableAt"`
		AddedAt               int     `json:"addedAt"`
		UpdatedAt             int     `json:"updatedAt"`
		Media                 []struct {
			ID                    int     `json:"id"`
			Duration              int     `json:"duration"`
			Bitrate               int     `json:"bitrate"`
			Width                 int     `json:"width"`
			Height                int     `json:"height"`
			AspectRatio           float64 `json:"aspectRatio"`
			AudioChannels         int     `json:"audioChannels"`
			AudioCodec            string  `json:"audioCodec"`
			VideoCodec            string  `json:"videoCodec"`
			VideoResolution       string  `json:"videoResolution"`
			Container             string  `json:"container"`
			VideoFrameRate        string  `json:"videoFrameRate"`
			OptimizedForStreaming int     `json:"optimizedForStreaming"`
			AudioProfile          string  `json:"audioProfile"`
			Has64BitOffsets       bool    `json:"has64bitOffsets"`
			VideoProfile          string  `json:"videoProfile"`
			Part                  []struct {
				ID                    int    `json:"id"`
				Key                   string `json:"key"`
				Duration              int    `json:"duration"`
				File                  string `json:"file"`
				Size                  int    `json:"size"`
				AudioProfile          string `json:"audioProfile"`
				Container             string `json:"container"`
				Indexes               string `json:"indexes"`
				VideoProfile          string `json:"videoProfile"`
				OptimizedForStreaming bool   `json:"optimizedForStreaming"`
				Has64BitOffsets       bool   `json:"has64bitOffsets"`
				Stream                []struct {
					ID                   int     `json:"id"`
					StreamType           int     `json:"streamType"`
					Codec                string  `json:"codec"`
					Index                int     `json:"index"`
					Bitrate              int     `json:"bitrate"`
					BitDepth             int     `json:"bitDepth,omitempty"`
					ChromaLocation       string  `json:"chromaLocation,omitempty"`
					ChromaSubsampling    string  `json:"chromaSubsampling,omitempty"`
					CodedHeight          int     `json:"codedHeight,omitempty"`
					CodedWidth           int     `json:"codedWidth,omitempty"`
					ColorRange           string  `json:"colorRange,omitempty"`
					ColorSpace           string  `json:"colorSpace,omitempty"`
					FrameRate            float64 `json:"frameRate,omitempty"`
					Height               int     `json:"height,omitempty"`
					Level                int     `json:"level,omitempty"`
					Profile              string  `json:"profile"`
					RefFrames            int     `json:"refFrames,omitempty"`
					StreamIdentifier     string  `json:"streamIdentifier"`
					Width                int     `json:"width,omitempty"`
					DisplayTitle         string  `json:"displayTitle"`
					ExtendedDisplayTitle string  `json:"extendedDisplayTitle"`
					Channels             int     `json:"channels,omitempty"`
					Language             string  `json:"language,omitempty"`
					LanguageCode         string  `json:"languageCode,omitempty"`
					AudioChannelLayout   string  `json:"audioChannelLayout,omitempty"`
					SamplingRate         int     `json:"samplingRate,omitempty"`
					Selected             bool    `json:"selected,omitempty"`
					HasScalingMatrix     bool    `json:"hasScalingMatrix,omitempty"`
					Default              bool    `json:"default"`
				} `json:"Stream"`
			} `json:"Part"`
		} `json:"Media"`
		TitleSort           string    `json:"titleSort,omitempty"`
		ViewOffset          int       `json:"viewOffset,omitempty"`
		LastViewedAt        int       `json:"lastViewedAt,omitempty"`
		ParentYear          int       `json:"parentYear,omitempty"`
		Studio              string    `json:"studio,omitempty"`
		AudienceRating      float64   `json:"audienceRating,omitempty"`
		ViewCount           int       `json:"viewCount,omitempty"`
		Tagline             string    `json:"tagline,omitempty"`
		AudienceRatingImage string    `json:"audienceRatingImage,omitempty"`
		ChapterSource       string    `json:"chapterSource,omitempty"`
		PrimaryExtraKey     string    `json:"primaryExtraKey,omitempty"`
		RatingImage         string    `json:"ratingImage,omitempty"`
		GuID                []*GUID   `json:"Guid,omitempty"`
		ExternalRating      []*Rating `json:"Rating,omitempty"`
		/* Notifiarr does not need these. :shrug:
		Country             []*Country  `json:"Country"`
		Director            []*Director `json:"Director"`
		Genre               []*Genre    `json:"Genre"`
		Producer            []*Producer `json:"Producer"`
		Role                []*Role     `json:"Role"`
		Similar             []*Similar  `json:"Similar"`
		Writer              []*Writer   `json:"Writer"`
		*/
	} `json:"Metadata"`
}

// GUID is a reusable type from the Section library.
type GUID struct {
	ID string `json:"id"`
}

// GetPlexSectionKey gets a section key from Plex based on a key path.
func (s *Server) GetPlexSectionKey(keyPath string) (*MediaSection, error) {
	return s.GetPlexSectionKeyWithContext(context.Background(), keyPath)
}

// GetPlexSectionKey gets a section key from Plex based on a key path.
func (s *Server) GetPlexSectionKeyWithContext(ctx context.Context, keyPath string) (*MediaSection, error) {
	url := s.config.URL + keyPath

	data, err := s.getPlexURL(ctx, url, nil)
	if err != nil {
		return nil, err
	}

	var output struct {
		MediaContainer *MediaSection `json:"MediaContainer"`
	}

	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("parsing library section from %s: %w; failed payload: %s", url, err, string(data))
	}

	return output.MediaContainer, nil
}

// GetDirectory returns data about all the library sections.
func (s *Server) GetDirectory() (*SectionDirectory, error) {
	return s.GetDirectoryWithContext(context.Background())
}

// GetDirectoryWithContext returns data about all the library sections.
func (s *Server) GetDirectoryWithContext(ctx context.Context) (*SectionDirectory, error) {
	url := s.config.URL + "/library/sections"

	data, err := s.getPlexURL(ctx, url, nil)
	if err != nil {
		return nil, err
	}

	var output struct {
		MediaContainer *SectionDirectory `json:"MediaContainer"`
	}

	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("unmarshaling directory from %s: %w; failed payload: %s", url, err, string(data))
	}

	return output.MediaContainer, nil
}

// GetDirectoryWithContext returns data about all the library sections.
func (s *Server) GetDirectoryTrashSizeWithContext(ctx context.Context, key string) (int, error) {
	uri := s.config.URL + path.Join("/library", "sections", key, "all")
	params := make(url.Values)
	params.Set("trash", "1")
	params.Set("episode.trash", "1")

	data, err := s.getPlexURL(ctx, uri, params)
	if err != nil {
		return 0, err
	}

	var output struct {
		MediaContainer *SectionDirectory `json:"MediaContainer"`
	}

	if err := json.Unmarshal(data, &output); err != nil {
		return 0, fmt.Errorf("unmarshaling trash directory from %s: %w; failed payload: %s", uri, err, string(data))
	}

	return output.MediaContainer.Size, nil
}
