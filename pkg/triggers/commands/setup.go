package commands

import (
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/commands/cmdconfig"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
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
	mu      sync.RWMutex
	ch      chan *common.ActionInput
	log     mnd.Logger
	website *website.Server
}

// New configures the library.
func New(config *common.Config, commands []*Command) *Action {
	for _, cmd := range commands {
		cmd.Setup(config.Logger, config.Server)
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
	output := []*cmdconfig.Config{}

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
	a.cmd.create()
}

// Stats for a command's invocations.
type Stats struct {
	Args       []*regexp.Regexp `json:"-"`
	Command    string           `json:"-"`
	Runs       int              `json:"runs"`
	Fails      int              `json:"fails"`
	LastOutput string           `json:"output"`
	LastRun    string           `json:"last"`
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
		LastArgs:   c.lastArg,
	}
}

func (c *cmd) create() {
	for _, cmd := range c.cmdlist {
		if err := cmd.SetupRegexpArgs(); err != nil {
			c.Errorf("Command Setup Failed: %v", err)
			cmd.disable = true //nolint:wsl
		}

		cmd.ch = make(chan *common.ActionInput, 1)

		c.Add(&common.Action{
			Name: common.TriggerName(fmt.Sprintf("Running Custom Command '%s'", cmd.Name)),
			Fn:   cmd.run,
			C:    cmd.ch,
		})
	}

	c.Printf("==> Custom Commands: %d provided", len(c.cmdlist))
}
