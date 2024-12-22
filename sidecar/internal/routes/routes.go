package routes

import (
	"github.com/unfamousthomas/thesis-sidecar/internal/app"
	"github.com/unfamousthomas/thesis-sidecar/internal/handlers"
	"log"
	"net/http"
)

func SetupRoutes(a *app.App) {

	a.Mux.HandleFunc("GET /allow_delete", handlers.IsDeleteAllowed(a))
	a.Mux.HandleFunc("POST /allow_delete", handlers.SetDeleteAllowed(a))
	a.Mux.HandleFunc("GET /delete_requested", handlers.IsShutdownRequestedByOperator(a))
	a.Mux.HandleFunc("POST /delete_requested", handlers.SetShutdownAllowedByServer(a))
	a.Mux.HandleFunc("/health", handlers.Health(a))
	err := http.ListenAndServe(":8080", a.Mux)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
