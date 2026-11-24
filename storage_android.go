package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vanam-san/fittrack/internal/config"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func storagePaths() (config.Storage, error) {
	dir := application.Android.StoragePath()
	if !filepath.IsAbs(dir) {
		return config.Storage{}, fmt.Errorf("Android private storage is unavailable")
	}
	// Clear exports left behind if Android stopped the app during a file picker.
	_ = os.RemoveAll(filepath.Join(dir, "exports"))
	return config.Storage{
		DatabasePath: filepath.Join(dir, "fittrack.sqlite3"),
		BackupPath:   filepath.Join(dir, "backups"),
	}, nil
}
