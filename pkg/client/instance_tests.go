package client

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/services"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/gorilla/mux"
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

func testInstance(response http.ResponseWriter, request *http.Request) { //nolint:funlen,cyclop,gocognit,gocyclo
	config := configfile.Config{}
	if err := configPostDecoder.Decode(&config, request.PostForm); err != nil {
		http.Error(response, "Decoding POST data into Go data structure failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	index, _ := strconv.Atoi(mux.Vars(request)["index"])
	reply, code := "Unknown Check Type Requested!", http.StatusNotImplemented

	switch mux.Vars(request)["type"] {
	// Downloaders.
	case "NZBGet":
		if len(config.Apps.NZBGet) > index {
			reply, code = testNZBGet(config.Apps.NZBGet[index].Config)
		}
	case "Deluge":
		if len(config.Apps.Deluge) > index {
			reply, code = testDeluge(config.Apps.Deluge[index].Config)
		}
	case "Qbit":
		if len(config.Apps.Qbit) > index {
			reply, code = testQbit(config.Apps.Qbit[index].Config)
		}
	case "rTorrent":
		if len(config.Apps.Rtorrent) > index {
			reply, code = testRtorrent(config.Apps.Rtorrent[index])
		}
	case "SabNZB":
		if len(config.Apps.SabNZB) > index {
			reply, code = testSabNZB(config.Apps.SabNZB[index])
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
	// Services.
	case "Tcp":
		if len(config.Service) > index {
			reply, code = testTCP(config.Service[index])
		}
	case "Http":
		if len(config.Service) > index {
			reply, code = testHTTP(config.Service[index])
		}
	case "Process":
		if len(config.Service) > index {
			reply, code = testProcess(config.Service[index])
		}
	// Media
	case "Plex":
		reply, code = testPlex(request.Context(), config.Plex)
	case "Tautulli":
		reply, code = testTautulli(config.Apps.Tautulli)
	}

	http.Error(response, reply, code)
}

func testDeluge(config *deluge.Config) (string, int) {
	if _, deluge, err := deluge.New(config); err != nil {
		return "Connecting: " + err.Error(), http.StatusBadGateway
	} else if _, xfers, err := deluge.GetXfers(); err != nil {
		return "Getting Transfers: " + err.Error(), http.StatusBadGateway
	} else {
		return fmt.Sprintf("Connection Successful! %d Transfers", len(xfers)), http.StatusOK
	}
}

func testNZBGet(config *nzbget.Config) (string, int) {
	ver, _, err := nzbget.New(config).Version()
	if err != nil {
		return "Getting Version: " + err.Error(), http.StatusBadGateway
	}

	return fmt.Sprintf("Connection Successful! Version: %s", ver), http.StatusOK
}

func testQbit(config *qbit.Config) (string, int) {
	if qbit, err := qbit.New(config); err != nil {
		return "Connecting: " + err.Error(), http.StatusBadGateway
	} else if _, xfers, err := qbit.GetXfers(); err != nil {
		return "Getting Transfers: " + err.Error(), http.StatusBadGateway
	} else {
		return fmt.Sprintf("Connection Successful! %d Transfers", len(xfers)), http.StatusOK
	}
}

func testRtorrent(config *apps.RtorrentConfig) (string, int) {
	config.Setup(time.Minute)

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

func testSabNZB(config *apps.SabNZBConfig) (string, int) {
	sab, err := config.GetQueue()
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

func testTCP(svc *services.Service) (string, int) {
	if err := svc.Validate(); err != nil {
		return "Validation: " + err.Error(), http.StatusBadRequest
	}

	res := svc.CheckOnly()
	if res.State != services.StateOK {
		return res.State.String() + " " + res.Output, http.StatusBadGateway
	}

	return "TCP Port is OPEN and reachable: " + res.Output, http.StatusOK
}

func testHTTP(svc *services.Service) (string, int) {
	if err := svc.Validate(); err != nil {
		return "Validation: " + err.Error(), http.StatusBadRequest
	}

	res := svc.CheckOnly()
	if res.State != services.StateOK {
		return res.State.String() + " " + res.Output, http.StatusBadGateway
	}

	// add test
	return "HTTP Response Code Acceptable! " + res.Output, http.StatusOK
}

func testProcess(svc *services.Service) (string, int) {
	if err := svc.Validate(); err != nil {
		return "Validation: " + err.Error(), http.StatusBadRequest
	}

	res := svc.CheckOnly()
	if res.State != services.StateOK {
		return res.State.String() + " " + res.Output, http.StatusBadGateway
	}

	return "Process Tested OK: " + res.Output, http.StatusOK
}

func testPlex(ctx context.Context, app *plex.Server) (string, int) {
	app.Validate()

	info, err := app.GetInfo(ctx)
	if err != nil {
		return "Getting Info: " + err.Error(), http.StatusBadGateway
	}

	return "Plex OK! Version: " + info.Version, http.StatusOK
}

func testTautulli(app *apps.TautulliConfig) (string, int) {
	users, err := app.GetUsers()
	if err != nil {
		return "Getting Users: " + err.Error(), http.StatusBadGateway
	}

	return fmt.Sprintf("Tautulli OK! Users: %d", len(users.Response.Data)), http.StatusOK
}
