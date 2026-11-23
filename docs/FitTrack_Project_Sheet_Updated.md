# FitTrack --- Project Sheet

> **Purpose:** Authoritative implementation brief for building FitTrack
> as a desktop application on the **Wails v3 + Vue 3 + Vite +
> TypeScript + Tailwind CSS v4 + shadcn-vue** starter-template
> architecture.
>
> **Product direction:** A private, local-first fitness tracker for
> Windows, macOS, and Linux.
>
> **Architecture:** One Wails desktop application, one Go backend, one
> Vue frontend, and one local SQLite database. No interactive terminal
> UI.
>
> **Runtime requirement:** Fully usable offline. No accounts, cloud
> services, external APIs, telemetry, or mandatory network access.

------------------------------------------------------------------------

## 1. Product summary

FitTrack brings four areas of personal fitness tracking into one desktop
application:

1.  **Nutrition** --- manually log meals and drinks, calories, and
    macros; save reusable food presets.
2.  **Workouts** --- maintain an exercise library, build routines, log
    live sessions and sets, and review progression.
3.  **Weight** --- record body weight, set a goal, and view progress
    over time.
4.  **Analytics** --- review daily, weekly, and monthly summaries,
    charts, training consistency, and personal records.

All records stay on the user's computer. The application does not depend
on an online food database, workout service, account, or synchronization
provider.

## 2. Foundation template and technology decisions

FitTrack should be generated from, or built directly on, the supplied
**Wails v3 + Vue 3 + Vite + TypeScript + Tailwind CSS v4** template.

The reference template is a reusable desktop starter foundation.
FitTrack is a separate product application that consumes and extends
that foundation; do not turn the FitTrack application repository into a
second copy of the template generator unless explicitly requested.

  -----------------------------------------------------------------------
  Layer                   Selection               FitTrack usage
  ----------------------- ----------------------- -----------------------
  Desktop framework       Wails v3                Native desktop shell
                                                  and Go integration

  Backend                 Go                      Domain logic,
                                                  validation,
                                                  persistence,
                                                  calculations

  Frontend framework      Vue 3                   Desktop UI using
                                                  Composition API and
                                                  `<script setup>`

  Frontend build          Vite                    Development server and
                                                  production bundling

  Language                TypeScript              Strictly typed frontend

  Styling                 Tailwind CSS v4         CSS-first design tokens
                                                  and utility styling

  UI primitives           shadcn-vue              Locally-owned,
                                                  accessible Vue
                                                  components

  Icons                   lucide-vue-next         Locally bundled icons

  Package manager         Bun                     Reproducible frontend
                                                  dependency management

  Database                SQLite                  Local persistent
                                                  storage

  SQLite driver           `modernc.org/sqlite`    Pure-Go SQLite option;
                                                  confirm compatibility
                                                  with chosen Go/Wails
                                                  versions

  Database access         `database/sql` + sqlc   Typed query layer

  Migrations              Versioned SQL           Safe schema evolution
                          migrations              

  Charts                  Apache ECharts          Interactive, locally
                                                  bundled analytics

  Logging                 Go `slog`               Diagnostics without
                                                  logging sensitive
                                                  fitness values

  Tests                   Go tests, Vitest, Vue   Backend, frontend, and
                          Test Utils              integration coverage

  Packaging               Wails v3 build pipeline Native platform
                                                  bundles/installers
  -----------------------------------------------------------------------

FitTrack starts from Wails Vue Tailwind TypeScript v1.1.0 
(commit d7f04e5f84cbd9b077764967c92d373fac4687e9). Use Bun 1.4.5 as 
pinned by the starter and install dependencies with bun install --frozen-lockfile. 
Node remains a development tool for frontend scripts that use it.

### Template conventions to preserve

-   Use **Wails v3 only**. Do not copy Wails v2 APIs, `wails.json`, or
    v2-specific conventions.
-   Use the template's official Wails v3 project structure, build
    assets, Taskfile/tasks, service registration, binding generation,
    and frontend embedding conventions.
-   Use Vue 3 + Vite as a client-rendered SPA. **Do not introduce Nuxt
    or another SSR/server framework.**
-   Use Tailwind CSS v4 through `@tailwindcss/vite`, not a Tailwind v3
    PostCSS configuration.
-   Use shadcn-vue components as local source files. Do not load UI
    components from a CDN.
