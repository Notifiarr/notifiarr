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
	Art       string `json:"art"`
	Composite string `json:"composite"`
	Thumb     string `json:"thumb"`
	Key       string `json:"key"` // this is the ID.
	Type      string `json:"type"`
	Title     string `json:"title"`
	Agent     string `json:"agent"`
	Scanner   string `json:"scanner"`
	Language  string `json:"language"`
	UUID      string `json:"uuid"`
	Location  []struct {
		Path string `json:"path"`
		ID   int64  `json:"id"`
	} `json:"Location"`
	TrashSize        int  `json:"trashSize,omitempty"` // added later, not part of payload.
	UpdatedAt        int  `json:"updatedAt"`
	CreatedAt        int  `json:"createdAt"`
	ScannedAt        int  `json:"scannedAt"`
	ContentChangedAt int  `json:"contentChangedAt"`
	Hidden           int  `json:"hidden"`
	AllowSync        bool `json:"allowSync"`
	Filters          bool `json:"filters"`
	Refreshing       bool `json:"refreshing"`
	Content          bool `json:"content"`
	Directory        bool `json:"directory"`
}

// SectionDirectory contains the directory of sections.
type SectionDirectory struct {
	Title1    string            `json:"title1"`
	Directory []*LibrarySection `json:"Directory"`
	Size      int               `json:"size"`
	AllowSync bool              `json:"allowSync"`
}

