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
	CountryCode                   string       `json:"countryCode"`
	Diagnostics                   string       `json:"diagnostics"`
	FriendlyName                  string       `json:"friendlyName"`
	MachineIdentifier             string       `json:"machineIdentifier"`
	MaxUploadBitrateReason        string       `json:"maxUploadBitrateReason"`
	MaxUploadBitrateReasonMessage string       `json:"maxUploadBitrateReasonMessage"`
	MyPlexMappingState            string       `json:"myPlexMappingState"`
	MyPlexSigninState             string       `json:"myPlexSigninState"`
	MyPlexUsername                string       `json:"myPlexUsername"`
	OwnerFeatures                 string       `json:"ownerFeatures"`
	Platform                      string       `json:"platform"`
	PlatformVersion               string       `json:"platformVersion"`
	TranscoderVideoBitrates       string       `json:"transcoderVideoBitrates"`
	TranscoderVideoQualities      string       `json:"transcoderVideoQualities"`
	TranscoderVideoResolutions    string       `json:"transcoderVideoResolutions"`
	Version                       string       `json:"version"`
	Directory                     []*Directory `json:"Directory"`
	LiveTV                        int64        `json:"livetv"`
	MaxUploadBitrate              int64        `json:"maxUploadBitrate"`
	OfflineTranscode              int64        `json:"offlineTranscode"`
	Size                          int64        `json:"size"`
	StreamingBrainABRVersion      int64        `json:"streamingBrainABRVersion"`
	StreamingBrainVersion         int64        `json:"streamingBrainVersion"`
	TranscoderActiveVideoSessions int64        `json:"transcoderActiveVideoSessions"`
	UpdatedAt                     int64        `json:"updatedAt"`
	AllowCameraUpload             bool         `json:"allowCameraUpload"`
	AllowChannelAccess            bool         `json:"allowChannelAccess"`
	AllowSharing                  bool         `json:"allowSharing"`
	AllowSync                     bool         `json:"allowSync"`
	AllowTuners                   bool         `json:"allowTuners"`
	BackgroundProcessing          bool         `json:"backgroundProcessing"`
	Certificate                   bool         `json:"certificate"`
	CompanionProxy                bool         `json:"companionProxy"`
	EventStream                   bool         `json:"eventStream"`
	HubSearch                     bool         `json:"hubSearch"`
	ItemClusters                  bool         `json:"itemClusters"`
	MediaProviders                bool         `json:"mediaProviders"`
	Multiuser                     bool         `json:"multiuser"`
	MyPlex                        bool         `json:"myPlex"`
	MyPlexSubscription            bool         `json:"myPlexSubscription"`
	PhotoAutoTag                  bool         `json:"photoAutoTag"`
	PluginHost                    bool         `json:"pluginHost"`
	PushNotifications             bool         `json:"pushNotifications"`
	ReadOnlyLibraries             bool         `json:"readOnlyLibraries"`
	RequestParametersInCookie     bool         `json:"requestParametersInCookie"`
	Sync                          bool         `json:"sync"`
	TranscoderAudio               bool         `json:"transcoderAudio"`
	TranscoderLyrics              bool         `json:"transcoderLyrics"`
	TranscoderPhoto               bool         `json:"transcoderPhoto"`
	TranscoderSubtitles           bool         `json:"transcoderSubtitles"`
	TranscoderVideo               bool         `json:"transcoderVideo"`
	Updater                       bool         `json:"updater"`
	VoiceSearch                   bool         `json:"voiceSearch"`
} // `json:"MediaContainer"`

// Directory is part of the PMSInfo.
type Directory struct {
	Key   string `json:"key"`
	Title string `json:"title"`
	Count int    `json:"count"`
}