-   Use Lucide Vue icons from the installed package.
-   Use the Bun version pinned by the starter and commit frontend/bun.lock.
-   Keep Wails-generated bindings generated by the Wails CLI. Never
    hand-edit generated binding files.
-   Keep the frontend build output and Wails embedded asset path
    consistent with the active template scaffold.

### Why desktop rather than terminal?

FitTrack relies on forms, tables, calendars, exercise history, and
interactive charts. These workflows are easier to use and understand in
a desktop interface. Maintaining both Bubble Tea and Wails would
duplicate navigation, screen design, testing, accessibility work, and
release checks.

**Do not include in the initial product:** Bubble Tea, Lip Gloss,
Bubbles, Cobra, terminal charts, Docker, or Docker Compose. A small
optional CLI may be considered later for scripted backup or automation,
but it is not part of the first release.

## 3. Product principles

-   **Offline first:** Every user-facing feature works without an
    internet connection.
-   **Local ownership:** The user can inspect, back up, export, and
    restore their data.
-   **One source of truth:** All calculations and business rules live in
    Go services, not frontend components.
-   **No duplicate implementations:** Vue displays data and collects
    input; Go validates, calculates, and persists it.
-   **Simple by default:** Manual food entry and reusable presets
    replace online food lookups.
-   **Safe data handling:** Migrations, imports, restores, and
    destructive actions must be recoverable and clearly explained.
-   **Original design:** Workout tracking may take inspiration from
    common fitness-app workflows, but must not copy Hevy's branding,
    assets, proprietary interface, or implementation.
-   **Keyboard-friendly desktop UX:** Use natural mouse interaction
    while supporting fast keyboard-driven entry and navigation.

## 4. Scope

### 4.1 Core release (must ship)

**Application foundation** - First-run setup and a local
profile/settings screen; no account or sign-in. - Dashboard with today's
nutrition, latest weight, and workout summary. - Persistent local SQLite
database and automatic migrations. - kg/lb preference and configurable
daily calorie/macronutrient goals. - Clear empty states, validation,
confirmations, and useful error messages. - Desktop navigation,
light/dark theme, and responsive layouts suitable for common desktop
window sizes.

**Nutrition** - Add meals, snacks, and drinks to a selected date. -
Manually enter food name, serving quantity, calories, protein,
carbohydrates, and fat. - Save frequently used foods as reusable
presets. - Edit and delete entries; browse daily history. - Calculate
meal and daily totals locally.

**Workouts** - Built-in starter exercise library plus custom
exercises. - Exercise attributes: name, primary/secondary muscle groups,
equipment, tracking type, and instructions/notes. - Create, edit,
duplicate, and delete workout routines. - Order exercises and configure
planned sets, reps, and rest duration. - Start, pause, resume, and
complete a workout session. - Log sets with weight, reps, completion
state, and notes. - Show previous performance for the exercise while
logging. - Rest timer and recovery of an unfinished workout after
application restart. - Workout history and session detail view.

**Weight** - Add, edit, and delete dated weight measurements. - Set and
edit a target weight. - Display current weight, target, change over
time, and a simple trend chart.

**Basic analytics** - Daily, weekly, and monthly calorie and macro
summaries. - Weight progression chart. - Workout frequency and duration
summaries. - Exercise progression by weight/reps and training volume. -
Personal records derived from saved session data.

**Data safety** - Export and import CSV/JSON. - Create a complete
database backup from the application. - Restore from a validated backup
with explicit confirmation. - Explain where the local database and
backups are stored.

### 4.2 Post-core enhancements

Do not block the first usable release on these:

-   Warm-up, drop, and failure set labels.

-   Supersets and linked exercises.

-   Exercise substitutions.

-   Weekly workout schedule and calendar heatmap.

-   Muscle-group distribution and training statistics.

-   Estimated one-rep max using the Epley formula for suitable strength
    sets:

    `estimated_1RM = weight * (1 + reps / 30)`

    This is an estimate, not a tested maximum. Label it clearly in the
    UI.

-   Advanced date filters and chart comparisons.

-   Optional CLI for automation, only if a real need emerges.

### 4.3 Explicit non-goals

-   Mobile application or web-hosted application.
-   User accounts, authentication, cloud sync, or multi-device
    synchronization.
