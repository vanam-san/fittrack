package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/vanam-san/fittrack/internal/analytics"
	"github.com/vanam-san/fittrack/internal/config"
	"github.com/vanam-san/fittrack/internal/database"
	"github.com/vanam-san/fittrack/internal/desktop"
	"github.com/vanam-san/fittrack/internal/nutrition"
	"github.com/vanam-san/fittrack/internal/settings"
	"github.com/vanam-san/fittrack/internal/weight"
	"github.com/vanam-san/fittrack/internal/workouts"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Wails serves the Vite bundle from the executable in production.
//
//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

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
	services := []application.Service{}
	storage, err := config.StoragePaths()
	var restoreError string
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		db, err = database.Open(ctx, storage.DatabasePath)
		cancel()
	}
	if err != nil {
		slog.Error("could not initialize local storage", "error", err)
	} else {
		defer db.Close()
		files := database.NewDataFiles(db, storage.DatabasePath, storage.BackupPath)
		if storage.RestorePath != "" {
			restoreError = restoreInstallerBackup(files, storage.RestorePath)
		}
		preferences = settings.NewService(db)
		services = append(services, application.NewService(desktop.NewWorkoutService(workouts.NewService(db))))
		services = append(services, application.NewService(desktop.NewNutritionService(nutrition.NewService(db))))
		services = append(services, application.NewService(desktop.NewWeightService(weight.NewService(db))))
		services = append(services, application.NewService(desktop.NewAnalyticsService(analytics.NewService(db))))
		services = append(services, application.NewService(desktop.NewDataFilesService(files, storage.BackupPath)))
	}
	// Keep Settings available so the UI can report a storage startup failure.
	services = append(services, application.NewService(desktop.NewSettingsService(preferences, storage, restoreError)))

	app := application.New(application.Options{
		Name:        "FitTrack",
		Description: "Offline fitness tracking",
		Icon:        appIcon,
		Services:    services,
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

func restoreInstallerBackup(files *database.DataFiles, source string) string {
	claimed, err := config.ClaimInstallerRestore(source)
	if err != nil {
		return fmt.Sprintf("FitTrack couldn't start the installer restore. %v", err)
	}
	if !claimed {
		return ""
	}

	checkpoint, err := files.Restore(context.Background(), source)
	if err == nil {
		return ""
	}
	message := fmt.Sprintf("FitTrack couldn't restore the installer backup. %v. You can retry below.", err)
	if checkpoint != "" {
		message += " Recovery backup saved to " + checkpoint
	}
	return message
}
