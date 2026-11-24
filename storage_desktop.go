package main

import "github.com/vanam-san/fittrack/internal/config"

func storagePaths() (config.Storage, error) {
	return config.StoragePaths()
}
