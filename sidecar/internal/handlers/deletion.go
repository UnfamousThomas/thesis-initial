package handlers

import (
	"encoding/json"
	"github.com/unfamousthomas/thesis-sidecar/internal/app"
	"log"
	"net/http"
)

var delete = false

type deleteRequest struct {
	Allowed bool `json:"allowed"`
}

func IsDeleteAllowed(a *app.App) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(deleteRequest{Allowed: delete})
		if err != nil {
			log.Printf("Error encoding response: %v", err)
			return
		}
	})
}

func SetDeleteAllowed(a *app.App) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var request deleteRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			log.Printf("Error decoding request: %v", err)
			return
		}
		delete = request.Allowed
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(request)
		if err != nil {
			log.Printf("Error encoding response: %v", err)
			return
		}
	})
}
