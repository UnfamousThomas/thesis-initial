package main

import (
	"github.com/UnfamousThomas/loputoo-fake-webhook/internal/app"
	"github.com/UnfamousThomas/loputoo-fake-webhook/internal/routes"
	"net/http"
)

var initialReplicas = 1
var initialScale = true

func main() {
	a := app.App{
		Mux: http.NewServeMux(),
		ScalingInfo: app.ScalingInfo{
			Replicas: &initialReplicas,
			Scale:    &initialScale,
		},
	}
	routes.SetupRoutes(&a)

}
