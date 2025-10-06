package checkapp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/Notifiarr/notifiarr/pkg/snapshot"
)

func MySQL(ctx context.Context, config snapshot.MySQLConfig) (string, int) {
	snaptest := &snapshot.Snapshot{}

	if config.Host == "" {
		return "Host is required", http.StatusBadRequest
	}

	if config.User == "" {
		return "Username is required", http.StatusBadRequest
	}

	errs := snaptest.GetMySQL(ctx, []snapshot.MySQLConfig{config}, 1)
	if len(errs) > 0 {
		msg := fmt.Sprintf("%d errors encountered: ", len(errs))
		for _, err := range errs {
			msg += err.Error()
		}

		return msg, http.StatusBadGateway
	}

	return "Connection Successful! Processes: " +
		strconv.Itoa(len(snaptest.MySQL[config.Host].Processes)), http.StatusOK
}

func Nvidia(ctx context.Context, config snapshot.NvidiaConfig) (string, int) {
	if config.SMIPath != "" {
		if _, err := os.Stat(config.SMIPath); err != nil {
			return fmt.Sprintf("nvidia-smi not found at provided path '%s': %v", config.SMIPath, err), http.StatusNotAcceptable
		}
	} else if _, err := exec.LookPath("nvidia-smi"); err != nil {
		return fmt.Sprintf("unable to locate nvidia-smi in PATH '%s'", os.Getenv("PATH")), http.StatusNotAcceptable
	}

	snaptest := &snapshot.Snapshot{}
	config.Disabled = false

	if err := snaptest.GetNvidia(ctx, config); err != nil {
		return err.Error(), http.StatusFailedDependency
	}

	msg := fmt.Sprintf("SMI found %d Graphics Adapter", len(snaptest.Nvidia))

	switch len(snaptest.Nvidia) {
	case 0:
		msg += "s."
	case 1:
		msg += ":"
	default:
		msg += "s:"
	}

	for idx, adapter := range snaptest.Nvidia {
		if idx != 0 {
			msg += ", "
		}

		msg += adapter.BusID
	}

	return msg, http.StatusOK
}
