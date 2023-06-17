package client

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/services"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/triggers/commands"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/gorilla/mux"
	"github.com/hekmon/transmissionrpc/v2"
	"golift.io/deluge"
	"golift.io/nzbget"
	"golift.io/qbit"
	"golift.io/starr"
	"golift.io/starr/lidarr"
	"golift.io/starr/prowlarr"
	"golift.io/starr/radarr"
	"golift.io/starr/readarr"
	"golift.io/starr/sonarr"
)

//nolint:funlen,cyclop,gocognit,gocyclo
func (c *Client) testInstance(response http.ResponseWriter, request *http.Request) {
	config := configfile.Config{}

	if err := configPostDecoder.Decode(&config, request.PostForm); err != nil {
		http.Error(response, "Decoding POST data into Go data structure failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	index, _ := strconv.Atoi(mux.Vars(request)["index"])
	reply, code := "Unknown Check Type Requested!", http.StatusNotImplemented

	switch mux.Vars(request)["type"] {
	case "Commands":
		if len(c.Config.Commands) > index {
			c.Config.Commands[index].Run(&common.ActionInput{Type: website.EventGUI})
			reply, code = fmt.Sprintf("Command Triggered: %s", c.Config.Commands[index].Name), http.StatusOK
		} else if len(config.Commands) > index { // check POST input for "new" command.
			config.Commands[index].Setup(c.Logger, c.website)
			if err := config.Commands[index].SetupRegexpArgs(); err != nil {
				reply, code = err.Error(), http.StatusInternalServerError
			} else {
				reply, code = testCustomCommand(request.Context(), config.Commands[index])
			}
		}
	// Downloaders.
	case "NZBGet":
		if len(config.Apps.NZBGet) > index {
			reply, code = testNZBGet(request.Context(), config.Apps.NZBGet[index].Config)
		}
	case "Deluge":
		if len(config.Apps.Deluge) > index {
			reply, code = testDeluge(request.Context(), config.Apps.Deluge[index].Config)
		}
	case "Qbit":
		if len(config.Apps.Qbit) > index {
			reply, code = testQbit(request.Context(), config.Apps.Qbit[index].Config)
		}
	case "Rtorrent":
		if len(config.Apps.Rtorrent) > index {
			reply, code = testRtorrent(config.Apps.Rtorrent[index])
		}
	case "Transmission":
		if len(config.Apps.Transmission) > index {
			reply, code = testTransmission(request.Context(), config.Apps.Transmission[index])
		}
	case "SabNZB":
		if len(config.Apps.SabNZB) > index {
			reply, code = testSabNZB(request.Context(), config.Apps.SabNZB[index])
		}
	// Starrs.
	case "Lidarr":
		if len(config.Apps.Lidarr) > index {
			reply, code = testLidarr(request.Context(), config.Apps.Lidarr[index].Config)
		}
	case "Prowlarr":
		if len(config.Apps.Prowlarr) > index {
			reply, code = testProwlarr(request.Context(), config.Apps.Prowlarr[index].Config)
		}
	case "Radarr":
		if len(config.Apps.Radarr) > index {
			reply, code = testRadarr(request.Context(), config.Apps.Radarr[index].Config)
		}
	case "Readarr":
		if len(config.Apps.Readarr) > index {
			reply, code = testReadarr(request.Context(), config.Apps.Readarr[index].Config)
		}
	case "Sonarr":
		if len(config.Apps.Sonarr) > index {
			reply, code = testSonarr(request.Context(), config.Apps.Sonarr[index].Config)
		}
	// Snapshots.
	case "MySQL":
		if config.Snapshot != nil && config.Snapshot.Plugins != nil && len(config.Snapshot.Plugins.MySQL) > index {
			reply, code = testMySQL(request.Context(), config.Snapshot.Plugins.MySQL[index])
		}
		// Snapshots.
	case "Nvidia":
		if config.Snapshot != nil && config.Snapshot.Plugins != nil && config.Snapshot.Plugins.Nvidia != nil {
			reply, code = testNvidia(request.Context(), config.Snapshot.Plugins.Nvidia)
		}
	// Services.
	case "Tcp":
		if len(config.Service) > index {
			reply, code = testTCP(request.Context(), config.Service[index])
		}
	case "Http":
		if len(config.Service) > index {
			reply, code = testHTTP(request.Context(), config.Service[index])
		}
	case "Process":
		if len(config.Service) > index {
			reply, code = testProcess(request.Context(), config.Service[index])
		}
	case "Ping", "Icmp":
		if len(config.Service) > index {
			reply, code = testPing(request.Context(), config.Service[index])
		}
	// Media
	case "Plex":
		reply, code = testPlex(request.Context(), config.Plex)
	case "Tautulli":
		reply, code = testTautulli(request.Context(), config.Apps.Tautulli)
	}

	http.Error(response, reply, code)
}

func testDeluge(ctx context.Context, config *deluge.Config) (string, int) {
	deluge, err := deluge.New(ctx, config)
	if err != nil {
		return "Connecting: " + err.Error(), http.StatusBadGateway
	}

	return fmt.Sprintf("Connection Successful! Version: %s", deluge.Version), http.StatusOK
}

func testNZBGet(ctx context.Context, config *nzbget.Config) (string, int) {
	ver, err := nzbget.New(config).VersionContext(ctx)
	if err != nil {
		return "Getting Version: " + err.Error(), http.StatusBadGateway
	}

	return fmt.Sprintf("Connection Successful! Version: %s", ver), http.StatusOK
}

func testCustomCommand(ctx context.Context, cmd *commands.Command) (string, int) {
	ctx, cancel := context.WithTimeout(ctx, cmd.Timeout.Duration)
	defer cancel()

	output, err := cmd.RunNow(ctx, &common.ActionInput{Type: website.EventGUI})
	if err != nil {
		return fmt.Sprintf("Command Failed! Error: %v", err), http.StatusInternalServerError
	}

	return fmt.Sprintf("Command Successful! Output: %s", output), http.StatusOK
}

func testQbit(ctx context.Context, config *qbit.Config) (string, int) {
	if qbit, err := qbit.New(ctx, config); err != nil {
		return "Connecting: " + err.Error(), http.StatusBadGateway
	} else if xfers, err := qbit.GetXfersContext(ctx); err != nil {
		return "Getting Transfers: " + err.Error(), http.StatusBadGateway
	} else {
		return fmt.Sprintf("Connection Successful! %d Transfers", len(xfers)), http.StatusOK
	}
}

func testRtorrent(config *apps.RtorrentConfig) (string, int) {
	config.Setup(0, nil)

	result, err := config.Client.Call("system.hostname")
	if err != nil {
		return "Getting Server Name: " + err.Error(), http.StatusBadGateway
	}

	if names, ok := result.([]interface{}); ok {
		result = names[0]
	}

	if name, ok := result.(string); ok {
		return fmt.Sprintf("Connection Successful! Server name: %s", name), http.StatusOK
	}

	return "Getting Server Name: result was not a string?", http.StatusBadGateway
}

func testTransmission(ctx context.Context, config *apps.XmissionConfig) (string, int) {
	client := transmissionrpc.NewClient(transmissionrpc.Config{
		URL:       config.URL,
		Username:  config.User,
		Password:  config.Pass,
		UserAgent: mnd.Title,
	})

	args, err := client.SessionArgumentsGetAll(ctx)
	if err != nil {
		return "Getting Server Version: " + err.Error(), http.StatusBadGateway
	}

	return fmt.Sprintln("Transmission Server version:", *args.Version), http.StatusOK
}

func testSabNZB(ctx context.Context, app *apps.SabNZBConfig) (string, int) {
	app.Setup(0, nil)

	sab, err := app.GetQueue(ctx)
	if err != nil {
		return "Getting Queue: " + err.Error(), http.StatusBadGateway
	}

	return "Connection Successful! Version: " + sab.Version, http.StatusOK
}

func testLidarr(ctx context.Context, config *starr.Config) (string, int) {
	status, err := lidarr.New(config).GetSystemStatusContext(ctx)
	if err != nil {
		return "Connecting: " + err.Error(), http.StatusBadGateway
	}

	return "Connection Successful! Version: " + status.Version, http.StatusOK
}

func testProwlarr(ctx context.Context, config *starr.Config) (string, int) {
	status, err := prowlarr.New(config).GetSystemStatusContext(ctx)
	if err != nil {
		return "Connecting: " + err.Error(), http.StatusBadGateway
	}

	return "Connection Successful! Version: " + status.Version, http.StatusOK
}

func testRadarr(ctx context.Context, config *starr.Config) (string, int) {
	status, err := radarr.New(config).GetSystemStatusContext(ctx)
	if err != nil {
		return "Connecting: " + err.Error(), http.StatusBadGateway
	}

	return "Connection Successful! Version: " + status.Version, http.StatusOK
}

func testReadarr(ctx context.Context, config *starr.Config) (string, int) {
	status, err := readarr.New(config).GetSystemStatusContext(ctx)
	if err != nil {
		return "Connecting: " + err.Error(), http.StatusBadGateway
	}

	return "Connection Successful! Version: " + status.Version, http.StatusOK
}

func testSonarr(ctx context.Context, config *starr.Config) (string, int) {
	status, err := sonarr.New(config).GetSystemStatusContext(ctx)
	if err != nil {
		return "Connecting: " + err.Error(), http.StatusBadGateway
	}

	return "Connection Successful! Version: " + status.Version, http.StatusOK
}

func testMySQL(ctx context.Context, config *snapshot.MySQLConfig) (string, int) {
	snaptest := &snapshot.Snapshot{}

	errs := snaptest.GetMySQL(ctx, []*snapshot.MySQLConfig{config}, 1)
	if len(errs) > 0 {
		msg := fmt.Sprintf("%d errors encountered: ", len(errs))
		for _, err := range errs {
			msg += err.Error()
		}

		return msg, http.StatusBadGateway
	}

	return "Connection Successful!", http.StatusOK
}

func testNvidia(ctx context.Context, config *snapshot.NvidiaConfig) (string, int) {
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
		return err.Error(), http.StatusBadGateway
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

	for _, adapter := range snaptest.Nvidia {
		msg += "<br>" + adapter.BusID
	}

	return msg, http.StatusOK
}

func testTCP(ctx context.Context, svc *services.Service) (string, int) {
	if err := svc.Validate(); err != nil {
		return "Validation: " + err.Error(), http.StatusBadRequest
	}

	res := svc.CheckOnly(ctx)
	if res.State != services.StateOK {
		return res.State.String() + " " + res.Output, http.StatusBadGateway
	}

	return "TCP Port is OPEN and reachable: " + res.Output, http.StatusOK
}

func testHTTP(ctx context.Context, svc *services.Service) (string, int) {
	if err := svc.Validate(); err != nil {
		return "Validation: " + err.Error(), http.StatusBadRequest
	}

	res := svc.CheckOnly(ctx)
	if res.State != services.StateOK {
		return res.State.String() + " " + res.Output, http.StatusBadGateway
	}

	// add test
	return "HTTP Response Code Acceptable! " + res.Output, http.StatusOK
}

func testProcess(ctx context.Context, svc *services.Service) (string, int) {
	if err := svc.Validate(); err != nil {
		return "Validation: " + err.Error(), http.StatusBadRequest
	}

	res := svc.CheckOnly(ctx)
	if res.State != services.StateOK {
		return res.State.String() + " " + res.Output, http.StatusBadGateway
	}

	return "Process Tested OK: " + res.Output, http.StatusOK
}

func testPing(ctx context.Context, svc *services.Service) (string, int) {
	if err := svc.Validate(); err != nil {
		return "Validation: " + err.Error(), http.StatusBadRequest
	}

	res := svc.CheckOnly(ctx)
	if res.State != services.StateOK {
		return res.State.String() + " " + res.Output, http.StatusBadGateway
	}

	return "Ping Tested OK: " + res.Output, http.StatusOK
}

func testPlex(ctx context.Context, app *apps.PlexConfig) (string, int) {
	app.Setup(0, nil)

	info, err := app.GetInfo(ctx)
	if err != nil {
		return "Getting Info: " + err.Error(), http.StatusBadGateway
	}

	return "Plex OK! Version: " + info.Version, http.StatusOK
}

func testTautulli(ctx context.Context, app *apps.TautulliConfig) (string, int) {
	app.Setup(0, nil)

	users, err := app.GetUsers(ctx)
	if err != nil {
		return "Getting Users: " + err.Error(), http.StatusBadGateway
	}

	return fmt.Sprintf("Tautulli OK! Users: %d", len(users.Response.Data)), http.StatusOK
}
