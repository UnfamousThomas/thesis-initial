package handlers

import (
	"context"
	"encoding/json"
	"github.com/unfamousthomas/thesis-service/internal/app"
	"github.com/unfamousthomas/thesis-service/internal/kube"

	"log"
	"net/http"
)

type CreateServerRequest struct {
	Server *kube.Server `json:"server"`
}

type DeleteObjectRequest struct {
	Metadata *kube.Metadata `json:"metadata"`
	Force    bool           `json:"force"`
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
		server := request.Server
		err = kube.CreateServer(context.WithValue(context.Background(), "kube", "create-server"), server, a.DynamicClient)
		if err != nil {
			log.Printf("Error creating server: %v", err)
			e := map[string]string{
				"message": "Error creating server",
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
		jsonData, err := json.Marshal(server)
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

// DeleteServer is used to delete an existing Server from the cluster, based on the namespace and name
func DeleteServer(a *app.App) func(http.ResponseWriter, *http.Request) {
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

		err = kube.DeleteServer(context.WithValue(context.Background(), "kube", "delete-server"), *request.Metadata, a.DynamicClient, a.ClientSet, request.Force)
		if err != nil {
			log.Printf("Error deleting server: %v\n", err)
			e := map[string]string{
				"message": "Error creating server",
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
