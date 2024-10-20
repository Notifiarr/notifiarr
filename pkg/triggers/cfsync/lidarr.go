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
	"golift.io/starr/lidarr"
)

const (
	TrigCFSyncLidarr    common.TriggerName = "Starting Lidarr profile and format sync."
	TrigCFSyncLidarrInt common.TriggerName = "Starting Lidarr %d profile and format sync."
)

// LidarrTrashPayload is the payload sent and received
// to/from notifarr.com when updating custom formats for Lidarr.
// This is used in other places, like the trash API handler in the 'client' module.
type LidarrTrashPayload struct {
	Instance           int                          `json:"instance"`
	Name               string                       `json:"name"`
	CustomFormats      []*lidarr.CustomFormatOutput `json:"customFormats,omitempty"`
	QualityProfiles    []*lidarr.QualityProfile     `json:"qualityProfiles,omitempty"`
	QualityDefinitions []*lidarr.QualityDefinition  `json:"qualityDefinitions,omitempty"`
	Naming             *lidarr.Naming               `json:"naming"`
	Error              string                       `json:"error"`
}

// SyncLidarrCF initializes a custom format sync with lidarr.
func (a *Action) SyncLidarrCF(event website.EventType) {
	a.cmd.Exec(&common.ActionInput{Type: event}, TrigCFSyncLidarr)
}

// SyncLidarrInstanceCF initializes a custom format sync with a specific lidarr instance.
func (a *Action) SyncLidarrInstanceCF(event website.EventType, instance int) error {
	if name := TrigCFSyncLidarrInt.WithInstance(instance); !a.cmd.Exec(&common.ActionInput{Type: event}, name) {
		return fmt.Errorf("%w: Lidarr instance: %d", common.ErrInvalidApp, instance)
	}

	return nil
}

// syncLidarr triggers a custom format sync for Lidarr.
func (c *cmd) syncLidarr(ctx context.Context, input *common.ActionInput) {
	info := clientinfo.Get()
	if info == nil || len(info.Actions.Sync.LidarrInstances) < 1 {
		c.Printf("[%s requested] Cannot sync Lidarr profiles and formats. Website provided 0 instances.", input.Type)
		return
	} else if len(c.Apps.Lidarr) < 1 {
		c.Printf("[%s requested] Cannot sync Lidarr profiles and formats. No Lidarr instances configured.", input.Type)
		return
	}

	for idx, app := range c.Apps.Lidarr {
		instance := idx + 1
		if !app.Enabled() || !info.Actions.Sync.LidarrInstances.Has(instance) {
			c.Printf("[%s requested] Profiles and formats sync skipping Lidarr instance %d. Not in sync list: %v",
				input.Type, instance, info.Actions.Sync.LidarrInstances)
			continue
		}

		(&lidarrApp{app: app, cmd: c, idx: idx}).syncLidarr(ctx, input)
	}
}

// syncLidarr sends the profiles for a single instance.
func (c *lidarrApp) syncLidarr(ctx context.Context, input *common.ActionInput) {
	start := time.Now()
	payload := c.cmd.getLidarrProfiles(ctx, input.Type, c.idx+1)

	c.cmd.SendData(&website.Request{
		Route:      website.CFSyncRoute,
		Event:      input.Type,
		Params:     []string{"app=lidarr"},
		Payload:    payload,
		LogMsg:     fmt.Sprintf("Lidarr profiles and formats sync (elapsed: %v)", time.Since(start).Round(time.Millisecond)),
		LogPayload: true,
	})
	c.cmd.Printf("[%s requested] Synced profiles and formats for Lidarr instance %d (%s/%s)",
		input.Type, c.idx+1, c.app.Name, c.app.URL)
}

func (c *cmd) getLidarrProfiles(ctx context.Context, event website.EventType, instance int) *LidarrTrashPayload {
	var (
		err     error
		app     = c.Config.Apps.Lidarr[instance-1]
		payload = LidarrTrashPayload{Instance: instance, Name: app.Name}
	)

	payload.QualityProfiles, err = app.GetQualityProfilesContext(ctx)
	if err != nil {
		errStr := fmt.Sprintf("getting quality profiles: %v ", err)
		payload.Error += errStr
		c.Errorf("[%s requested] Getting Lidarr data from instance %d (%s): %v", event, instance, app.Name, errStr)
	}

	payload.CustomFormats, err = app.GetCustomFormatsContext(ctx)
	if err != nil {
		errStr := fmt.Sprintf("getting custom formats: %v ", err)
		payload.Error += errStr
		c.Errorf("[%s requested] Getting Lidarr data from instance %d (%s): %v", event, instance, app.Name, errStr)
	}

	payload.QualityDefinitions, err = app.GetQualityDefinitionsContext(ctx)
	if err != nil {
		errStr := fmt.Sprintf("getting quality definitions: %v ", err)
		payload.Error += errStr
		c.Errorf("[%s requested] Getting Lidarr data from instance %d (%s): %v", event, instance, app.Name, errStr)
	}

	payload.Naming, err = app.GetNamingContext(ctx)
	if err != nil {
		errStr := fmt.Sprintf("getting naming: %v ", err)
		payload.Error += errStr
		c.Errorf("[%s requested] Getting Lidarr data from instance %d (%s): %v", event, instance, app.Name, errStr)
	}

	return &payload
}

// aggregateTrashLidarr is fired by the api handler.
func (c *cmd) aggregateTrashLidarr(
	ctx context.Context,
	wait *sync.WaitGroup,
	instances clientinfo.IntList,
) []*LidarrTrashPayload {
	output := []*LidarrTrashPayload{}
	event := website.EventAPI

	// Create our known+requested instances, so we can write slice values in go routines.
	for idx, app := range c.Config.Apps.Lidarr {
		if instance := idx + 1; instances.Has(instance) {
			if app.Enabled() {
				output = append(output, &LidarrTrashPayload{Instance: instance, Name: app.Name})
			} else {
				c.Errorf("[%s requested] Profiles and formats aggregate for disabled Lidarr instance %d (%s)",
					event, instance, app.Name)
			}
		}
	}

	// Grab data for each requested instance in parallel/go routine.
	for idx := range output {
		if c.Config.Apps.Serial {
			output[idx] = c.getLidarrProfiles(ctx, event, output[idx].Instance)
			continue
		}

		wait.Add(1)

		go func(idx int) {
			output[idx] = c.getLidarrProfiles(ctx, event, output[idx].Instance)
			wait.Done() //nolint:wsl
		}(idx)
	}

	return output
}
