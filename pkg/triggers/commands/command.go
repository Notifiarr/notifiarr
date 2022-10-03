// package commands provides the interfaces and structures to trigger and run shell commands.
package commands

import (
	"bytes"
	"context"
	"crypto/md5" //nolint:gosec
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/hugelgupf/go-shlex"
)

var ErrDisabled = fmt.Errorf("the command is disabled due to an error")

const (
	argPfx = "({"
	argSfx = "})"
)

// Setup must run in the creation routine.
func (c *Command) Setup(logger mnd.Logger, website *website.Server) error {
	if c.Name == "" {
		if args := shlex.Split(c.Command); len(args) > 0 {
			c.Name = args[0]
		}
	}

	if err := c.setupRegexpArgs(); err != nil {
		return err
	}

	hash := md5.Sum([]byte(fmt.Sprint(c.Name, c.Command, c.Shell, c.Log, c.Notify, c.Timeout))) //nolint:gosec
	c.Hash = fmt.Sprintf("%x", hash)
	c.log = logger
	c.website = website
	c.Args = len(c.args)

	if c.Timeout.Duration == 0 {
		c.Timeout.Duration = defaultTimeout
	}

	return nil
}

func (c *Command) setupRegexpArgs() error {
	c.cmd = c.Command

	matches := regexp.MustCompile(`\({([^}]*)}\)`).FindAllStringIndex(c.cmd, -1)
	if matches == nil {
		if strings.Contains(c.Command, argPfx) {
			return fmt.Errorf("%w: missing custom argument regexp terminator: %s", ErrArgValue, argSfx)
		}

		return nil
	}

	c.args = make([]*regexp.Regexp, len(matches))

	for idx := len(matches) - 1; idx >= 0; idx-- {
		arg := matches[idx]
		instance := idx + 1

		re, err := regexp.Compile(c.cmd[arg[0]+2 : arg[1]-2])
		if err != nil {
			return fmt.Errorf("parsing command '%s' arg %d regexp: %w", c.Name, instance, err)
		}

		c.cmd = c.cmd[:arg[0]] + fmt.Sprintf("%s%d%s", argPfx, instance, argSfx) + c.cmd[arg[1]:]
		c.args[idx] = re
	}

	return nil
}

// run executes this command and logs the output. This is executed from the trigger channel.
func (c *Command) run(ctx context.Context, input *common.ActionInput) {
	_, _ = c.RunNow(ctx, input)
}

// RunNow runs the command immediately, waits for and returns the output.
func (c *Command) RunNow(ctx context.Context, input *common.ActionInput) (string, error) {
	if c.disable {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.output = ErrDisabled.Error()

		return "<command disabled>", ErrDisabled
	}

	output, err := c.exec(ctx, input)
	oStr := output.String()
	eStr := ""

	if err != nil {
		eStr = err.Error()
	}

	// Send the notification before the lock.
	if c.Notify {
		c.website.SendData(&website.Request{
			Route: website.CommandRoute,
			Event: input.Type,
			Payload: map[string]string{
				"name":   c.Name,
				"hash":   c.Hash,
				"output": oStr,
				"error":  eStr,
			},
			LogMsg:     fmt.Sprintf("Custom Command '%s' Output", c.Name),
			LogPayload: c.Log,
		})
	}

	c.logOutput(input, oStr, eStr, err)

	return oStr, err
}

func (c *Command) logOutput(input *common.ActionInput, oStr, eStr string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.runs++
	c.lastRun = time.Now().Round(time.Second)
	c.output = oStr

	if err != nil {
		c.fails++
		c.output = "error: " + eStr + ", output: " + oStr

		if c.Log && oStr != "" {
			c.log.Errorf("[%s requested] Custom Command '%s' Failed: %v, Output: %s", input.Type, c.Name, err, oStr)
		} else {
			c.log.Errorf("[%s requested] Custom Command '%s' Failed: %v", input.Type, c.Name, err)
		}
	} else if c.Log && oStr != "" {
		c.log.Printf("[%s requested] Custom Command '%s' Output: %s", input.Type, c.Name, oStr)
	}
}

// run read locks and runs the command then returns the output.
func (c *Command) exec(ctx context.Context, input *common.ActionInput) (*bytes.Buffer, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.Timeout.Duration)
	defer cancel()

	builder := &cmdBuilder{
		cmd:          c.cmd,
		expectedArgs: c.args,
		providedArgs: input.Args,
		shell:        c.Shell,
	}

	cmd, err := builder.getCmd(ctx)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return &out, fmt.Errorf(`running cmd %s: %w`, cmd.Args, err)
	}

	return &out, nil
}
