package handlers

import (
	"context"
	"encoding/json"
	"github.com/unfamousthomas/thesis-service/internal/app"
	"github.com/unfamousthomas/thesis-service/internal/kube"
	"log"
	"net/http"
)

type CreateGameRequest struct {
	Game *kube.GameType `json:"game"`
}

// CreateGame is used to create a new kube.GameType in the cluster
func CreateGame(a *app.App) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var request CreateGameRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			log.Printf("Error decoding request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if request.Game == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		game := request.Game
		err = kube.CreateGame(context.WithValue(context.Background(), "kube", "create-game"), game, a.DynamicClient)
		if err != nil {
			log.Printf("Error creating game: %v", err)
			e := map[string]string{
				"message": "Error creating game",
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
		jsonData, err := json.Marshal(game)
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

// DeleteGame is used to delete a game from the cluster, using the namespace and name
func DeleteGame(a *app.App) func(http.ResponseWriter, *http.Request) {
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

		err = kube.DeleteGame(context.WithValue(context.Background(), "kube", "delete-game"), *request.Metadata, a.DynamicClient, a.ClientSet, request.Force)
		if err != nil {
			log.Printf("Error deleting game: %v\n", err)
			e := map[string]string{
				"message": "Error deleting game",
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
