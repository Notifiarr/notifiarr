//nolint:tagliatelle
package plex

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetInfo retrieves Plex Server Info. This also sets the friendly name, so s.Name() works.
func (s *Server) GetInfo(ctx context.Context) (*PMSInfo, error) {
	data, err := s.getPlexURL(ctx, s.config.URL, nil)
	if err != nil {
		return nil, err
	}

	var output struct {
		MediaContainer *PMSInfo `json:"MediaContainer"`
	}

	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("unmarshaling main page from %s: %w", s.config.URL, err)
	}

	s.name = output.MediaContainer.FriendlyName

	return output.MediaContainer, nil
}

// PMSInfo is the `/` path on Plex.
type PMSInfo struct {
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
	Directory                     []*Directory `json:"Directory"`
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
	Size                          int64        `json:"size"`
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
} // `json:"MediaContainer"`

// Directory is part of the PMSInfo.
type Directory struct {
	Count int    `json:"count"`
	Key   string `json:"key"`
	Title string `json:"title"`
}
