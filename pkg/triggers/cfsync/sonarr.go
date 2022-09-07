package cfsync

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/starr"
	"golift.io/starr/sonarr"
)

const TrigRPSyncSonarr common.TriggerName = "Starting Sonarr Release Profile TRaSH sync."

// SonarrTrashPayload is the payload sent and received
// to/from notifarr.com when updating custom formats for Sonarr.
type SonarrTrashPayload struct {
	Instance           int                         `json:"instance"`
	Name               string                      `json:"name"`
	ReleaseProfiles    []*sonarr.ReleaseProfile    `json:"releaseProfiles,omitempty"`
	QualityProfiles    []*sonarr.QualityProfile    `json:"qualityProfiles,omitempty"`
	CustomFormats      []*sonarr.CustomFormat      `json:"customFormats,omitempty"`
	QualityDefinitions []*sonarr.QualityDefinition `json:"qualityDefinitions,omitempty"`
	Error              string                      `json:"error"`
}

// SyncSonarrRP initializes a release profile sync with sonarr.
func (a *Action) SyncSonarrRP(event website.EventType) {
	a.cmd.Exec(event, TrigRPSyncSonarr)
}

// syncSonarr triggers a custom format sync for Sonarr.
func (c *cmd) syncSonarr(event website.EventType) {
	if c.ClientInfo == nil || len(c.ClientInfo.Actions.Sync.SonarrInstances) < 1 {
		c.Debugf("[%s requested] Cannot sync Sonarr Release Profiles. Website provided 0 instances.", event)
		return
	} else if len(c.Apps.Sonarr) < 1 {
		c.Debugf("[%s requested] Cannot sync Sonarr Release Profiles. No Sonarr instances configured.", event)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), maxSyncTime)
	defer cancel()

	for i, app := range c.Apps.Sonarr {
		instance := i + 1
		if app.URL == "" || app.APIKey == "" || app.Timeout.Duration < 0 ||
			!c.ClientInfo.Actions.Sync.SonarrInstances.Has(instance) {
			c.Debugf("[%s requested] CF Sync Skipping Sonarr instance %d. Not in sync list: %v",
				event, instance, c.ClientInfo.Actions.Sync.SonarrInstances)
			continue
		}

		start := time.Now()
		payload := c.getSonarrProfiles(ctx, event, instance)
		c.SendData(&website.Request{
			Route:      website.CFSyncRoute,
			Event:      event,
			Params:     []string{"app=sonarr"},
			Payload:    payload,
			LogMsg:     fmt.Sprintf("Sonarr TRaSH Sync (elapsed: %v)", time.Since(start).Round(time.Millisecond)),
			LogPayload: true,
		})
		c.Printf("[%s requested] Synced Release Profiles for Sonarr instance %d (%s/%s)", event, instance, app.Name, app.URL)
	}
}

func (c *cmd) getSonarrProfiles(ctx context.Context, event website.EventType, instance int) *SonarrTrashPayload {
	var (
		err     error
		app     = c.Config.Apps.Sonarr[instance-1]
		payload = SonarrTrashPayload{Instance: instance, Name: app.Name}
	)

	payload.QualityProfiles, err = app.GetQualityProfilesContext(ctx)
	if err != nil {
		errStr := fmt.Sprintf("getting quality profiles: %v ", err)
		payload.Error += errStr
		c.Errorf("[%s requested] Getting Sonarr data from instance %d (%s): %v", event, instance, app.Name, errStr)
	}

	payload.ReleaseProfiles, err = app.GetReleaseProfilesContext(ctx)
	if err != nil {
		errStr := fmt.Sprintf("getting release profiles: %v ", err)
		payload.Error += errStr
		c.Errorf("[%s requested] Getting Sonarr data from instance %d (%s): %v", event, instance, app.Name, errStr)
	}

	payload.QualityDefinitions, err = app.GetQualityDefinitionsContext(ctx)
	if err != nil {
		errStr := fmt.Sprintf("getting quality definitions: %v ", err)
		payload.Error += errStr
		c.Errorf("[%s requested] Getting Sonarr data from instance %d (%s): %v", event, instance, app.Name, errStr)
	}

	payload.CustomFormats, err = app.GetCustomFormatsContext(ctx)
	if err != nil && !errors.Is(err, starr.ErrInvalidStatusCode) {
		errStr := fmt.Sprintf("getting custom formats: %v ", err)
		payload.Error += errStr
		c.Errorf("[%s requested] Getting Sonarr data from instance %d (%s): %v", event, instance, app.Name, errStr)
	} else if errors.Is(err, starr.ErrInvalidStatusCode) {
		// This error is required so the site knows it's sonarr v3.
		errStr := fmt.Sprintf("getting custom formats: %v ", err)
		payload.Error += errStr
	}

	return &payload
}

// aggregateTrashSonarr is fired by the api handler.
func (c *cmd) aggregateTrashSonarr(
	ctx context.Context,
	wait *sync.WaitGroup,
	instances website.IntList,
) []*SonarrTrashPayload {
	output := []*SonarrTrashPayload{}
	event := website.EventAPI

	// Create our known+requested instances, so we can write slice values in go routines.
	for idx, app := range c.Config.Apps.Sonarr {
		if instance := idx + 1; instances.Has(instance) && app.Enabled() {
			output = append(output, &SonarrTrashPayload{Instance: instance, Name: app.Name})
		} else {
			c.Errorf("[%s requested] Aggegregate request for disabled Sonarr instance %d (%s)", event, instance, app.Name)
		}
	}

	// Grab data for each requested instance in parallel/go routine.
	for idx := range output {
		if c.Config.Serial {
			output[idx] = c.getSonarrProfiles(ctx, event, output[idx].Instance)
			continue
		}

		wait.Add(1)

		go func(idx int) {
			output[idx] = c.getSonarrProfiles(ctx, event, output[idx].Instance)
			wait.Done() //nolint:wsl
		}(idx)
	}

	return output
}
