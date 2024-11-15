//nolint:dupl
package cfsync

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"golift.io/starr/radarr"
)

const (
	TrigCFSyncRadarr    common.TriggerName = "Starting Radarr profile and format sync."
	TrigCFSyncRadarrInt common.TriggerName = "Starting Radarr %d profile and format sync."
)

// RadarrTrashPayload is the payload sent and received
// to/from notifarr.com when updating custom formats for Radarr.
// This is used in other places, like the trash API handler in the 'client' module.
type RadarrTrashPayload struct {
	Instance           int                          `json:"instance"`
	Name               string                       `json:"name"`
	CustomFormats      []*radarr.CustomFormatOutput `json:"customFormats,omitempty"`
	QualityProfiles    []*radarr.QualityProfile     `json:"qualityProfiles,omitempty"`
	QualityDefinitions []*radarr.QualityDefinition  `json:"qualityDefinitions,omitempty"`
	Naming             *radarr.Naming               `json:"naming"`
	Error              string                       `json:"error"`
}

// SyncRadarrCF initializes a custom format sync with radarr.
func (a *Action) SyncRadarrCF(event website.EventType) {
	a.cmd.Exec(&common.ActionInput{Type: event}, TrigCFSyncRadarr)
}

// SyncRadarrInstanceCF initializes a custom format sync with a specific radarr instance.
func (a *Action) SyncRadarrInstanceCF(event website.EventType, instance int) error {
	if name := TrigCFSyncRadarrInt.WithInstance(instance); !a.cmd.Exec(&common.ActionInput{Type: event}, name) {
		return fmt.Errorf("%w: Radarr instance: %d", common.ErrInvalidApp, instance)
	}

	return nil
}

// syncRadarr triggers a custom format sync for Radarr.
func (c *cmd) syncRadarr(ctx context.Context, input *common.ActionInput) {
	info := clientinfo.Get()
	if info == nil || len(info.Actions.Sync.RadarrInstances) < 1 {
		c.Printf("[%s requested] Cannot sync Radarr profiles and formats. Website provided 0 instances.", input.Type)
		return
	} else if len(c.Apps.Radarr) < 1 {
		c.Printf("[%s requested] Cannot sync Radarr profiles and formats. No Radarr instances configured.", input.Type)
		return
	}

	for idx, app := range c.Apps.Radarr {
		instance := idx + 1
		if !app.Enabled() || !info.Actions.Sync.RadarrInstances.Has(instance) {
			c.Printf("[%s requested] Profiles and formats sync skipping Radarr instance %d. Not in sync list: %v",
				input.Type, instance, info.Actions.Sync.RadarrInstances)
			continue
		}

		(&radarrApp{app: app, cmd: c, idx: idx}).syncRadarr(ctx, input)
	}
}

// syncRadarr sends the profiles for a single instance.
func (c *radarrApp) syncRadarr(ctx context.Context, input *common.ActionInput) {
	start := time.Now()
	payload := c.cmd.getRadarrProfiles(ctx, input.Type, c.idx+1)

	c.cmd.SendData(&website.Request{
		Route:      website.CFSyncRoute,
		Event:      input.Type,
		Params:     []string{"app=radarr"},
		Payload:    payload,
		LogMsg:     fmt.Sprintf("Radarr profiles and formats sync (elapsed: %v)", time.Since(start).Round(time.Millisecond)),
		LogPayload: true,
	})
	c.cmd.Printf("[%s requested] Synced profiles and formats for Radarr instance %d (%s/%s)",
		input.Type, c.idx+1, c.app.Name, c.app.URL)
}

func (c *cmd) getRadarrProfiles(ctx context.Context, event website.EventType, instance int) *RadarrTrashPayload {
	var (
		err     error
		app     = c.Config.Apps.Radarr[instance-1]
		payload = RadarrTrashPayload{Instance: instance, Name: app.Name}
	)

	payload.QualityProfiles, err = app.GetQualityProfilesContext(ctx)
	if err != nil {
		errStr := fmt.Sprintf("getting quality profiles: %v ", err)
		payload.Error += errStr
		c.Errorf("[%s requested] Getting Radarr data from instance %d (%s): %v", event, instance, app.Name, errStr)
	}

	payload.CustomFormats, err = app.GetCustomFormatsContext(ctx)
	if err != nil {
		errStr := fmt.Sprintf("getting custom formats: %v ", err)
		payload.Error += errStr
		c.Errorf("[%s requested] Getting Radarr data from instance %d (%s): %v", event, instance, app.Name, errStr)
	}

	payload.QualityDefinitions, err = app.GetQualityDefinitionsContext(ctx)
	if err != nil {
		errStr := fmt.Sprintf("getting quality definitions: %v ", err)
		payload.Error += errStr
		c.Errorf("[%s requested] Getting Radarr data from instance %d (%s): %v", event, instance, app.Name, errStr)
	}

	payload.Naming, err = app.GetNamingContext(ctx)
	if err != nil {
		errStr := fmt.Sprintf("getting naming: %v ", err)
		payload.Error += errStr
		c.Errorf("[%s requested] Getting Radarr data from instance %d (%s): %v", event, instance, app.Name, errStr)
	}

	return &payload
}

// aggregateTrashRadarr is fired by the api handler.
func (c *cmd) aggregateTrashRadarr(
	ctx context.Context,
	wait *sync.WaitGroup,
	instances clientinfo.IntList,
) []*RadarrTrashPayload {
	output := []*RadarrTrashPayload{}
	event := website.EventAPI

	// Create our known+requested instances, so we can write slice values in go routines.
	for idx, app := range c.Config.Apps.Radarr {
		if instance := idx + 1; instances.Has(instance) {
			if app.Enabled() {
				output = append(output, &RadarrTrashPayload{Instance: instance, Name: app.Name})
			} else {
				c.Errorf("[%s requested] Profiles and formats aggregate for disabled Radarr instance %d (%s)",
					event, instance, app.Name)
			}
		}
	}

	// Grab data for each requested instance in parallel/go routine.
	for idx := range output {
		if c.Config.Apps.Serial {
			output[idx] = c.getRadarrProfiles(ctx, event, output[idx].Instance)
			continue
		}

		wait.Add(1)

		go func(idx int) {
			output[idx] = c.getRadarrProfiles(ctx, event, output[idx].Instance)
			wait.Done() //nolint:wsl
		}(idx)
	}

	return output
}
