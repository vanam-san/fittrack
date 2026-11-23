package main

import (
	"context"
	"database/sql"
	"embed"
	"log/slog"
	"os"
	"time"

	"github.com/vanam-san/fittrack/internal/config"
	"github.com/vanam-san/fittrack/internal/database"
	"github.com/vanam-san/fittrack/internal/desktop"
	"github.com/vanam-san/fittrack/internal/settings"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Wails serves the Vite bundle from the executable in production.
//
//go:embed all:frontend/dist
var assets embed.FS

var windowsOptions application.WindowsOptions

func main() {
	if err := run(); err != nil {
		slog.Error("FitTrack could not run", "error", err)
		os.Exit(1)
	}
}

func run() error {
	var db *sql.DB
	var preferences *settings.Service
	path, err := config.DatabasePath()
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		db, err = database.Open(ctx, path)
		cancel()
	}
	if err != nil {
		slog.Error("could not initialize local storage", "error", err)
	} else {
		defer db.Close()
		preferences = settings.NewService(db)
	}

	app := application.New(application.Options{
		Name:        "FitTrack",
		Description: "Offline fitness tracking",
		Services: []application.Service{
			application.NewService(desktop.NewSettingsService(preferences, path)),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
		Windows: windowsOptions,
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "FitTrack",
		Width:     1120,
		Height:    760,
		MinWidth:  380,
		MinHeight: 580,
		URL:       "/",
	})

	return app.Run()
}
