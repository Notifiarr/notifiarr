package emptytrash

import (
	"context"
	"fmt"

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
func (a *Action) Plex(event website.EventType, libraryKeys []string) {
	a.cmd.Exec(&common.ActionInput{Type: event, Args: libraryKeys}, TrigPlexEmptyTrash)
}

func (c *cmd) emptyPlexTrash(ctx context.Context, input *common.ActionInput) {
	status := make(map[string]string)
	errors := 0

	for _, key := range input.Args {
		if _, err := c.Apps.Plex.EmptyTrashWithContext(ctx, key); err != nil {
			c.ErrorfNoShare("[%s requested] Emptying Plex trash for library '%s' failed: %v", input.Type, key, err)

			status[key] = err.Error()
			errors++
		} else {
			status[key] = "ok"
		}
	}

	if len(status) > 0 {
		c.SendData(&website.Request{
			Route:      website.PlexRoute,
			Event:      input.Type,
			Params:     []string{"emptylibrary=true"},
			Payload:    status,
			LogMsg:     fmt.Sprintf("Emptied %d Plex library trashes with %d errors.", len(status), errors),
			LogPayload: true,
		})
	} else {
		c.Printf("[%s requested] Emptied %d Plex library trashes with %d errors.", input.Type, len(status), errors)
	}
}
