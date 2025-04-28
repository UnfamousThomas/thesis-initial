package handlers

import (
	"context"
	"encoding/json"
	"github.com/unfamousthomas/thesis-service/internal/app"
	"github.com/unfamousthomas/thesis-service/internal/kube"
	"log"
	"net/http"
)

type AddPodLabelsRequest struct {
	Metadata *kube.Metadata `json:"metadata"`
}

// AddPodLabel is used to add a new label to a pod
func AddPodLabel(a *app.App) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var request AddPodLabelsRequest
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

		err = kube.AddPodLabels(context.WithValue(context.Background(), "job", "update-pod-label"), *request.Metadata, a.ClientSet)

		if err != nil {
			log.Printf("Error adding pod labels: %v", err)
			e := map[string]string{
				"message": "Error adding pod labels",
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
	})
}

type RemovePodLabelRequest struct {
	Metadata *kube.Metadata `json:"metadata"`
	Label    *string        `json:"label"`
}

// RemovePodLabel is used to remove a label from a pod
func RemovePodLabel(a *app.App) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var request RemovePodLabelRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			log.Printf("Error decoding request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if request.Metadata == nil || request.Label == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = kube.RemovePodLabel(context.WithValue(context.Background(), "job", "remove-pod-label"), *request.Metadata, *request.Label, a.ClientSet)

		if err != nil {
			log.Printf("Error adding pod label: %v", err)
			e := map[string]string{
				"message": "Error removing pod label",
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
	})
}
