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
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/google/shlex"
)

// Setup must run in the creation routine.
func (c *Command) Setup(logger mnd.Logger, website *website.Server) {
	if c.Name == "" {
		if args, _ := shlex.Split(c.Command); len(args) > 0 {
			c.Name = args[0]
		}
	}

	c.Hash = fmt.Sprintf("%x", md5.Sum([]byte(c.Name+c.Command+strconv.FormatBool(c.Shell)))) //nolint:gosec
	c.log = logger
	c.website = website

	if c.Timeout.Duration == 0 {
		c.Timeout.Duration = defaultTimeout
	}
}

// run executes this command and logs the output. This is executed from the trigger channel.
func (c *Command) run(event website.EventType) {
	_, _ = c.RunNow(context.Background(), event)
}

// RunNow runs the command immediately, waits for and returns the output.
func (c *Command) RunNow(ctx context.Context, event website.EventType) (string, error) {
	output, err := c.exec(ctx)
	oLen := 0
	oStr := output.String()

	eStr := ""
	if err != nil {
		eStr = err.Error()
	} else {
		oLen = output.Len()
	}

	// Send the notification before the lock.
	if c.Notify {
		c.website.SendData(&website.Request{
			Route: "/no/path/for/commands/output/yet/sorry",
			Event: event,
			Payload: map[string]string{
				"name":    c.Name,
				"command": c.Command,
				"hash":    c.Hash,
				"output":  oStr,
				"error":   eStr,
			},
			LogMsg:     fmt.Sprintf("Custom Command '%s' Output", c.Name),
			LogPayload: c.Log,
		})
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.runs++
	c.lastRun = time.Now().Round(time.Second)
	c.output = oStr

	if err != nil {
		c.fails++
		c.output = eStr + ": " + oStr

		if c.Log && oStr != "" {
			c.log.Errorf("[%s requested] Custom Command '%s' Failed: %v, Output: %s", event, c.Name, err, oStr)
		} else {
			c.log.Errorf("[%s requested] Custom Command '%s' Failed: %v", event, c.Name, err)
		}
	} else if c.Log && oLen > 0 {
		c.log.Printf("[%s requested] Custom Command '%s' Output: %s", event, c.Name, oStr)
	}

	return oStr, err
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

	if err := cmd.Run(); err != nil {
		return &out, fmt.Errorf(`running cmd %q: %w`, cmd.Args, err)
	}

	return &out, nil
}

// getCmd returns the exec.Cmd for the provided arguments.
func getCmd(ctx context.Context, command string, shell bool) (*exec.Cmd, error) {
	args, err := getArgs(command, shell)
	if err != nil {
		return nil, err
	}

	var cmd *exec.Cmd
	// nolint:gosec
	switch len(args) {
	case 0:
		return nil, ErrNoCmd
	case 1:
		cmd = exec.CommandContext(ctx, args[0])
	default:
		cmd = exec.CommandContext(ctx, args[0], args[1:]...)
	}

	return cmd, nil
}

func getArgs(command string, shell bool) ([]string, error) {
	if runtime.GOOS != mnd.Windows && shell {
		return []string{"/bin/sh", "-c", command}, nil
	}

	// Special shell-split command.
	args, err := shlex.Split(command)
	if err != nil {
		return nil, fmt.Errorf("splitting shell command: %w", err)
	}

	if len(args) == 0 {
		return nil, ErrNoCmd
	}

	if args[0], err = filepath.Abs(args[0]); err != nil {
		return nil, fmt.Errorf("finding command path: %w", err)
	}

	if shell { // if shell is set, we know it's windows.
		return append([]string{"cmd", "/C"}, args...), nil
	}

	return args, nil
}
