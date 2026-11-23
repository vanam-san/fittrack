//go:build fittrack_smoke

package main

import (
	"os"
	"strconv"
)

// The Windows loader ignores WebView2 environment overrides. Supply these
// options only in the smoke binary, which runs with a temporary user profile.
func init() {
	port, err := strconv.ParseUint(os.Getenv("FITTRACK_SMOKE_PORT"), 10, 16)
	dataPath := os.Getenv("FITTRACK_SMOKE_WEBVIEW_DATA")
	if err != nil || port == 0 || dataPath == "" {
		return
	}

	windowsOptions.WebviewUserDataPath = dataPath
	windowsOptions.AdditionalBrowserArgs = []string{
		"--remote-debugging-port=" + strconv.FormatUint(port, 10),
		"--remote-debugging-address=127.0.0.1",
		"--proxy-server=http://127.0.0.1:1",
		"--proxy-bypass-list=wails.localhost;localhost;127.0.0.1",
	}
}
