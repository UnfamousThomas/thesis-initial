package handlers

import (
	"context"
	"encoding/json"
	"github.com/unfamousthomas/thesis-service/internal/app"
	"github.com/unfamousthomas/thesis-service/internal/kube"
	"log"
	"net/http"
)

type CreateFleetRequest struct {
	Fleet *kube.Fleet `json:"fleet"`
}

// CreateFleet is used to create a new server in the cluster
func CreateFleet(a *app.App) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var request CreateFleetRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			log.Printf("Error decoding request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if request.Fleet == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err, obj := kube.CreateFleet(context.WithValue(context.Background(), "kube", "create-fleet"), *request.Fleet, a.DynamicClient)
		if err != nil {
			log.Printf("Error creating fleet: %v", err)
			e := map[string]string{
				"message": "Error creating fleet",
				"error":   err.Error(),
			}
			jsonData, err := json.Marshal(e)
			if err != nil {
				log.Println("Error marshaling json:", err)
				return
			}

			_, err = w.Write(jsonData)
			if err != nil {
				log.Println("Error writing response:", err)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		jsonData, err := json.Marshal(obj)
		if err != nil {
			log.Println("Error marshaling json:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = w.Write(jsonData)
		if err != nil {
			log.Println("Error writing response:", err)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

func DeleteFleet(a *app.App) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var request DeleteObjectRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			log.Printf("Error decoding request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if request.Metadata == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = kube.DeleteFleet(context.WithValue(context.Background(), "kube", "delete-fleet"), *request.Metadata, a.DynamicClient, a.ClientSet, request.Force)
		if err != nil {
			log.Printf("Error deleting fleet: %v\n", err)
			e := map[string]string{
				"message": "Error deleting fleet",
				"error":   err.Error(),
			}
			jsonData, err := json.Marshal(e)
			if err != nil {
				log.Println("Error marshaling json:", err)
				return
			}

			_, err = w.Write(jsonData)
			if err != nil {
				log.Println("Error writing response:", err)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}
