package routes

import (
	"github.com/unfamousthomas/thesis-service/internal/app"
	"github.com/unfamousthomas/thesis-service/internal/handlers"
	"log"
	"net/http"
)

// SetupRoutes is used to define routes and their matching handlers
func SetupRoutes(a *app.App) {

	a.Mux.HandleFunc("POST /server", handlers.CreateServer(a))
	a.Mux.HandleFunc("DELETE /server", handlers.DeleteServer(a))
	a.Mux.HandleFunc("POST /server/pod/labels", handlers.AddPodLabel(a))
	a.Mux.HandleFunc("DELETE /server/pod/labels", handlers.RemovePodLabel(a))

	a.Mux.HandleFunc("POST /fleet", handlers.CreateFleet(a))
	a.Mux.HandleFunc("DELETE /fleet", handlers.DeleteFleet(a))

	a.Mux.HandleFunc("POST /game", handlers.CreateGame(a))
	a.Mux.HandleFunc("DELETE /game", handlers.DeleteGame(a))

	a.Mux.HandleFunc("POST /scaler", handlers.CreateScaler(a))
	a.Mux.HandleFunc("DELETE /scaler", handlers.DeleteScaler(a))

	a.Mux.HandleFunc("/health", handlers.Health(a))
	err := http.ListenAndServe(":8080", a.Mux)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
