package config

import (
	"os"
	"path/filepath"
	"runtime"
)

type Storage struct {
	DatabasePath string
	BackupPath   string
	RestorePath  string
}

func StoragePaths() (Storage, error) {
	dir, err := defaultDataDirectory()
	if err != nil {
		return Storage{}, err
	}
	return installerStorage(dir)
}

func defaultDataDirectory() (string, error) {
	var base string
	var err error
	switch runtime.GOOS {
	case "windows":
		base = os.Getenv("LOCALAPPDATA")
		if !filepath.IsAbs(base) {
			base, err = os.UserHomeDir()
			base = filepath.Join(base, "AppData", "Local")
		}
	case "darwin":
		base, err = os.UserConfigDir()
	default:
		base = os.Getenv("XDG_DATA_HOME")
		if !filepath.IsAbs(base) {
			base, err = os.UserHomeDir()
			base = filepath.Join(base, ".local", "share")
		}
	}
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "FitTrack"), nil
}

func storagePaths(databaseDir, backupDir, restore string) Storage {
	if backupDir == "" {
		backupDir = filepath.Join(databaseDir, "backups")
	}
	return Storage{filepath.Join(databaseDir, "fittrack.sqlite3"), filepath.Clean(backupDir), restore}
}
