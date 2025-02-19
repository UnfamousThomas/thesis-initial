package routes

import (
	"github.com/unfamousthomas/thesis-sidecar/internal/app"
	"github.com/unfamousthomas/thesis-sidecar/internal/handlers"
	"log"
	"net/http"
)

func SetupRoutes(a *app.App) {

	a.Mux.HandleFunc("POST /server", handlers.CreateServer(a))
	a.Mux.HandleFunc("/health", handlers.Health(a))
	err := http.ListenAndServe(":8080", a.Mux)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
