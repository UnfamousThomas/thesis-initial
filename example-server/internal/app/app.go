package app

import (
	"log/slog"
	"net/http"
	"os"
)

type App struct {
	Mux    *http.ServeMux
	Logger *slog.Logger
	State  *ServerState
}

type ServerState struct {
	ServerRunning bool
	StopRequested bool
	StopAllowed   bool
}

func CreateApp() *App {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).
		With("service", "example-server")
	slog.SetDefault(logger)
	return &App{
		Mux:    http.NewServeMux(),
		Logger: logger,
		State: &ServerState{
			ServerRunning: false,
			StopRequested: false,
			StopAllowed:   false,
		},
	}
}
