package fileupload

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

const TrigUploadFile common.TriggerName = "File Upload"

var ErrInvalidFile = fmt.Errorf("invalid file provided")

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
		Name: TrigUploadFile,
		Fn:   c.uploadFiles,
		C:    make(chan *common.ActionInput, 1),
	})
}

// Log uploads a specific log file to Notifiarr.com.
func (a *Action) Log(event website.EventType, file string) error {
	switch file {
	case "app":
		return a.Upload(event, a.cmd.Logger.LogConfig.LogFile)
	case "debug":
		return a.Upload(event, a.cmd.Logger.LogConfig.DebugLog)
	case "http":
		return a.Upload(event, a.cmd.Logger.LogConfig.HTTPLog)
	default:
		return ErrInvalidFile
	}
}

// Upload a file or files to Notifiarr.com.
func (a *Action) Upload(event website.EventType, filePath ...string) error {
	// Make sure the files exist.
	for _, file := range filePath {
		if _, err := os.Stat(file); err != nil {
			return fmt.Errorf("file stat fail: %w", err)
		}
	}

	a.cmd.Exec(&common.ActionInput{Type: event, Args: filePath}, TrigUploadFile)

	return nil
}

func (c *cmd) uploadFiles(ctx context.Context, input *common.ActionInput) {
	for _, fileName := range input.Args {
		c.uploadFile(ctx, input.Type, fileName)
	}
}

func (c *cmd) uploadFile(_ context.Context, event website.EventType, fileName string) {
	// Add a file to the request
	file, err := os.Open(fileName)
	if err != nil {
		c.Errorf("[%s requested] Opening file '%s' for Upload failed: %v", event, fileName, err)
		return
	}

	c.SendData(&website.Request{
		Route:  website.UploadRoute,
		Event:  event,
		LogMsg: fmt.Sprintf("Upload file %s", fileName),
		UploadFile: &website.UploadFile{
			FileName:   filepath.Base(fileName),
			ReadCloser: file,
		},
	})
}
