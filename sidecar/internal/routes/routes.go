package routes

import (
	"github.com/unfamousthomas/thesis-sidecar/internal/app"
	"github.com/unfamousthomas/thesis-sidecar/internal/handlers"
	"log"
	"net/http"
)

// SetupRoutes sets up the nessecary routes, their handlers and starts serving http.
func SetupRoutes(a *app.App) {

	a.Mux.HandleFunc("GET /allow_delete", handlers.IsDeleteAllowed(a))
	a.Mux.HandleFunc("POST /allow_delete", handlers.SetDeleteAllowed(a))
	a.Mux.HandleFunc("GET /shutdown", handlers.IsShutdownRequested(a))
	a.Mux.HandleFunc("POST /shutdown", handlers.SetShutdownRequested(a))
	a.Mux.HandleFunc("/health", handlers.Health(a))
	err := http.ListenAndServe(":8080", a.Mux)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
