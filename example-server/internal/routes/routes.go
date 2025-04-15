package routes

import (
	"github.com/UnfamousThomas/thesis-example-server/internal/app"
	"log"
	"net/http"
)

func SetupRoutes(a *app.App) {

	log.Printf("Starting listening on port %d", 8081)
	err := http.ListenAndServe(":8081", a.Mux) //8081 to avoid issues related to the sidecar port conflicts
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
