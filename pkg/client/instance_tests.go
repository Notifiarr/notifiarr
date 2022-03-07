package client

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/gorilla/mux"
	"golift.io/deluge"
	"golift.io/qbit"
	"golift.io/starr"
	"golift.io/starr/lidarr"
	"golift.io/starr/prowlarr"
	"golift.io/starr/radarr"
	"golift.io/starr/readarr"
	"golift.io/starr/sonarr"
)

func testInstance(response http.ResponseWriter, request *http.Request) {
	config := configfile.Config{}
	if err := configPostDecoder.Decode(&config, request.PostForm); err != nil {
		http.Error(response, "Decoding POST data into Go data structure failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	index, _ := strconv.Atoi(mux.Vars(request)["index"])
	reply, code := "This check does not exist yet, sorry. :(", http.StatusNotImplemented

	switch mux.Vars(request)["type"] {
	// Downloaders.
	case "Deluge":
		if len(config.Apps.Deluge) > index {
			reply, code = testDeluge(config.Apps.Deluge[index].Config)
		}
	case "Qbit":
		if len(config.Apps.Qbit) > index {
			reply, code = testQbit(config.Apps.Qbit[index].Config)
		}
	case "SabNZB":
		if len(config.Apps.SabNZB) > index {
			reply, code = testSabNZB(config.Apps.SabNZB[index])
		}
	// Starrs.
	case "Lidarr":
		if len(config.Apps.Lidarr) > index {
			reply, code = testLidarr(config.Apps.Lidarr[index].Config)
		}
	case "Prowlarr":
		if len(config.Apps.Prowlarr) > index {
			reply, code = testProwlarr(config.Apps.Prowlarr[index].Config)
		}
	case "Radarr":
		if len(config.Apps.Radarr) > index {
			reply, code = testRadarr(config.Apps.Radarr[index].Config)
		}
	case "Readarr":
		if len(config.Apps.Readarr) > index {
			reply, code = testReadarr(config.Apps.Readarr[index].Config)
		}
	case "Sonarr":
		if len(config.Apps.Sonarr) > index {
			reply, code = testSonarr(config.Apps.Sonarr[index].Config)
		}
	// Snapshots.
	case "MySQL":
	// Services.
	case "TCP":
	case "HTTP":
	case "Process":
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

func testQbit(config *qbit.Config) (string, int) {
	if qbit, err := qbit.New(config); err != nil {
		return "Connecting: " + err.Error(), http.StatusBadGateway
	} else if _, xfers, err := qbit.GetXfers(); err != nil {
		return "Getting Transfers: " + err.Error(), http.StatusBadGateway
	} else {
		return fmt.Sprintf("Connection Successful! %d Transfers", len(xfers)), http.StatusOK
	}
}

func testSabNZB(config *apps.SabNZBConfig) (string, int) {
	if sab, err := config.GetQueue(); err != nil {
		return "Getting Queue: " + err.Error(), http.StatusBadGateway
	} else {
		return "Connection Successful! Version: " + sab.Version, http.StatusOK
	}
}

func testLidarr(config *starr.Config) (string, int) {
	status, err := lidarr.New(config).GetSystemStatus()
	if err != nil {
		return "Connecting: " + err.Error(), http.StatusBadGateway
	}

	return "Connection Successful! Version: " + status.Version, http.StatusOK
}

func testProwlarr(config *starr.Config) (string, int) {
	status, err := prowlarr.New(config).GetSystemStatus()
	if err != nil {
		return "Connecting: " + err.Error(), http.StatusBadGateway
	}

	return "Connection Successful! Version: " + status.Version, http.StatusOK
}

func testRadarr(config *starr.Config) (string, int) {
	status, err := radarr.New(config).GetSystemStatus()
	if err != nil {
		return "Connecting: " + err.Error(), http.StatusBadGateway
	}

	return "Connection Successful! Version: " + status.Version, http.StatusOK
}

func testReadarr(config *starr.Config) (string, int) {
	status, err := readarr.New(config).GetSystemStatus()
	if err != nil {
		return "Connecting: " + err.Error(), http.StatusBadGateway
	}

	return "Connection Successful! Version: " + status.Version, http.StatusOK
}

func testSonarr(config *starr.Config) (string, int) {
	status, err := sonarr.New(config).GetSystemStatus()
	if err != nil {
		return "Connecting: " + err.Error(), http.StatusBadGateway
	}

	return "Connection Successful! Version: " + status.Version, http.StatusOK
}
