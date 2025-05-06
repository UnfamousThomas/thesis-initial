package app

import "net/http"

// App struct is where the state of is stored, along with the used http Mux.
type App struct {
	Mux         *http.ServeMux
	ScalingInfo ScalingInfo
}

type ScalingInfo struct {
	Replicas *int  `json:"desired_replicas"`
	Scale    *bool `json:"scale"`
}
