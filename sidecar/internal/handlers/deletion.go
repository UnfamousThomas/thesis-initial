package handlers

import (
	"encoding/json"
	"github.com/unfamousthomas/thesis-sidecar/internal/app"
	"log"
	"net/http"
)

type DeleteRequest struct {
	Allowed bool `json:"allowed"`
}

// IsDeleteAllowed is used by the operator to check if this can be deleted
func IsDeleteAllowed(a *app.App) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(DeleteRequest{Allowed: a.DeleteAllowed})
		if err != nil {
			log.Printf("Error encoding response: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	})
}

// SetDeleteAllowed is used by the server to tell the operator this can be deleted
func SetDeleteAllowed(a *app.App) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var request DeleteRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			log.Printf("Error decoding request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		a.DeleteAllowed = request.Allowed
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(request)
		if err != nil {
			log.Printf("Error encoding response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}
