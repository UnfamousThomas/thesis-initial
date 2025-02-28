package handlers

import (
	"context"
	"encoding/json"
	"github.com/unfamousthomas/thesis-service/internal/app"
	"github.com/unfamousthomas/thesis-service/internal/kube"
	"log"
	"net/http"
)

type CreateScalerRequest struct {
	Scaler *kube.GameAutoscaler `json:"scaler"`
}

// CreateScaler is used to create a new kube.GameAutoscaler in the cluster
func CreateScaler(a *app.App) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var request CreateScalerRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			log.Printf("Error decoding request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if request.Scaler == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err, obj := kube.CreateScaler(context.WithValue(context.Background(), "kube", "create-scaler"), *request.Scaler, a.DynamicClient)
		if err != nil {
			log.Printf("Error creating scaler: %v", err)
			e := map[string]string{
				"message": "Error creating scaler",
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

func DeleteScaler(a *app.App) func(http.ResponseWriter, *http.Request) {
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

		err = kube.DeleteScaler(context.WithValue(context.Background(), "kube", "delete-scaler"), *request.Metadata, a.DynamicClient, a.ClientSet)
		if err != nil {
			log.Printf("Error deleting scaler: %v\n", err)
			e := map[string]string{
				"message": "Error deleting scaler",
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