// MediaSection is a plex response struct.
type MediaSection struct {
	Identifier          string `json:"identifier"`
	LibrarySectionTitle string `json:"librarySectionTitle"`
	LibrarySectionUUID  string `json:"librarySectionUUID"`
	MediaTagPrefix      string `json:"mediaTagPrefix"`
	Metadata            []struct {
		RatingKey             string `json:"ratingKey"`
		Key                   string `json:"key"`
		ParentRatingKey       string `json:"parentRatingKey,omitempty"`
		GrandparentRatingKey  string `json:"grandparentRatingKey,omitempty"`
		GUID                  string `json:"guid"`
		ParentGUID            string `json:"parentGuid,omitempty"`
		GrandparentGUID       string `json:"grandparentGuid,omitempty"`
		Type                  string `json:"type"`
		Title                 string `json:"title"`
		GrandparentKey        string `json:"grandparentKey,omitempty"`
		ParentKey             string `json:"parentKey,omitempty"`
		LibrarySectionTitle   string `json:"librarySectionTitle"`
		LibrarySectionKey     string `json:"librarySectionKey"`
		GrandparentTitle      string `json:"grandparentTitle,omitempty"`
		ParentTitle           string `json:"parentTitle,omitempty"`
		ContentRating         string `json:"contentRating"`
		Summary               string `json:"summary"`
		Thumb                 string `json:"thumb"`
		Art                   string `json:"art"`
		ParentThumb           string `json:"parentThumb,omitempty"`
		GrandparentThumb      string `json:"grandparentThumb,omitempty"`
		GrandparentArt        string `json:"grandparentArt,omitempty"`
		GrandparentTheme      string `json:"grandparentTheme,omitempty"`
		OriginallyAvailableAt string `json:"originallyAvailableAt"`
		TitleSort             string `json:"titleSort,omitempty"`
		Studio                string `json:"studio,omitempty"`
		Tagline               string `json:"tagline,omitempty"`
		AudienceRatingImage   string `json:"audienceRatingImage,omitempty"`
		ChapterSource         string `json:"chapterSource,omitempty"`
		PrimaryExtraKey       string `json:"primaryExtraKey,omitempty"`
		RatingImage           string `json:"ratingImage,omitempty"`
		Media                 []struct {
			AudioCodec      string `json:"audioCodec"`
			VideoCodec      string `json:"videoCodec"`
			VideoResolution string `json:"videoResolution"`
			Container       string `json:"container"`
			VideoFrameRate  string `json:"videoFrameRate"`
			AudioProfile    string `json:"audioProfile"`
			VideoProfile    string `json:"videoProfile"`
			Part            []struct {
				Key          string `json:"key"`
				File         string `json:"file"`
				AudioProfile string `json:"audioProfile"`
				Container    string `json:"container"`
				Indexes      string `json:"indexes"`
				VideoProfile string `json:"videoProfile"`
				Stream       []struct {
					Codec                string  `json:"codec"`
					ChromaLocation       string  `json:"chromaLocation,omitempty"`
					ChromaSubsampling    string  `json:"chromaSubsampling,omitempty"`
					ColorRange           string  `json:"colorRange,omitempty"`
					ColorSpace           string  `json:"colorSpace,omitempty"`
					Profile              string  `json:"profile"`
					StreamIdentifier     string  `json:"streamIdentifier"`
					DisplayTitle         string  `json:"displayTitle"`
					ExtendedDisplayTitle string  `json:"extendedDisplayTitle"`
					Language             string  `json:"language,omitempty"`
					LanguageCode         string  `json:"languageCode,omitempty"`
					AudioChannelLayout   string  `json:"audioChannelLayout,omitempty"`
					ID                   int     `json:"id"`
					StreamType           int     `json:"streamType"`
					Index                int     `json:"index"`
					Bitrate              int     `json:"bitrate"`
					BitDepth             int     `json:"bitDepth,omitempty"`
					CodedHeight          int     `json:"codedHeight,omitempty"`
					CodedWidth           int     `json:"codedWidth,omitempty"`
					FrameRate            float64 `json:"frameRate,omitempty"`
					Height               int     `json:"height,omitempty"`
					Level                int     `json:"level,omitempty"`
					RefFrames            int     `json:"refFrames,omitempty"`
					Width                int     `json:"width,omitempty"`
					Channels             int     `json:"channels,omitempty"`
					SamplingRate         int     `json:"samplingRate,omitempty"`
					Selected             bool    `json:"selected,omitempty"`
					HasScalingMatrix     bool    `json:"hasScalingMatrix,omitempty"`
					Default              bool    `json:"default"`
				} `json:"Stream"`
				ID                    int  `json:"id"`
				Duration              int  `json:"duration"`
				Size                  int  `json:"size"`
				OptimizedForStreaming bool `json:"optimizedForStreaming"`
				Has64BitOffsets       bool `json:"has64bitOffsets"`
			} `json:"Part"`
			ID                    int     `json:"id"`
			Duration              int     `json:"duration"`
			Bitrate               int     `json:"bitrate"`
			Width                 int     `json:"width"`
			Height                int     `json:"height"`
			AspectRatio           float64 `json:"aspectRatio"`
			AudioChannels         int     `json:"audioChannels"`
			OptimizedForStreaming int     `json:"optimizedForStreaming"`
			Has64BitOffsets       bool    `json:"has64bitOffsets"`
		} `json:"Media"`
		GuID           []*GUID   `json:"Guid,omitempty"`
		ExternalRating []*Rating `json:"Rating,omitempty"`
		/* Notifiarr does not need these. :shrug:
		Country             []*Country  `json:"Country"`
		Director            []*Director `json:"Director"`
		Genre               []*Genre    `json:"Genre"`
		Producer            []*Producer `json:"Producer"`
		Role                []*Role     `json:"Role"`
		Similar             []*Similar  `json:"Similar"`
		Writer              []*Writer   `json:"Writer"`
		*/
		LibrarySectionID int     `json:"librarySectionID"`
		Index            int     `json:"index,omitempty"`
		ParentIndex      int     `json:"parentIndex,omitempty"`
		Rating           float64 `json:"rating,omitempty"`
		Year             int     `json:"year,omitempty"`
		Duration         int     `json:"duration"`
		AddedAt          int     `json:"addedAt"`
		UpdatedAt        int     `json:"updatedAt"`
		ViewOffset       int     `json:"viewOffset,omitempty"`
		LastViewedAt     int     `json:"lastViewedAt,omitempty"`
		ParentYear       int     `json:"parentYear,omitempty"`
		AudienceRating   float64 `json:"audienceRating,omitempty"`
		ViewCount        int     `json:"viewCount,omitempty"`
	} `json:"Metadata"`
	Size             int  `json:"size"`
	LibrarySectionID int  `json:"librarySectionID"`
	MediaTagVersion  int  `json:"mediaTagVersion"`
	AllowSync        bool `json:"allowSync"`
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
