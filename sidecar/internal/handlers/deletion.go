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

// IsDeleteAllowed is used by the operator to check if this can be deleted
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

// SetDeleteAllowed is used by the server to tell the operator this can be deleted
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
