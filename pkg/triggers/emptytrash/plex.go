package emptytrash

import (
	"context"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

/* Gaps allows filling gaps in Radarr collections. */

const TrigPlexEmptyTrash common.TriggerName = "Emptying Plex Trash"

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
func (a *Action) Plex(event website.EventType, libraryKey string) {
	a.cmd.Exec(&common.ActionInput{Type: event, Args: []string{libraryKey}}, TrigPlexEmptyTrash)
}

func (c *cmd) emptyPlexTrash(ctx context.Context, input *common.ActionInput) {
	_, err := c.Apps.Plex.EmptyTrashWithContext(ctx, input.Args[0])
	if err != nil {
		c.Errorf("[%s requested] Emptying Plex trash for library '%s' failed: %v", input.Type, input.Args[0], err)
		return
	}

	c.Printf("[%s requested] Emptied Plex library '%s' trash.", input.Type, input.Args[0])
}
