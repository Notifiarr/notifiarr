package checkapp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

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
		var msgSb28 strings.Builder
		for _, err := range errs {
			msgSb28.WriteString(err.Error())
		}

		return msg + msgSb28.String(), http.StatusBadGateway
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
		return "SMI: No graphics devices found.", http.StatusFailedDependency
	case 1:
		msg += ":"
	default:
		msg += "s:"
	}

	var msgSb66 strings.Builder
	for idx, adapter := range snaptest.Nvidia {
		if idx != 0 {
			msgSb66.WriteString(", ")
		}

		msgSb66.WriteString(adapter.BusID)
	}
	msg += msgSb66.String()

	return msg, http.StatusOK
}
