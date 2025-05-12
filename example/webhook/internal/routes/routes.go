package routes

import (
	"github.com/UnfamousThomas/loputoo-fake-webhook/internal/app"
	"github.com/UnfamousThomas/loputoo-fake-webhook/internal/handlers"
	"log"
	"net/http"
)

func SetupRoutes(a *app.App) {
	a.Mux.HandleFunc("POST /scale", handlers.SetScalingInfo(a))
	a.Mux.HandleFunc("GET /scale", handlers.GetScalingInfo(a))
	a.Mux.HandleFunc("/health", handlers.Health(a))
	err := http.ListenAndServe(":8080", a.Mux)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
