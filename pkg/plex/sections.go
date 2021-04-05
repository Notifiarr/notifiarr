package plex

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type LibrarySection struct {
	Size                int    `json:"size"`
	AllowSync           bool   `json:"allowSync"`
	Identifier          string `json:"identifier"`
	LibrarySectionID    int    `json:"librarySectionID"`
	LibrarySectionTitle string `json:"librarySectionTitle"`
	LibrarySectionUUID  string `json:"librarySectionUUID"`
	MediaTagPrefix      string `json:"mediaTagPrefix"`
	MediaTagVersion     int    `json:"mediaTagVersion"`
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
		LibrarySectionID      int    `json:"librarySectionID"`
		LibrarySectionKey     string `json:"librarySectionKey"`
		GrandparentTitle      string `json:"grandparentTitle,omitempty"`
		ParentTitle           string `json:"parentTitle,omitempty"`
		ContentRating         string `json:"contentRating"`
		Summary               string `json:"summary"`
		Index                 int    `json:"index,omitempty"`
		ParentIndex           int    `json:"parentIndex,omitempty"`
		Rating                int    `json:"rating,omitempty"`
		Year                  int    `json:"year,omitempty"`
		Thumb                 string `json:"thumb"`
		Art                   string `json:"art"`
		ParentThumb           string `json:"parentThumb,omitempty"`
		GrandparentThumb      string `json:"grandparentThumb,omitempty"`
		GrandparentArt        string `json:"grandparentArt,omitempty"`
		GrandparentTheme      string `json:"grandparentTheme,omitempty"`
		Duration              int    `json:"duration"`
		OriginallyAvailableAt string `json:"originallyAvailableAt"`
		AddedAt               int    `json:"addedAt"`
		UpdatedAt             int    `json:"updatedAt"`
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
				Has64BitOffsets       bool   `json:"has64bitOffsets"`
				Indexes               string `json:"indexes"`
				OptimizedForStreaming bool   `json:"optimizedForStreaming"`
				VideoProfile          string `json:"videoProfile"`
				Stream                []struct {
					ID                   int     `json:"id"`
					StreamType           int     `json:"streamType"`
					Default              bool    `json:"default"`
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
					HasScalingMatrix     bool    `json:"hasScalingMatrix,omitempty"`
					Height               int     `json:"height,omitempty"`
					Level                int     `json:"level,omitempty"`
					Profile              string  `json:"profile"`
					RefFrames            int     `json:"refFrames,omitempty"`
					StreamIdentifier     string  `json:"streamIdentifier"`
					Width                int     `json:"width,omitempty"`
					DisplayTitle         string  `json:"displayTitle"`
					ExtendedDisplayTitle string  `json:"extendedDisplayTitle"`
					Selected             bool    `json:"selected,omitempty"`
					Channels             int     `json:"channels,omitempty"`
					Language             string  `json:"language,omitempty"`
					LanguageCode         string  `json:"languageCode,omitempty"`
					AudioChannelLayout   string  `json:"audioChannelLayout,omitempty"`
					SamplingRate         int     `json:"samplingRate,omitempty"`
				} `json:"Stream"`
			} `json:"Part"`
		} `json:"Media"`
		TitleSort           string      `json:"titleSort,omitempty"`
		ViewOffset          int         `json:"viewOffset,omitempty"`
		LastViewedAt        int         `json:"lastViewedAt,omitempty"`
		ParentYear          int         `json:"parentYear,omitempty"`
		Studio              string      `json:"studio,omitempty"`
		AudienceRating      float64     `json:"audienceRating,omitempty"`
		ViewCount           int         `json:"viewCount,omitempty"`
		Tagline             string      `json:"tagline,omitempty"`
		AudienceRatingImage string      `json:"audienceRatingImage,omitempty"`
		ChapterSource       string      `json:"chapterSource,omitempty"`
		PrimaryExtraKey     string      `json:"primaryExtraKey,omitempty"`
		RatingImage         string      `json:"ratingImage,omitempty"`
		GuID                []*GUID     `json:"Guid,omitempty"`
		Country             []*Country  `json:"Country"`
		Director            []*Director `json:"Director"`
		Genre               []*Genre    `json:"Genre"`
		Producer            []*Producer `json:"Producer"`
		Role                []*Role     `json:"Role"`
		Similar             []*Similar  `json:"Similar"`
		Writer              []*Writer   `json:"Writer"`
	} `json:"Metadata"`
}

type GUID struct {
	ID string `json:"id"`
}

func (s *Server) GetPlexSectionKey(keyPath string) (*LibrarySection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout.Duration)
	defer cancel()

	data, err := s.getPlexSectionKey(ctx, keyPath)
	if err != nil {
		return nil, err
	}

	var v struct {
		MediaContainer struct {
			Section *LibrarySection `json:"Metadata"`
		} `json:"MediaContainer"`
	}

	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("unmarshaling library section: %w", err)
	}

	return v.MediaContainer.Section, nil
}

func (s *Server) getPlexSectionKey(ctx context.Context, keyPath string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL+keyPath, nil)
	if err != nil {
		return nil, fmt.Errorf("creating http request: %w", err)
	}

	req.Header.Set("X-Plex-Token", s.Token)

	resp, err := s.getClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("making http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading http response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return body, ErrBadStatus
	}

	return body, nil
}