-   Online food/nutrition lookup or external fitness APIs.
-   Hevy integration or API dependency.
-   Social features, leaderboards, coaching, or AI recommendations.
-   Telemetry, advertising, remote analytics, or remote configuration.
-   Mandatory Docker installation or a database server.
-   Nuxt, SSR, Nitro server routes, or a Node.js server in the packaged
    application.
-   A second terminal/TUI interface in the initial release.

## 5. Application architecture

``` text
┌────────────────────────────────────────────────────────────┐
│                     FitTrack Desktop                        │
│                                                            │
│  Wails v3 native desktop shell                              │
│  └── Embedded Vue 3 + Vite + TypeScript frontend            │
│      ├── Tailwind CSS v4                                    │
│      ├── shadcn-vue local components                        │
│      ├── Lucide Vue icons                                   │
│      ├── Dashboard / Nutrition / Workouts / Weight           │
│      └── History / Analytics / Settings                     │
└────────────────────────────┬───────────────────────────────┘
                             │ Wails v3 generated bindings
┌────────────────────────────▼───────────────────────────────┐
│                       Go application                        │
│                                                            │
│  Application services                                      │
│  ├── Nutrition service                                     │
│  ├── Workout service                                       │
│  ├── Weight service                                        │
│  ├── Analytics service                                     │
│  └── Import / Export / Backup service                      │
│                                                            │
│  Validation · Domain rules · Calculations                   │
│  Repository/query layer                                    │
└────────────────────────────┬───────────────────────────────┘
                             │
┌────────────────────────────▼───────────────────────────────┐
│                 Local SQLite database                      │
│        Versioned migrations · Constraints · Indexes         │
└────────────────────────────────────────────────────────────┘
```

### Architectural rules

1.  Frontend code must not connect to SQLite directly.
2.  Frontend code must not independently calculate calories, macros,
    volume, personal records, or estimated 1RM.
3.  Wails bindings should expose small, task-oriented service methods
    and return typed data-transfer objects.
4.  Domain packages must not import Wails or frontend-specific code.
5.  Repository methods own persistence; services own validation and
    business rules.
6.  Database migrations are versioned. Never modify an already-released
    migration; add a new migration instead.
7.  Embed production frontend assets in the desktop binary. No CDN,
    remote font, remote image, or runtime asset download.
8.  Store the database in the OS-standard per-user application-data
    directory, not beside the executable or in a temporary directory.
9.  Store dates consistently. Use a local calendar date for day-based
    logs and timestamps for event ordering; document the convention and
    test date-boundary behavior.
10. Derive personal records from workout history initially. Add a cache
    only if performance measurements justify it.
11. Use transactions when creating/completing sessions, importing
    records, or restoring data.
12. Keep Wails service methods thin; call application services rather
    than putting business logic in binding methods.
13. Keep the frontend independent of generated binding file layout by
    using app-owned TypeScript service wrappers.
14. Keep the app functional with network access disabled.
    Development-time dependency installation may require internet, but
    runtime must not.

## 6. FitTrack repository layout

Adapt the exact Wails v3 scaffold rather than replacing it. Preserve all
generated platform assets, build files, Taskfile conventions, and
metadata required by the current stable Wails release.

``` text
fittrack/
├── cmd/                         # Only if the selected Wails scaffold uses this layout
├── internal/
│   ├── app/                     # Startup and dependency wiring
│   ├── config/                  # Local paths and application settings
│   ├── database/
│   │   ├── migrations/          # Versioned SQL migrations
│   │   ├── queries/             # sqlc query source
│   │   └── generated/           # Generated query code
│   ├── nutrition/               # Food library and meal logging
│   ├── workout/                 # Routines, sessions, sets, history
│   ├── exercise/                # Exercise catalogue and muscle groups
│   ├── weight/                  # Measurements and goals
│   ├── analytics/               # Reports and calculations
│   ├── dataio/                  # CSV, JSON, backup, and restore
│   └── desktop/                 # Wails services, bindings-facing DTOs
├── frontend/
│   ├── src/
│   │   ├── main.ts
│   │   ├── App.vue
│   │   ├── style.css
│   │   ├── components/
│   │   │   ├── layout/
│   │   │   └── ui/              # Locally generated shadcn-vue components
│   │   ├── views/
│   │   ├── features/
│   │   │   ├── dashboard/
│   │   │   ├── nutrition/
│   │   │   ├── workouts/
│   │   │   ├── weight/
│   │   │   ├── analytics/
│   │   │   └── settings/
│   │   ├── composables/
│   │   ├── services/            # Typed wrappers around generated bindings
│   │   ├── lib/
│   │   └── types/
│   ├── vite.config.ts
│   ├── package.json
│   ├── bun.lock
│   └── tsconfig*.json
├── build/                       # Wails-generated build/installer assets
├── docs/
├── go.mod
├── go.sum
└── README.md
```

