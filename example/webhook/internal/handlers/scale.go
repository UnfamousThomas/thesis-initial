package handlers

import (
	"encoding/json"
	"github.com/UnfamousThomas/loputoo-fake-webhook/internal/app"
	"log"
	"net/http"
)

// SetScalingInfo is used by the user to set the scaling to some amounts
func SetScalingInfo(a *app.App) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var request app.ScalingInfo
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			log.Printf("Error decoding request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if request.Scale != nil {
			a.ScalingInfo.Scale = request.Scale
		}
		if request.Replicas != nil {
			a.ScalingInfo.Replicas = request.Replicas
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(a.ScalingInfo)
		if err != nil {
			log.Printf("Error encoding response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

func GetScalingInfo(a *app.App) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(a.ScalingInfo)
		if err != nil {
			log.Printf("Error encoding response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}
