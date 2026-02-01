package emptytrash

import (
	"context"
	"fmt"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

const TrigPlexEmptyTrash common.TriggerName = "Emptying Plex Trash."

// Action contains the exported methods for this package.
type Action struct {
	cmd *cmd
}

type cmd struct {
	*common.Config
}

// New configures the library.
func New(config *common.Config) *Action {
	return &Action{cmd: &cmd{Config: config}}
}

// Create initializes the library.
func (a *Action) Create() {
	a.cmd.create()
}

func (c *cmd) create() {
	c.Add(&common.Action{
		Key:  "TrigPlexEmptyTrash",
		Name: TrigPlexEmptyTrash,
		Fn:   c.emptyPlexTrash,
		C:    make(chan *common.ActionInput, 1),
	})
}

// Send radarr collection gaps to the website.
func (a *Action) Send(input *common.ActionInput) {
	a.cmd.Exec(input, TrigPlexEmptyTrash)
}

// Plex empties the trash for a Library in Plex.
func (a *Action) Plex(input *common.ActionInput, libraryKeys []string) {
	input.Args = libraryKeys
	a.cmd.Exec(input, TrigPlexEmptyTrash)
}

func (c *cmd) emptyPlexTrash(ctx context.Context, input *common.ActionInput) {
	status := make(map[string]string)
	errors := 0

	for _, key := range input.Args {
		if _, err := c.Apps.Plex.EmptyTrashWithContext(ctx, key); err != nil {
			mnd.Log.ErrorfNoShare("{trace:%s} [%s requested] Emptying Plex trash for library '%s' failed: %v",
				input.ReqID, input.Type, key, err)

			status[key] = err.Error()
			errors++
		} else {
			status[key] = "ok"
		}
	}

	if len(status) > 0 {
		website.SendData(&website.Request{
			ReqID:      input.ReqID,
			Route:      website.PlexRoute,
			Event:      input.Type,
			Params:     []string{"emptylibrary=true"},
			Payload:    status,
			LogMsg:     fmt.Sprintf("Emptied %d Plex library trashes with %d errors.", len(status), errors),
			LogPayload: true,
		})
	} else {
		mnd.Log.Printf("{trace:%s} [%s requested] Emptied %d Plex library trashes with %d errors.",
			input.ReqID, input.Type, len(status), errors)
	}
}
