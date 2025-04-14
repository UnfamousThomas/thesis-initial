package handlers

import (
	"encoding/json"
	"github.com/UnfamousThomas/thesis-example-server/internal/app"
	"log"
	"net/http"
)

type ServerState struct {
	Started           *bool `json:"started"`
	ShutdownRequested *bool `json:"shutdown_requested"`
	ShutdownAllowed   *bool `json:"shutdown_allowed"`
}

func ServerNewStateRequest(a *app.App) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var request ServerState
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			log.Printf("Error decoding request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if request.Started != nil {
			a.State.ServerRunning = *request.Started
			log.Printf("Server started state %v", request.Started)
		}

		if request.ShutdownRequested != nil {
			a.State.StopRequested = *request.ShutdownRequested
			log.Printf("Server shutdown requested state %v", request.ShutdownRequested)
		}

		if request.ShutdownAllowed != nil {
			a.State.StopAllowed = *request.ShutdownAllowed
			log.Printf("Server shutdown allowed state %v", request.ShutdownAllowed)
		}

		resp := ServerState{
			Started:           &a.State.ServerRunning,
			ShutdownRequested: &a.State.StopRequested,
			ShutdownAllowed:   &a.State.StopAllowed,
		}

		data, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error encoding response: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err = w.Write(data)
		if err != nil {
			log.Printf("Error encoding response: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})
}
