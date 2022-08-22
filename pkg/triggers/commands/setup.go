package commands

import (
	"fmt"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/cnfg"
)

// Errors produced by this file.
var (
	ErrNoCmd = fmt.Errorf("cmd provided without a command configured; fix it")
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
	Name    string        `json:"name" toml:"name" xml:"name" yaml:"name"`
	Command string        `json:"-" toml:"command" xml:"command" yaml:"command"`
	Shell   bool          `json:"shell" toml:"shell" xml:"shell" yaml:"shell"`
	Log     bool          `json:"log" toml:"log" xml:"log" yaml:"log"`
	Notify  bool          `json:"notify" toml:"notify" xml:"notify" yaml:"notify"`
	Timeout cnfg.Duration `json:"-" toml:"timeout" xml:"timeout" yaml:"timeout"`
	Hash    string        `json:"hash" toml:"-" xml:"-" yaml:"-"`
	fails   int
	runs    int
	output  string // last output logged
	lastRun time.Time
	mu      sync.RWMutex
	ch      chan website.EventType
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
func (c *Command) Run(event website.EventType) {
	if c.ch == nil {
		return
	}

	c.ch <- event
}

// List returns a list of active triggers that can be executed.
func (a *Action) List() []*Command {
	return a.cmd.cmdlist
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
	Runs       int    `json:"runs"`
	Fails      int    `json:"fails"`
	LastOutput string `json:"output"`
	LastRun    string `json:"last"`
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
		Runs:       c.runs,
		Fails:      c.fails,
		LastOutput: c.output,
		LastRun:    last,
	}
}

func (c *cmd) create() {
	for _, cmd := range c.cmdlist {
		cmd.ch = make(chan website.EventType, 1)

		c.Add(&common.Action{
			Name: common.TriggerName(fmt.Sprintf("Running Custom Command '%s'", cmd.Name)),
			Fn:   cmd.run,
			C:    cmd.ch,
		})
	}

	c.Printf("==> Custom Commands: %d provided", len(c.cmdlist))
}
