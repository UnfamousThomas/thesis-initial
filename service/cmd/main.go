package main

import (
	"github.com/unfamousthomas/thesis-sidecar/internal/app"
	"github.com/unfamousthomas/thesis-sidecar/internal/routes"
)

func main() {
	a := app.CreateApp()
	routes.SetupRoutes(a)
}