Do not create empty folders that serve no purpose. Use Vue Router or
Pinia only if the selected app structure genuinely needs them; they are
not mandatory solely because the template supports Vue.

## 7. Frontend implementation requirements

### 7.1 Vue + Vite

-   Use Vue 3 Composition API and `<script setup lang="ts">`.
-   Use Vite as the frontend development server and production bundler.
-   Keep the Vite configuration, Wails dev-server port integration,
    asset base path, and output directory consistent with the supplied
    template and current Wails v3 documentation.
-   Do not introduce Nuxt, SSR, Nitro, or a runtime Node server.
-   Use strict TypeScript and `vue-tsc` for Vue-aware type checking.
-   Avoid `any`; isolate and explain any unavoidable use.
-   Keep generated Wails bindings separate from app-owned services.
-   Do not manually modify generated binding files.

### 7.2 Tailwind CSS v4

-   Use `tailwindcss` and `@tailwindcss/vite`.
-   Use the official Tailwind v4 Vite plugin.
-   Start the global stylesheet with `@import "tailwindcss";`.
-   Use CSS-first theme configuration and centralise design tokens in
    the global stylesheet.
-   Do not add a Tailwind v3 PostCSS configuration or legacy config file
    unless a verified compatibility requirement makes it necessary.
-   Verify light and dark themes in the actual Wails webview.

### 7.3 shadcn-vue and icons

-   Use shadcn-vue's current Vite setup and CLI.
-   Keep generated component source in the repository under the
    configured UI components directory.
-   Add only components needed by each feature; likely candidates
    include button, input, textarea, card, badge, dialog, dropdown-menu,
    select, tabs, table, calendar/date picker, tooltip, separator, and
    toast/sonner.
-   Do not fetch components or styles from a CDN at runtime.
-   Use `lucide-vue-next` for icons and bundle icons locally.
-   Keep components accessible, keyboard-operable, and visually
    consistent.

### 7.4 Frontend service layer

Create app-owned TypeScript wrappers around generated Wails bindings,
for example:

``` text
src/services/
├── nutrition.ts
├── workouts.ts
├── weight.ts
├── analytics.ts
└── data.ts
```

Vue components should call these wrappers rather than importing
generated binding paths throughout the UI. Keep request/response types
explicit and handle backend errors consistently.

### 7.5 Charts

Use Apache ECharts (or another locally bundled chart library only if a
concrete reason is documented). Charts must: - Work without internet
access. - Display units, date ranges, and meaningful labels. - Have
empty and loading states. - Use Go-calculated datasets and metrics. -
Avoid misleading or fabricated sample values. - Support resizing with
the desktop window.

## 8. Data model outline

Use normalized relational tables, foreign keys, suitable indexes, and
transactions for multi-record operations. Exact columns should be
finalized during schema design.

  -----------------------------------------------------------------------
  Table                               Purpose
  ----------------------------------- -----------------------------------
  `app_settings`                      Units, daily goals, preferences

  `foods`                             Reusable manually entered food
                                      presets and nutrient values

  `food_logs`                         Food consumed, serving amount,
                                      date, meal category, nutrient
                                      snapshot

  `exercises`                         Built-in and custom exercise
                                      definitions

  `muscle_groups`                     Standard muscle-group catalogue

  `exercise_muscles`                  Primary and secondary
                                      exercise-to-muscle mapping

  `workout_routines`                  Reusable workout templates

  `routine_exercises`                 Exercise ordering and planned
                                      set/repetition/rest defaults

  `workout_schedule`                  Optional weekday-to-routine
                                      assignments (post-core)

  `workout_sessions`                  Actual workout sessions and session
                                      status

  `session_exercises`                 Exercises performed in a session
                                      and their ordering

  `workout_sets`                      Weight, reps, duration/distance
                                      where applicable, set type,
                                      completion, notes

  `weight_entries`                    Dated body-weight measurements
  -----------------------------------------------------------------------

