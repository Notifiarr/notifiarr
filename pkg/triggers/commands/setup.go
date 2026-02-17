package commands

import (
	"errors"
	"regexp"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/commands/cmdconfig"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
)

// Errors produced by this file.
var (
	ErrNoCmd = errors.New("cmd provided without a command configured; fix it")
)

const defaultTimeout = 15 * time.Second

// Action contains the exported methods for this package.
type Action struct {
	cmd *cmd
}

type cmd struct {
	*common.Config
	cmdlist []*Command
}

// Command contains the input data for a defined command.
// It also contains some saved data about the command being run.
type Command struct {
	cmdconfig.Config
	cmd     string
	args    []*regexp.Regexp
	disable bool
	fails   int
	runs    int
	output  string // last output logged
	lastRun time.Time
	lastArg []string
	lastCmd string
	mu      sync.RWMutex
	ch      chan *common.ActionInput
}

// New configures the library.
func New(config *common.Config, commands []*Command) *Action {
	for _, cmd := range commands {
		cmd.Setup()
	}

	return &Action{cmd: &cmd{Config: config, cmdlist: commands}}
}

// Run fires a custom command.
func (c *Command) Run(input *common.ActionInput) {
	if c.ch == nil {
		return
	}

	c.ch <- input
}

// List returns a list of active triggers that can be executed.
func (a *Action) List() []*cmdconfig.Config {
	output := make([]*cmdconfig.Config, 0, len(a.cmd.cmdlist))

	for _, c := range a.cmd.cmdlist {
		config := c.Config
		output = append(output, &config)
	}

	return output
}

// GetByHash returns a command by the hash ID.
func (a *Action) GetByHash(hash string) *Command {
	for _, cmd := range a.cmd.cmdlist {
		if cmd.Hash == hash {
			return cmd
		}
	}

	return nil
}

// Create initializes the library.
func (a *Action) Create() {
	reqID := mnd.ReqID()
	a.cmd.create(reqID)
}

// Stats for a command's invocations.
type Stats struct {
	Args       []*regexp.Regexp `json:"-"`
	Command    string           `json:"-"`
	Runs       int              `json:"runs"`
	Fails      int              `json:"fails"`
	LastOutput string           `json:"output"`
	LastRun    string           `json:"last"`
	LastCmd    string           `json:"lastCmd"`
	LastTime   time.Time        `json:"lastTime"`
	LastArgs   []string         `json:"lastArgs"`
}

// Stats returns statistics about a command.
func (c *Command) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	last := time.Since(c.lastRun).Round(time.Second).String()
	if c.lastRun.IsZero() {
		last = "never"
	}

	return Stats{
		Args:       c.args,
		Command:    c.cmd,
		Runs:       c.runs,
		Fails:      c.fails,
		LastOutput: c.output,
		LastRun:    last,
		LastTime:   c.lastRun,
		LastArgs:   c.lastArg,
		LastCmd:    c.lastCmd,
	}
}

func (c *cmd) create(reqID string) {
	for _, cmd := range c.cmdlist {
		cmd.ch = make(chan *common.ActionInput, 1)

		c.Add(&common.Action{
			Key:  "TrigCustomCommand",
			Name: common.TriggerName(cmd.Name),
			Fn:   cmd.run,
			C:    cmd.ch,
		})
	}

	mnd.Log.Printf(reqID, "==> Custom Commands: %d provided", len(c.cmdlist))
}
