package checkapp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

func testCommand(ctx context.Context, input *Input) (string, int) {
	action := &common.ActionInput{Type: website.EventGUI}
	for _, arg := range input.Args {
		action.Args = append(action.Args, arg...)
	}

	if len(input.Real.Commands) > input.Index {
		input.Real.Commands[input.Index].Run(action)
		return "Command Triggered: " + input.Real.Commands[input.Index].Name, http.StatusOK
	} else if len(input.Post.Commands) > input.Index { // check POST input for "new" command.
		return testCustomCommand(ctx, input, action)
	}

	return ErrBadIndex.Error(), http.StatusBadRequest
}

func testCustomCommand(ctx context.Context, input *Input, action *common.ActionInput) (string, int) {
	cmd := input.Post.Commands[input.Index]
	cmd.Setup()

	if err := cmd.SetupRegexpArgs(); err != nil {
		return err.Error(), http.StatusInternalServerError
	}

	ctx, cancel := context.WithTimeout(ctx, cmd.Timeout.Duration)
	defer cancel()

	output, err := cmd.RunNow(ctx, action)
	if err != nil {
		return fmt.Sprintf("Command Failed! Error: %v", err), http.StatusInternalServerError
	}

	return "Command Successful! Output: " + output, http.StatusOK
}
