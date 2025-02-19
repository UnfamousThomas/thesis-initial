package handlers

import (
	"context"
	"encoding/json"
	"github.com/unfamousthomas/thesis-sidecar/internal/app"
	"github.com/unfamousthomas/thesis-sidecar/internal/kube"
	"log"
	"net/http"
)

type CreateServerRequest struct {
	Server *kube.Server `json:"server"`
}

// CreateServer is used to create a new server in the cluster
func CreateServer(a *app.App) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var request CreateServerRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			log.Printf("Error decoding request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if request.Server == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = kube.CreateServer(context.WithValue(context.Background(), "kube", "create-server"), *request.Server, a.DynamicClient)
		if err != nil {
			log.Printf("Error creating server: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}
