package handlers

import (
	"github.com/unfamousthomas/thesis-service/internal/app"
	"net/http"
)

func Health(a *app.App) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
