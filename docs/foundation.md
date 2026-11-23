# Desktop foundation

FitTrack was generated with Wails v3.1.0 from the [Wails Vue Tailwind TypeScript v1.1.0 release](https://github.com/vanam-san/wails-template-vue-tailwind/releases/tag/v1.1.0), commit `d7f04e5f84cbd9b077764967c92d373fac4687e9`. The source was extracted into a separate temporary app before the native scaffold was brought into FitTrack. The starter's Bun workflow, Vue and Tailwind setup, shadcn-vue UI primitives, binding generation, and embedded asset layout carry forward. The starter's MIT notice is included in `LICENSE`.

The Windows build used Go 1.27.4, Node 24.22.3, and Bun 1.4.5. The frontend retains the starter's Vue 3.5.46, Vite 8.3.8, TypeScript 5.9.3, Tailwind 4.3.5, shadcn-vue 2.9.0, and `lucide-vue-next` 1.0.0 pins. `frontend/bun.lock` also records the development-only Vue Test Utils and Happy DOM dependencies.

## Layout

- `main.go` resolves storage, runs migrations at startup, and registers the Wails settings service.
- `internal/desktop` exposes settings data to the generated bindings. `internal/settings` owns validation and calls typed sqlc queries.
- `internal/database` opens SQLite, embeds and applies migrations, and contains the SQL source and generated query code. The pure-Go driver is `modernc.org/sqlite` v1.63.0. Regenerate queries with `sqlc generate` using sqlc v1.31.1.
- `frontend/src/services` imports CLI-generated bindings. The Vue dashboard and settings form use app-owned wrappers. Native build and packaging tasks are under `build/`.

## Data and upgrades

Migration `001_settings.sql` creates the settings table. The app writes its single row when first-run setup is saved. Database startup creates the migration ledger and applies pending SQL in an immediate transaction. A failed migration rolls back. A database from a newer schema is rejected before writes. New migrations should be added in sequence and tested against a copy of the prior schema.

Goals are whole kcal or grams. Zero means unset. Phase 2 can extend the schema with SQLite migrations and sqlc queries for exercises, routines, and sessions. Session dates and timestamps, tracking types, and snapshot fields should be settled before creating those tables.
