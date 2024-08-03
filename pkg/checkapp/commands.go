package checkapp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Notifiarr/notifiarr/pkg/triggers/commands"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

func testCommand(ctx context.Context, input *Input) (string, int) {
	if len(input.Real.Commands) > input.Index {
		input.Real.Commands[input.Index].Run(&common.ActionInput{Type: website.EventGUI})
		return "Command Triggered: " + input.Real.Commands[input.Index].Name, http.StatusOK
	} else if len(input.Post.Commands) > input.Index { // check POST input for "new" command.
		input.Post.Commands[input.Index].Setup(input.Real.Logger, input.Real.Server)

		if err := input.Post.Commands[input.Index].SetupRegexpArgs(); err != nil {
			return err.Error(), http.StatusInternalServerError
		}

		return testCustomCommand(ctx, input.Post.Commands[input.Index])
	}

	return ErrBadIndex.Error(), http.StatusBadRequest
}

func testCustomCommand(ctx context.Context, cmd *commands.Command) (string, int) {
	ctx, cancel := context.WithTimeout(ctx, cmd.Timeout.Duration)
	defer cancel()

	output, err := cmd.RunNow(ctx, &common.ActionInput{Type: website.EventGUI})
	if err != nil {
		return fmt.Sprintf("Command Failed! Error: %v", err), http.StatusInternalServerError
	}

	return "Command Successful! Output: " + output, http.StatusOK
}
