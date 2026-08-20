package cfsync

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"golift.io/starr"
	"golift.io/starr/sonarr"
)

const (
	TrigCFSyncSportarr    common.TriggerName = "Starting Sportarr profile and format sync."
	TrigCFSyncSportarrInt common.TriggerName = "Starting Sportarr %d profile and format sync."
)

// SportarrTrashPayload is the payload sent and received
// to/from notifarr.com when updating custom formats for Sportarr.
type SportarrTrashPayload struct {
	Instance           int                          `json:"instance"`
	Name               string                       `json:"name"`
	ReleaseProfiles    []*sonarr.ReleaseProfile     `json:"releaseProfiles,omitempty"`
	QualityProfiles    []*sonarr.QualityProfile     `json:"qualityProfiles,omitempty"`
	CustomFormats      []*sonarr.CustomFormatOutput `json:"customFormats,omitempty"`
	QualityDefinitions []*sonarr.QualityDefinition  `json:"qualityDefinitions,omitempty"`
	Naming             *sonarr.Naming               `json:"naming"`
	Error              string                       `json:"error"`
}

// SyncSportarrRP initializes a release profile sync with sportarr.
func (a *Action) SyncSportarrRP(input *common.ActionInput) {
	if !a.cmd.Exec(input, TrigCFSyncSportarr) {
		mnd.Log.Errorf(input.ReqID,
			"[%s requested] Cannot sync Sportarr profiles and formats. No Sportarr instances configured.", input.Type)
	}
}

// SyncSportarrInstanceRP initializes a release profile sync with a specific sportarr instance.
func (a *Action) SyncSportarrInstanceRP(input *common.ActionInput, instance int) error {
	if name := TrigCFSyncSportarrInt.WithInstance(instance); !a.cmd.Exec(input, name) {
		return fmt.Errorf("%w: Sportarr instance: %d", common.ErrInvalidApp, instance)
	}

	return nil
}

// syncSportarr triggers a custom format sync for Sportarr.
func (c *cmd) syncSportarr(ctx context.Context, input *common.ActionInput) {
	info := clientinfo.Get()
	if info == nil || len(info.Actions.Sync.SportarrInstances) < 1 {
		mnd.Log.Printf(input.ReqID,
			"[%s requested] Cannot sync Sportarr profiles and formats. Website provided 0 instances.", input.Type)
		return
	} else if len(c.Apps.Sportarr) < 1 {
		mnd.Log.Printf(input.ReqID,
			"[%s requested] Cannot sync Sportarr profiles and formats. No Sportarr instances configured.", input.Type)
		return
	}

	for idx, app := range c.Apps.Sportarr {
		instance := idx + 1
		if !app.Enabled() || !info.Actions.Sync.SportarrInstances.Has(instance) {
			mnd.Log.Printf(input.ReqID,
				"[%s requested] Profiles and formats sync skipping Sportarr instance %d. Not in sync list: %v",
				input.Type, instance, info.Actions.Sync.SportarrInstances)
			continue
		}

		(&sportarrApp{app: &app, cmd: c, idx: idx}).syncSportarr(ctx, input)
	}
}

// syncSportarr sends the profiles for a single instance.
func (c *sportarrApp) syncSportarr(ctx context.Context, input *common.ActionInput) {
	start := time.Now()
	payload := c.cmd.getSportarrProfiles(ctx, input.Type, c.idx+1)
	website.SendData(&website.Request{
		ReqID:   input.ReqID,
		Route:   website.CFSyncRoute,
		Event:   input.Type,
		Params:  []string{"app=sportarr"},
		Payload: payload,
		LogMsg: fmt.Sprintf("Sportarr profiles and formats sync (elapsed: %v)",
			time.Since(start).Round(time.Millisecond)),
		LogPayload: true,
	})
	mnd.Log.Printf(input.ReqID, "[%s requested] Synced profiles and formats for Sportarr instance %d (%s/%s)",
		input.Type, c.idx+1, c.app.Name, c.app.URL)
}

