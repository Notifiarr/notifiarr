package plex

import (
	"context"
	"encoding/json"
	"fmt"
)

func (s *Server) GetInfo() (*PMSInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout.Duration)
	defer cancel()

	data, err := s.getPlexURL(ctx, s.URL, nil)
	if err != nil {
		return nil, err
	}

	var v struct {
		MediaContainer *PMSInfo `json:"MediaContainer"`
	}

	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("unmarshaling main page from %s: %w", s.URL, err)
	}

	return v.MediaContainer, nil
}

// PMSInfo is the `/` path on Plex.
type PMSInfo struct {
	Size                          int64        `json:"size"`
	AllowCameraUpload             bool         `json:"allowCameraUpload"`
	AllowChannelAccess            bool         `json:"allowChannelAccess"`
	AllowSharing                  bool         `json:"allowSharing"`
	AllowSync                     bool         `json:"allowSync"`
	AllowTuners                   bool         `json:"allowTuners"`
	BackgroundProcessing          bool         `json:"backgroundProcessing"`
	Certificate                   bool         `json:"certificate"`
	CompanionProxy                bool         `json:"companionProxy"`
	CountryCode                   string       `json:"countryCode"`
	Diagnostics                   string       `json:"diagnostics"`
	EventStream                   bool         `json:"eventStream"`
	FriendlyName                  string       `json:"friendlyName"`
	HubSearch                     bool         `json:"hubSearch"`
	ItemClusters                  bool         `json:"itemClusters"`
	LiveTV                        int64        `json:"livetv"`
	MachineIdentifier             string       `json:"machineIdentifier"`
	MaxUploadBitrate              int64        `json:"maxUploadBitrate"`
	MaxUploadBitrateReason        string       `json:"maxUploadBitrateReason"`
	MaxUploadBitrateReasonMessage string       `json:"maxUploadBitrateReasonMessage"`
	MediaProviders                bool         `json:"mediaProviders"`
	Multiuser                     bool         `json:"multiuser"`
	MyPlex                        bool         `json:"myPlex"`
	MyPlexMappingState            string       `json:"myPlexMappingState"`
	MyPlexSigninState             string       `json:"myPlexSigninState"`
	MyPlexSubscription            bool         `json:"myPlexSubscription"`
	MyPlexUsername                string       `json:"myPlexUsername"`
	OfflineTranscode              int64        `json:"offlineTranscode"`
	OwnerFeatures                 string       `json:"ownerFeatures"`
	PhotoAutoTag                  bool         `json:"photoAutoTag"`
	Platform                      string       `json:"platform"`
	PlatformVersion               string       `json:"platformVersion"`
	PluginHost                    bool         `json:"pluginHost"`
	PushNotifications             bool         `json:"pushNotifications"`
	ReadOnlyLibraries             bool         `json:"readOnlyLibraries"`
	RequestParametersInCookie     bool         `json:"requestParametersInCookie"`
	StreamingBrainABRVersion      int64        `json:"streamingBrainABRVersion"`
	StreamingBrainVersion         int64        `json:"streamingBrainVersion"`
	Sync                          bool         `json:"sync"`
	TranscoderActiveVideoSessions int64        `json:"transcoderActiveVideoSessions"`
	TranscoderAudio               bool         `json:"transcoderAudio"`
	TranscoderLyrics              bool         `json:"transcoderLyrics"`
	TranscoderPhoto               bool         `json:"transcoderPhoto"`
	TranscoderSubtitles           bool         `json:"transcoderSubtitles"`
	TranscoderVideo               bool         `json:"transcoderVideo"`
	TranscoderVideoBitrates       string       `json:"transcoderVideoBitrates"`
	TranscoderVideoQualities      string       `json:"transcoderVideoQualities"`
	TranscoderVideoResolutions    string       `json:"transcoderVideoResolutions"`
	UpdatedAt                     int64        `json:"updatedAt"`
	Updater                       bool         `json:"updater"`
	Version                       string       `json:"version"`
	VoiceSearch                   bool         `json:"voiceSearch"`
	Directory                     []*Directory `json:"Directory"`
} // `json:"MediaContainer"`

// Directory is part of the PMSInfo.
type Directory struct {
	Count int    `json:"count"`
	Key   string `json:"key"`
	Title string `json:"title"`
}
