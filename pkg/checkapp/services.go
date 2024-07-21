package checkapp

import (
	"context"
	"net/http"

	"github.com/Notifiarr/notifiarr/pkg/services"
)

func testTCP(ctx context.Context, svc *services.Service) (string, int) {
	if err := svc.Validate(); err != nil {
		return validation + err.Error(), http.StatusBadRequest
	}

	res := svc.CheckOnly(ctx)
	if res.State != services.StateOK {
		return res.State.String() + " " + res.Output.String(), http.StatusBadGateway
	}

	return "TCP Port is OPEN and reachable: " + res.Output.String(), http.StatusOK
}

func testHTTP(ctx context.Context, svc *services.Service) (string, int) {
	if err := svc.Validate(); err != nil {
		return validation + err.Error(), http.StatusBadRequest
	}

	res := svc.CheckOnly(ctx)
	if res.State != services.StateOK {
		return res.State.String() + " " + res.Output.String(), http.StatusBadGateway
	}

	// add test
	return "HTTP Response Code Acceptable! " + res.Output.String(), http.StatusOK
}

func testProcess(ctx context.Context, svc *services.Service) (string, int) {
	if err := svc.Validate(); err != nil {
		return validation + err.Error(), http.StatusBadRequest
	}

	res := svc.CheckOnly(ctx)
	if res.State != services.StateOK {
		return res.State.String() + " " + res.Output.String(), http.StatusBadGateway
	}

	return "Process Tested OK: " + res.Output.String(), http.StatusOK
}

func testPing(ctx context.Context, svc *services.Service) (string, int) {
	if err := svc.Validate(); err != nil {
		return validation + err.Error(), http.StatusBadRequest
	}

	res := svc.CheckOnly(ctx)
	if res.State != services.StateOK {
		return res.State.String() + " " + res.Output.String(), http.StatusBadGateway
	}

	return "Ping Tested OK: " + res.Output.String(), http.StatusOK
}
