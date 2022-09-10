//nolint:dupl
package cfsync

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/starr/radarr"
)

const TrigCFSyncRadarr common.TriggerName = "Starting Radarr Custom Format TRaSH sync."

// RadarrTrashPayload is the payload sent and received
// to/from notifarr.com when updating custom formats for Radarr.
// This is used in other places, like the trash API handler in the 'client' module.
type RadarrTrashPayload struct {
	Instance           int                         `json:"instance"`
	Name               string                      `json:"name"`
	CustomFormats      []*radarr.CustomFormat      `json:"customFormats,omitempty"`
	QualityProfiles    []*radarr.QualityProfile    `json:"qualityProfiles,omitempty"`
	QualityDefinitions []*radarr.QualityDefinition `json:"qualityDefinitions,omitempty"`
	Error              string                      `json:"error"`
}

// SyncRadarrCF initializes a custom format sync with radarr.
func (a *Action) SyncRadarrCF(event website.EventType) {
	a.cmd.Exec(event, TrigCFSyncRadarr)
}

// syncRadarr triggers a custom format sync for Radarr.
func (c *cmd) syncRadarr(event website.EventType) {
	if c.ClientInfo == nil || len(c.ClientInfo.Actions.Sync.RadarrInstances) < 1 {
		c.Debugf("[%s requested] Cannot sync Radarr Custom Formats. Website provided 0 instances.", event)
		return
	} else if len(c.Apps.Radarr) < 1 {
		c.Debugf("[%s requested] Cannot sync Radarr Custom Formats. No Radarr instances configured.", event)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), maxSyncTime)
	defer cancel()

	for i, app := range c.Apps.Radarr {
		instance := i + 1
		if !app.Enabled() || !c.ClientInfo.Actions.Sync.RadarrInstances.Has(instance) {
			c.Debugf("[%s requested] CF Sync Skipping Radarr instance %d. Not in sync list: %v",
				event, instance, c.ClientInfo.Actions.Sync.RadarrInstances)
			continue
		}

		start := time.Now()
		payload := c.getRadarrProfiles(ctx, event, instance)
		c.SendData(&website.Request{
			Route:      website.CFSyncRoute,
			Event:      event,
			Params:     []string{"app=radarr"},
			Payload:    payload,
			LogMsg:     fmt.Sprintf("Radarr TRaSH Sync (elapsed: %v)", time.Since(start).Round(time.Millisecond)),
			LogPayload: true,
		})
		c.Printf("[%s requested] Synced Custom Formats for Radarr instance %d (%s/%s)", event, instance, app.Name, app.URL)
	}
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

	return &payload
}

// aggregateTrashRadarr is fired by the api handler.
func (c *cmd) aggregateTrashRadarr(
	ctx context.Context,
	wait *sync.WaitGroup,
	instances website.IntList,
) []*RadarrTrashPayload {
	output := []*RadarrTrashPayload{}
	event := website.EventAPI

	// Create our known+requested instances, so we can write slice values in go routines.
	for idx, app := range c.Config.Apps.Radarr {
		if instance := idx + 1; instances.Has(instance) {
			if app.Enabled() {
				output = append(output, &RadarrTrashPayload{Instance: instance, Name: app.Name})
			} else {
				c.Errorf("[%s requested] Aggegregate request for disabled Radarr instance %d (%s)", event, instance, app.Name)
			}
		}
	}

	// Grab data for each requested instance in parallel/go routine.
	for idx := range output {
		if c.Config.Serial {
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
