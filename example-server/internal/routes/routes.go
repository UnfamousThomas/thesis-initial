package routes

import (
	"github.com/UnfamousThomas/thesis-example-server/internal/app"
	"github.com/UnfamousThomas/thesis-example-server/internal/handlers"
	"log"
	"net/http"
)

func SetupRoutes(a *app.App) {

	a.Mux.HandleFunc("POST /server", handlers.ServerNewStateRequest(a))
	err := http.ListenAndServe(":8080", a.Mux)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
