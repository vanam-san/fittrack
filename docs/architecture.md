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

## Follow a data flow

- **Logging a set** starts in `SetEditor.vue`, which keeps unsaved field edits locally. `WorkoutsView.vue` calls the workout service and reloads the saved session. The Go service saves the set and starts rest in one transaction when a draft first becomes logged. The countdown uses that saved deadline, so it can resume after reopening the app.
- **Reviewing history** uses the routine snapshot captured at workout start, so later routine edits do not rewrite the session. The calendar groups completed sessions by their recorded date and can display multiple workouts for one day.
- **Logging food from a preset** copies the current per-serving nutrients into the journal row. The journal and analytics multiply those stored values by the serving quantity. This keeps past totals stable when a preset is edited.
- **Displaying weight** converts the stored kilogram value at the UI boundary. If an edit leaves the rounded display value unchanged, the form sends the original kilogram value to avoid conversion drift.
- **Recovering data** uses a native file picker followed by a confirmation in the UI. The Go storage layer saves a recovery snapshot before replacing the active records. Backup restore first migrates and validates a temporary copy.

See the [user guide](user-guide.md) for the operating system paths and the differences between portable exports and full SQLite backups.
