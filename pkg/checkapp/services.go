package checkapp

import (
	"context"
	"log"
	"net/http"

	"github.com/Notifiarr/notifiarr/pkg/services"
)

func SvcTCP(ctx context.Context, svc services.ServiceConfig) (string, int) {
	if err := svc.Validate(); err != nil {
		return validation + err.Error(), http.StatusBadRequest
	}

	res := svc.CheckOnly(ctx)
	if res.State != services.StateOK {
		return res.State.String() + " " + res.Output.String(), http.StatusFailedDependency
	}

	return "TCP Port is OPEN and reachable: " + res.Output.String(), http.StatusOK
}

func SvcHTTP(ctx context.Context, svc services.ServiceConfig) (string, int) {
	log.Println("testHTTP", svc)
	if err := svc.Validate(); err != nil {
		return validation + err.Error(), http.StatusBadRequest
	}

	res := svc.CheckOnly(ctx)
	if res.State != services.StateOK {
		return res.State.String() + " " + res.Output.String(), http.StatusFailedDependency
	}

	// add test
	return "HTTP Response Code Acceptable! " + res.Output.String(), http.StatusOK
}

func SvcProcess(ctx context.Context, svc services.ServiceConfig) (string, int) {
	if err := svc.Validate(); err != nil {
		return validation + err.Error(), http.StatusBadRequest
	}

	res := svc.CheckOnly(ctx)
	if res.State != services.StateOK {
		return res.State.String() + " " + res.Output.String(), http.StatusFailedDependency
	}

	return "Process Tested OK: " + res.Output.String(), http.StatusOK
}

func SvcPing(ctx context.Context, svc services.ServiceConfig) (string, int) {
	if err := svc.Validate(); err != nil {
		return validation + err.Error(), http.StatusBadRequest
	}

	res := svc.CheckOnly(ctx)
	if res.State != services.StateOK {
		return res.State.String() + " " + res.Output.String(), http.StatusFailedDependency
	}

	return "Ping Tested OK: " + res.Output.String(), http.StatusOK
}
