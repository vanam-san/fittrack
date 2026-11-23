# FitTrack

FitTrack is a local desktop fitness tracker. The current foundation includes first-run setup, a dashboard with empty states, and settings for a display name, weight units, appearance, and daily nutrition goals.

## Development

Install Go 1.26 or newer, Node.js 24 or 26, Bun 1.4.5, and the native dependencies for your Wails platform. Install the Wails v3.1.0 CLI and make it available on your `PATH`.

```sh
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.1.0
wails3 doctor
wails3 dev
```

Wails runs Vite on loopback in development and embeds `frontend/dist` in the production binary. `wails3 build` runs the platform build and puts the executable in `bin/`. The frontend dependencies are pinned in `frontend/bun.lock`.

## Checks

Run `wails3 build` before the Go checks so the embedded frontend exists.

```sh
wails3 build
go test . ./internal/...
go vet . ./internal/...
cd frontend
bun run format:check
bun run lint
bun run typecheck
bun run test
```

On Windows, build the smoke executable after `wails3 build` so it uses the production frontend. The smoke tag enables the loopback test connection and blocks external WebView traffic. The test uses a temporary user data directory, completes setup, changes a setting, and verifies both saves after restart.

```powershell
go build -tags 'production,fittrack_smoke' -o bin/fittrack-smoke.exe .
node scripts/smoke-windows.mjs
```

## Local data

Settings shows the SQLite database path. Windows uses `%LOCALAPPDATA%\FitTrack`, macOS uses `~/Library/Application Support/FitTrack`, and Linux uses `$XDG_DATA_HOME/FitTrack` or `~/.local/share/FitTrack`.

See the [project brief](docs/FitTrack_Project_Sheet_Updated.md) and [foundation decisions](docs/foundation.md) for the implementation scope and migration strategy.