Important data rules: - A routine is a template; a workout session is a
historical event. Editing a routine must not rewrite completed
sessions. - Store nutrient values as a snapshot on the food log so
changing a reusable food preset does not silently rewrite historical
intake. - Preserve unfinished workout session state in SQLite so it
survives application restarts. - Use transactions when
creating/completing sessions, importing records, or restoring data. -
Use consistent internal units and convert for display; never mix kg and
lb in calculations. - Validate numeric ranges and reject invalid or
non-finite values.

## 9. UX direction

Use a clean, calm desktop layout with persistent navigation and a clear
hierarchy:

-   **Dashboard:** today's calories/macros, latest weight, next/planned
    workout if configured, and recent activity.
-   **Nutrition:** date selector, meal sections, quick-add food, saved
    food presets, daily totals.
-   **Workouts:** start workout, routines, exercise library, and
    history.
-   **Live workout:** focused exercise cards, editable set rows,
    previous performance, rest timer, add-set, finish-workout action.
-   **Weight:** measurement entry, goal progress, and historical trend.
-   **Analytics:** date filters, interactive charts, and concise
    summaries.
-   **Settings:** units, goals, database location, backup, restore,
    import, and export.

Interaction requirements: - Prioritize quick keyboard entry for
repetitive workout and food logging while keeping mouse interaction
natural. - Make destructive actions reversible where practical and
confirm permanent deletion. - Clearly distinguish empty data, loading,
validation errors, and saved state. - Do not make a dashboard look
populated with fake sample metrics. Sample data, if offered, must be
clearly labelled and removable. - Keep charts informative but
restrained; include units, date ranges, and empty states.

## 10. Delivery phases

### Phase 0 --- Template and repository audit

-   Inspect the supplied Wails v3 Vue/Vite/Tailwind template and the
    target FitTrack repository.
-   Confirm the template's current stable Wails v3 conventions, build
    tasks, binding generation, asset embedding, and frontend output
    path.
-   Check the latest stable compatible versions of Wails, Go, Node, Vue,
    Vite, TypeScript, Tailwind, shadcn-vue, Lucide Vue, Bun, and SQLite
    dependencies.
-   Generate or initialise a clean FitTrack project from the template in
    a separate working directory if appropriate.
-   Preserve useful existing FitTrack work and user data; do not
    overwrite the template repository.
-   Create a feature branch and record architecture decisions.
-   Produce a concise implementation checklist and data migration plan.

**Exit criteria:** template integration is proven, repository structure
is understood, and no unresolved UI or runtime architecture remains.

### Phase 1 --- Desktop foundation

-   Set up the Wails v3 desktop entry point and service registration
    using the template's conventions.
-   Configure Vue 3 + Vite + TypeScript, Tailwind v4, shadcn-vue, Lucide
    Vue, and Bun according to the template.
-   Add application-data directory resolution and SQLite initialization.
-   Add embedded migrations, repository/query layer, logging, and error
    handling.
-   Build the app shell, navigation, settings, and empty-state
    dashboard.
-   Generate Wails bindings using the official CLI; never hand-write
    generated bindings.
-   Embed the production frontend assets using the template's Wails v3
    mechanism.
-   Add automated offline launch and database persistence tests.

**Exit criteria:** app launches on the target OS, creates its local
database, migrates it, closes cleanly, and reopens with data intact
while disconnected from the network.

### Phase 2 --- Workout tracking

-   Exercise library and custom exercise creation.
-   Routine builder and routine editing.
-   Start/resume/complete session flow.
-   Set logging, previous performance, rest timer, and workout history.
-   Persist an in-progress session after each meaningful change.

**Exit criteria:** user can create a routine, log a complete workout,
restart the app, and review the saved workout accurately.

### Phase 3 --- Nutrition tracking

-   Food preset library and manual food entry.
-   Meal categories and daily log.
-   Nutrient snapshots and daily calorie/macro totals.
-   Food history editing and deletion.

**Exit criteria:** user can save a food preset, log it on multiple
dates, change the preset without altering previous logs, and see correct
totals.

### Phase 4 --- Weight tracking

-   Measurement entry and history.
-   Goal settings and unit conversion.
-   Weight trend chart.

**Exit criteria:** measurements persist, display in the selected unit,
and chart correctly over a selected date range.

### Phase 5 --- Analytics

-   Nutrition summaries.
-   Workout frequency, duration, and volume.
-   Exercise progression and personal records.
-   Basic date filtering and interactive charts.

