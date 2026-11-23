# FitTrack architecture

FitTrack is a Wails desktop application. In production, the Go executable embeds the Vite frontend. Vue renders the interface in the operating system’s WebView and calls Go through generated Wails bindings. The app reads and writes a local SQLite database.

## Where changes belong

| Location | Responsibility |
| --- | --- |
| `main.go` | Startup, database initialization, desktop window, and Wails service registration |
| `frontend/src/features/` | Dashboard, workout, nutrition, weight, analytics, and settings screens |
| `frontend/src/services/` | Frontend entry points for calling generated Wails bindings |
| `internal/desktop/` | Go services and desktop file dialogs exposed to the frontend |
| `internal/workouts/`, `internal/nutrition/`, `internal/weight/`, and `internal/analytics/` | Domain validation, calculations, and tracking operations |
| `internal/database/` | SQLite migrations, sqlc query sources, and import, export, and backup |
| `build/` | Platform build and installer configuration |

Change the Go service before regenerating Wails bindings. Change SQL in `internal/database/queries/` before regenerating typed queries with `sqlc generate`. For schema changes, add a migration and verify it against a database from an older release.

Workout sessions keep a snapshot of their routine plan when they start, so later routine edits do not rewrite completed history. The workout calendar groups completed sessions by their recorded calendar date. A single date can contain multiple sessions.

See the [user guide](user-guide.md) for the operating system paths and the differences between portable exports and full SQLite backups.
