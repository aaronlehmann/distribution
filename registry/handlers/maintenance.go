package handlers

import (
	"net/http"

	"github.com/gorilla/handlers"
)

// maintenanceModeDispatcher constructs and returns the blob upload handler for the
// given request context.
func maintenanceModeDispatcher(ctx *Context, r *http.Request) http.Handler {
	mmh := &maintenanceModeHandler{
		Context: ctx,
	}

	return http.Handler(handlers.MethodHandler{
		"POST": mmh,
	})
}

type maintenanceModeHandler struct {
	*Context
}

func (mmh *maintenanceModeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	action := getMaintenanceModeAction(mmh.Context)
	if action == "on" {
		mmh.Context.App.maintenanceMode = true
	} else if action == "off" {
		mmh.Context.App.maintenanceMode = false
	}

	w.WriteHeader(http.StatusOK)
}