**Exit criteria:** analytics are calculated in Go, reconcile with source
records, and update after edits/deletions.

### Phase 6 --- Import, export, backup, and restore

-   CSV/JSON export and import with documented formats.
-   Full backup and restore workflow.
-   Validate imported data before committing it.
-   Protect existing data by creating a pre-restore backup.

**Exit criteria:** a user can export, restore on a clean installation,
and recover the same records without network access.

### Phase 7 --- Release hardening

-   Test Windows, macOS, and Linux builds on their appropriate build
    environments.
-   Package installers/bundles and a portable release where appropriate.
-   Test fresh installation, upgrade, migration, uninstall/reinstall,
    and backup recovery.
-   Verify all production frontend assets are embedded.
-   Complete README, user guide, troubleshooting, and release notes.

**Exit criteria:** release artifacts are reproducible, core workflows
pass, and the application runs fully offline after installation. Do not
claim signing/notarisation unless actually configured and tested.

### Phase 8 --- Optional enhancements

Only after the core release is stable: supersets, advanced set types,
workout heatmap, muscle distribution, substitutions, advanced
chart comparisons, and optional CLI automation.

### Phase 9 --- Workout logging and history

FitTrack already covers several workflows in the FitNotes listing through its
routine editor, previous-session set reference, distance and duration tracking,
set notes, rest timer, and local backup and export. [FitNotes on Google Play](https://play.google.com/store/apps/details?id=com.github.jamesgay.fitnotes)

- Add a calendar to workout history that highlights completed workout dates and
  lists all saved sessions for the selected day.
- Add newer and older workout navigation within the current history results.
- Add **Save and add another** to the custom exercise form.

These features use existing workout date and exercise records and require no
database migration. Exercise-specific calendar filters, workout reordering during
a session, and timer alerts are candidates for separate work.

## 11. Offline and privacy requirements

The shipped application must not initiate network connections during
normal use.

Prohibited at runtime: - External APIs, food databases, or fitness
integrations. - Cloud storage, synchronization, accounts, or
authentication services. - Telemetry, analytics, crash-report uploads,
or advertising SDKs. - CDN scripts/styles, remotely hosted fonts,
images, or chart assets. - Remote configuration or mandatory update
checks.

Required safeguards: - Bundle frontend assets, fonts (if any), and chart
libraries locally. - Do not add a runtime network client unless a future
product decision explicitly changes the offline requirement. - Keep
fitness records and food logs out of routine diagnostic logs. - Test
core functionality with network access disabled. - Inspect dependencies
for unexpected network or telemetry behavior. - Document that
downloading/installing updates may require internet access, but using
the installed app does not.

## 12. Testing and definition of done

A feature is complete only when: - Its business rules and validation are
implemented in Go. - Its data is stored through the repository layer and
survives restart. - Its UI has loading, empty, success, and error states
as applicable. - Unit tests cover calculations and validation. -
Persistence/integration tests cover important database behavior. - It
works without internet access. - It does not break import/export or
historical data integrity. - User-facing labels, units, and date ranges
are clear. - Relevant documentation is updated.

### Release acceptance checklist

-   [ ] App launches with the network disabled.
-   [ ] First launch creates the application-data directory and SQLite
    database.
-   [ ] Migrations run safely on a fresh database and an existing
    database.
-   [ ] Nutrition, workout, and weight records persist after restart.
-   [ ] An unfinished workout can be resumed after restart.
-   [ ] Historical food logs do not change when a food preset is edited.
-   [ ] Routine edits do not change completed workout sessions.
-   [ ] kg/lb conversion and displayed units are consistent.
-   [ ] Analytics reconcile with saved records after edits and
    deletions.
-   [ ] CSV/JSON import/export and full backup/restore work offline.
-   [ ] Restore validates input and protects existing data.
-   [ ] Wails v3 bindings are generated using the official CLI and are
    not manually edited.
-   [ ] Vue 3 + Vite production build succeeds and outputs to the path
    expected by the Wails scaffold.
-   [ ] Tailwind CSS v4 styles render correctly in the packaged Wails
    webview.
-   [ ] shadcn-vue components are local source files and work without a
    network connection.
-   [ ] Lucide Vue icons and ECharts assets are bundled locally.
-   [ ] Production assets are embedded; no runtime CDN or external
    request is needed.
-   [ ] No Nuxt/Nitro server or Node.js runtime is required by the
    installed application.
-   [ ] No account, cloud sync, telemetry, or external API dependency
    exists.
-   [ ] Windows, macOS, and Linux packaging is tested on supported build
    environments.

## 13. Instructions for the implementation agent

1.  Read this project sheet and inspect the actual FitTrack repository
    and supplied starter template before making changes.
2.  Treat the supplied Wails v3 + Vue 3 + Vite + TypeScript + Tailwind
    CSS v4 template as the frontend/desktop foundation and preserve its
    valid conventions.
3.  Use the current official Wails v3 custom-template/scaffold
    conventions. Do not invent Wails APIs, task names, template fields,
    or CLI flags.
4.  Check the latest stable Wails v3 release and compatible toolchain
    versions at implementation time. Do not assume the versions in an
    old template brief are still current.
5.  Work in a dedicated feature branch. Preserve existing user data and
    avoid destructive migrations.
6.  Do not introduce Nuxt, SSR, Bubble Tea, Docker deployment, online
    services, or a second database unless the project owner explicitly
    changes the decision.
7.  Implement one phase at a time. Keep each change small enough to
    review and test.
8.  Before changing schema, explain the migration impact and add a
    migration; never silently discard data.
9.  Keep business logic out of Vue components and Wails binding methods.
10. Generate Wails bindings with the official CLI; never hand-author or
    edit generated binding files.
11. Keep shadcn-vue components in the project source and avoid runtime
    CDN dependencies.
12. Add or update tests with every feature. Do not claim a test, build,
    offline check, or installer was run unless it actually was.
13. At the end of each phase, report changed files, decisions,
    commands/tests run, results, known limitations, and the next phase's
    prerequisites.
14. Do not use fake production data, hard-coded analytics, placeholder
    success states, or silent error suppression.
15. Keep the application local-first and offline by default; treat any
    proposed network dependency as an architecture change requiring
    explicit approval.

## 14. Final architecture decision

**Build FitTrack as one Wails v3 desktop application using the supplied
Vue 3 + Vite + TypeScript + Tailwind CSS v4 + shadcn-vue template,
backed by Go services and one local SQLite database.**

The FitTrack application should reuse the template's Wails v3
service/binding, frontend build, asset embedding, and packaging
conventions, then add the nutrition, workout, weight, analytics, and
local data-management features. Do not add Nuxt or Bubble Tea. This
provides a single maintainable desktop experience with interactive
charts, fast data entry, local data ownership, and fully offline runtime
operation.

## 15. Reference links

Use official documentation and re-check versions and commands at
implementation time:

-   Wails v3: https://v3.wails.io/
-   Wails v3 releases: https://github.com/wailsapp/wails/releases
-   Wails v3 custom templates:
    https://wails.io/guides/advanced/custom-templates/
-   Wails frontend framework integration:
    https://wails.io/guides/dev/frontend-frameworks/
-   Vue 3: https://vuejs.org/
-   Vite: https://vite.dev/
-   TypeScript: https://www.typescriptlang.org/
-   Tailwind CSS v4 with Vite:
    https://tailwindcss.com/docs/installation/using-vite
-   shadcn-vue: https://shadcn-vue.com/docs/installation/vite
-   Lucide Vue: https://lucide.dev/guide/packages/lucide-vue-next
-   Bun: https://bun.sh/docs/pm/cli/install
-   Apache ECharts: https://echarts.apache.org/
-   SQLite: https://www.sqlite.org/docs.html

------------------------------------------------------------------------

## 16. Implementation handoff report

At the end of implementation, report:

1.  Repository tree summary and how the starter template was reused.
2.  Selected stable versions of Wails, Go, Node, Vue, Vite, TypeScript,
    Tailwind, shadcn-vue, Lucide Vue, Bun, and SQLite dependencies.
3.  Commands used to initialise/configure the FitTrack project.
4.  Commands run for formatting, linting, type checking, tests, and
    builds, with pass/fail status.
5.  Platforms actually tested.
6.  Any issue that could not be verified locally.
7.  Any remaining maintainer placeholders.

**Definition of done:** FitTrack launches in Wails v3 development mode,
calls Go services from Vue through generated bindings, stores and
retrieves fitness data in SQLite, passes the documented checks, and
builds a standalone desktop application for the current host platform.
