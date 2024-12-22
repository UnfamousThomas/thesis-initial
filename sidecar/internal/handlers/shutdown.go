package handlers

import (
	"encoding/json"
	"github.com/unfamousthomas/thesis-sidecar/internal/app"
	"log"
	"net/http"
)

var shutdown = false

type shutdownRequest struct {
	Shutdown bool `json:"shutdown"`
}

func IsShutdownRequestedByOperator(a *app.App) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(shutdownRequest{Shutdown: shutdown})
		if err != nil {
			log.Printf("Error encoding response: %v", err)
			return
		}
	})
}

func SetShutdownAllowedByServer(a *app.App) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var request shutdownRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			log.Printf("Error decoding request: %v", err)
			return
		}
		shutdown = request.Shutdown
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(request)
		if err != nil {
			log.Printf("Error encoding response: %v", err)
			return
		}
	})
}
