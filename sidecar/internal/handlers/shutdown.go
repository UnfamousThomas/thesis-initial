package handlers

import (
	"encoding/json"
	"github.com/unfamousthomas/thesis-sidecar/internal/app"
	"log"
	"net/http"
)

type ShutdownRequest struct {
	Shutdown bool `json:"shutdown"`
}

// IsShutdownRequested is used by the gameserver to check for shutdown requests
func IsShutdownRequested(a *app.App) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(ShutdownRequest{Shutdown: a.ShutdownRequested})
		if err != nil {
			log.Printf("Error encoding response: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	})
}

// SetShutdownRequested is used by the operator to request shutdowns
func SetShutdownRequested(a *app.App) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var request ShutdownRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			log.Printf("Error decoding request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		a.ShutdownRequested = request.Shutdown
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(request)
		if err != nil {
			log.Printf("Error encoding response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}
