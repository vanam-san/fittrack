//go:build !windows

package config

func installerStorage(defaultDir string) (Storage, error) {
	return storagePaths(defaultDir, "", ""), nil
}

func ClaimInstallerRestore(string) (bool, error) {
	return false, nil
}
