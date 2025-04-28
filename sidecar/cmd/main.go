package main

import (
	"github.com/unfamousthomas/thesis-sidecar/internal/app"
	"github.com/unfamousthomas/thesis-sidecar/internal/routes"
	"net/http"
)

func main() {
	a := app.App{
		Mux:               http.NewServeMux(),
		ShutdownRequested: false,
		DeleteAllowed:     false,
	}

	routes.SetupRoutes(&a)
}
