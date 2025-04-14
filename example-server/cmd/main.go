package main

import (
	"github.com/UnfamousThomas/thesis-example-server/internal/app"
	"github.com/UnfamousThomas/thesis-example-server/internal/routes"
)

func main() {
	a := app.CreateApp()
	routes.SetupRoutes(a)
}
