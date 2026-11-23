package config

import (
	"os"
	"path/filepath"
	"runtime"
)

func DatabasePath() (string, error) {
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
	return filepath.Join(base, "FitTrack", "fittrack.sqlite3"), nil
}
