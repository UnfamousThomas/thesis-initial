package main

import (
	"github.com/unfamousthomas/thesis-service/internal/app"
	"github.com/unfamousthomas/thesis-service/internal/routes"
)

func main() {
	a := app.CreateApp()
	routes.SetupRoutes(a)
}
