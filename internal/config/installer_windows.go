package config

import (
	"errors"
	"fmt"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const storageKey = `Software\FitTrack\Storage`

func installerStorage(defaultDir string) (Storage, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, storageKey, registry.QUERY_VALUE|registry.WOW64_64KEY)
	if errors.Is(err, registry.ErrNotExist) {
		return storagePaths(defaultDir, "", ""), nil
	}
	if err != nil {
		return Storage{}, err
	}
	defer key.Close()

	databaseDir, err := installerPath(key, "DatabaseDirectory")
	if err != nil {
		return Storage{}, err
	}
	if databaseDir == "" {
		databaseDir = defaultDir
	}
	backupDir, err := installerPath(key, "BackupDirectory")
	if err != nil {
		return Storage{}, err
	}
	restore, err := installerPath(key, "RestoreSource")
	if err != nil {
		return Storage{}, err
	}
	return storagePaths(databaseDir, backupDir, restore), nil
}

func installerPath(key registry.Key, name string) (string, error) {
	path, _, err := key.GetStringValue(name)
	if errors.Is(err, registry.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read installer setting %s: %w", name, err)
	}
	if path != "" && !filepath.IsAbs(path) {
		return "", fmt.Errorf("installer setting %s must be an absolute path", name)
	}
	return path, nil
}

// ClaimInstallerRestore consumes the request before touching the database so
// later launches cannot replay a restore over newer records.
func ClaimInstallerRestore(source string) (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, storageKey, registry.QUERY_VALUE|registry.SET_VALUE|registry.WOW64_64KEY)
	if errors.Is(err, registry.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	defer key.Close()
	pending, err := installerPath(key, "RestoreSource")
	if err != nil || pending != source {
		return false, err
	}
	if err := key.DeleteValue("RestoreSource"); err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
