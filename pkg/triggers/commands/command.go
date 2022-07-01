// package commands provides the interfaces and structures to trigger and run shell commands.
package commands

import (
	"bytes"
	"context"
	"crypto/md5" //nolint:gosec
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

// setup must run in the creation routine.
func (c *Command) setup(logger mnd.Logger) {
	c.log = logger
	c.Hash = fmt.Sprintf("%x", md5.Sum([]byte(c.Command+strconv.FormatBool(c.Shell)))) //nolint:gosec

	if c.Name == "" {
		if args := strings.Fields(c.Command); len(args) > 0 {
			c.Name = args[0]
		}
	}

	if c.Timeout.Duration == 0 {
		c.Timeout.Duration = defaultTimeout
	}
}

// run executes this command and logs the output.
func (c *Command) run(event website.EventType) {
	output, err := c.exec(context.Background())

	c.mu.Lock()
	defer c.mu.Unlock()
	c.runs++

	if c.LogOutput && output.Len() > 0 {
		c.log.Printf("[%s requested] Custom Command '%s' Output: %s", output)
	}

	if err != nil {
		c.fails++
		c.log.Errorf("[%s requested] Custom Command '%s' Failed: %w", c.Name, err)
	}
}

// exec read-locks a command before running it and returning the output.
func (c *Command) exec(ctx context.Context) (*bytes.Buffer, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.Timeout.Duration)
	defer cancel()

	return run(ctx, c.Command, c.Shell)
}

// run runs any provided command and returns the output.
func run(ctx context.Context, command string, shell bool) (*bytes.Buffer, error) {
	cmd, err := getCmd(ctx, command, shell)
	if err != nil {
		return nil, err
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

// getCmd returns the exec.Cmd for the provided arguments.
func getCmd(ctx context.Context, command string, shell bool) (*exec.Cmd, error) {
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

	var cmd *exec.Cmd

	switch len(args) {
	case 0:
		return nil, ErrNoCmd
	case 1:
		cmd = exec.CommandContext(ctx, cmdPath)
	default:
		cmd = exec.CommandContext(ctx, cmdPath, args[1:]...)
	}

	return cmd, nil
}