func (c *cmd) getSportarrProfiles(ctx context.Context, event website.EventType, instance int) *SportarrTrashPayload {
	reqID := mnd.Log.Trace(mnd.GetID(ctx), "start: getSportarrProfiles", event, instance)
	defer mnd.Log.Trace(reqID, "end: getSportarrProfiles", event, instance)

	var (
		err     error
		app     = c.Config.Apps.Sportarr[instance-1]
		payload = SportarrTrashPayload{Instance: instance, Name: app.Name}
	)

	payload.QualityProfiles, err = app.GetQualityProfilesContext(ctx)
	if err != nil {
		errStr := fmt.Sprintf("getting quality profiles: %v ", err)
		payload.Error += errStr
		mnd.Log.Errorf(reqID, "[%s requested] Getting Sportarr data from instance %d (%s): %v",
			event, instance, app.Name, errStr)
	}

	payload.ReleaseProfiles, err = app.GetReleaseProfilesContext(ctx)
	if err != nil {
		errStr := fmt.Sprintf("getting release profiles: %v ", err)
		payload.Error += errStr
		mnd.Log.Errorf(reqID, "[%s requested] Getting Sportarr data from instance %d (%s): %v",
			event, instance, app.Name, errStr)
	}

	payload.QualityDefinitions, err = app.GetQualityDefinitionsContext(ctx)
	if err != nil {
		errStr := fmt.Sprintf("getting quality definitions: %v ", err)
		payload.Error += errStr
		mnd.Log.Errorf(reqID, "[%s requested] Getting Sportarr data from instance %d (%s): %v",
			event, instance, app.Name, errStr)
	}

	payload.CustomFormats, err = app.GetCustomFormatsContext(ctx)
	if err != nil && !errors.Is(err, starr.ErrInvalidStatusCode) {
		errStr := fmt.Sprintf("getting custom formats: %v ", err)
		payload.Error += errStr
		mnd.Log.Errorf(reqID, "[%s requested] Getting Sportarr data from instance %d (%s): %v",
			event, instance, app.Name, errStr)
	} else if errors.Is(err, starr.ErrInvalidStatusCode) {
		// This error is required so the site knows it speaks the sonarr v3 api.
		errStr := fmt.Sprintf("getting custom formats: %v ", err)
		payload.Error += errStr
	}

	payload.Naming, err = app.GetNamingContext(ctx)
	if err != nil {
		errStr := fmt.Sprintf("getting naming: %v ", err)
		payload.Error += errStr
		mnd.Log.Errorf(reqID, "[%s requested] Getting Sportarr data from instance %d (%s): %v",
			event, instance, app.Name, errStr)
	}

	return &payload
}

// aggregateTrashSportarr is fired by the api handler.
func (c *cmd) aggregateTrashSportarr(
	ctx context.Context,
	wait *sync.WaitGroup,
	instances clientinfo.IntList,
) []*SportarrTrashPayload {
	reqID := mnd.Log.Trace(mnd.GetID(ctx), "start: aggregateTrashSportarr", instances)
	defer mnd.Log.Trace(reqID, "end: aggregateTrashSportarr", instances)

	output := []*SportarrTrashPayload{}
	event := website.EventAPI

	// Create our known+requested instances, so we can write slice values in go routines.
	for idx, app := range c.Config.Apps.Sportarr {
		if instance := idx + 1; instances.Has(instance) {
			if app.Enabled() {
				output = append(output, &SportarrTrashPayload{Instance: instance, Name: app.Name})
			} else {
				mnd.Log.Errorf(reqID, "[%s requested] Profiles and formats aggregate for disabled Sportarr instance %d (%s)",
					event, instance, app.Name)
			}
		}
	}

	// Grab data for each requested instance in parallel/go routine.
	for idx := range output {
		if c.Config.Apps.Serial {
			output[idx] = c.getSportarrProfiles(ctx, event, output[idx].Instance)
			continue
		}

		wait.Add(1)

		go func(idx int) {
			output[idx] = c.getSportarrProfiles(ctx, event, output[idx].Instance)
			wait.Done() //nolint:wsl
		}(idx)
	}

	return output
}
