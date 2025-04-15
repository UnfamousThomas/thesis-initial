package app

import (
	"log/slog"
	"net/http"
	"os"
)

type App struct {
	Mux    *http.ServeMux
	Logger *slog.Logger
}

func CreateApp() *App {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).
		With("service", "example-server")
	slog.SetDefault(logger)
	return &App{
		Mux:    http.NewServeMux(),
		Logger: logger,
	}
}
