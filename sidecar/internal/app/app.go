package app

import "net/http"

type App struct {
	Mux *http.ServeMux
}
