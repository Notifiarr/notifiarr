// package commands provides the interfaces and structures to trigger and run shell commands.
package commands

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"golift.io/cnfg"
)

// Errors produced by this file.
var (
	ErrNoCmd = fmt.Errorf("cmdhook without a command configured; fix it")
)

const defaultTimeout = 15 * time.Second

// Command container the input data for a defined command.
// It also contains some saved data about the command being run.
type Command struct {
	Name      string        `json:"name" toml:"name" xml:"name" yaml:"name"`
	Command   string        `json:"command" toml:"command" xml:"command" yaml:"command"`
	Shell     bool          `json:"shell" toml:"shell" xml:"shell" yaml:"shell"`
	Timeout   cnfg.Duration `json:"timeout" toml:"timeout" xml:"timeout" yaml:"timeout"`
	LogOutput bool          `json:"logOutput" toml:"log_output" xml:"log_output" yaml:"logOutput"`
	fails     int
	runs      int
	mu        sync.RWMutex
}

func (c *Command) Validate() error {
	if c.Command == "" {
		return ErrNoCmd
	}

	if c.Name == "" {
		c.Name = strings.Fields(c.Command)[0]
	}

	if c.Timeout.Duration == 0 {
		c.Timeout.Duration = defaultTimeout
	}

	return nil
}

// Stats for a command's invocations.
type Stats struct {
	Runs  int
	Fails int
}

// Stats returns statistics about a command.
func (c *Command) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return Stats{
		Runs:  (c.runs),
		Fails: (c.fails),
	}
}

func (c *Command) Run() (*bytes.Buffer, error) {
	output, err := c.run()

	c.mu.Lock()
	defer c.mu.Unlock()
	c.runs++

	if err != nil {
		c.fails++
		return output, fmt.Errorf("running '%s': %w", c.Name, err)
	}

	return output, nil
}

// run read-locks a command before running it.
func (c *Command) run() (*bytes.Buffer, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return Run(c.Command, c.Shell, c.Timeout.Duration)
}

// Run runs any provided command and returns the output. Exported for convenience only.
func Run(command string, shell bool, timeout time.Duration) (*bytes.Buffer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var cmd *exec.Cmd

	args := strings.Fields(command)
	if len(args) == 0 {
		return nil, ErrNoCmd
	}

	cmdPath, err := filepath.Abs(args[0])
	if err != nil {
		return nil, fmt.Errorf("finding command path: %w", err)
	}

	if shell {
		if runtime.GOOS == mnd.Windows {
			args = append([]string{"cmd", "/C"}, args...)
		} else {
			args = append([]string{"/bin/sh", "-c"}, args...)
		}
	}

	switch len(args) {
	case 0:
		return nil, ErrNoCmd
	case 1:
		cmd = exec.CommandContext(ctx, cmdPath)
	default:
		cmd = exec.CommandContext(ctx, cmdPath, args[1:]...)
	}

	var out bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &out
	// cmd.Env = env.Env()
	// cmd.Env = append(cmd.Env, "PATH="+os.Getenv("PATH"))

	if err := cmd.Run(); err != nil {
		return &out, fmt.Errorf("running cmd %q: %w", strings.Join(cmd.Args, " "), err)
	}

	return &out, nil
}
