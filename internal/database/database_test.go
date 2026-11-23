package database

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenUpgradesPhaseOneDatabase(t *testing.T) {
	ctx := t.Context()
	path := filepath.Join(t.TempDir(), "fittrack.sqlite3")
	legacy, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer legacy.Close()
	script, err := migrations.ReadFile("migrations/001_settings.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := legacy.ExecContext(ctx, string(script)+`
		CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY);
		INSERT INTO schema_migrations VALUES (1);
		INSERT INTO app_settings VALUES (1, 'Sam', 'lb', 'dark', 2000, 100, 200, 50);
	`); err != nil {
		t.Fatal(err)
	}
	if err := legacy.Close(); err != nil {
		t.Fatal(err)
	}

	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var name, unit string
	var goal int
	if err := db.QueryRowContext(ctx, "SELECT display_name, weight_unit, calorie_goal FROM app_settings").
		Scan(&name, &unit, &goal); err != nil || name != "Sam" || unit != "lb" || goal != 2000 {
		t.Fatalf("phase 1 settings after migration: %s, %s, %d, %v", name, unit, goal, err)
	}
	var exercises, logs int
	if err := db.QueryRowContext(ctx, "SELECT (SELECT COUNT(*) FROM exercises), (SELECT COUNT(*) FROM food_logs)").
		Scan(&exercises, &logs); err != nil || exercises == 0 || logs != 0 {
		t.Fatalf("unexpected migrated libraries: exercises=%d logs=%d err=%v", exercises, logs, err)
	}
}

func TestOpenRefusesNewerSchema(t *testing.T) {
	ctx := t.Context()
	path := filepath.Join(t.TempDir(), "fittrack.sqlite3")
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES (999)")
	db.Close()
	if err != nil {
		t.Fatal(err)
	}
	newer, err := Open(ctx, path)
	if newer != nil {
		newer.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "requires a newer FitTrack release") {
		t.Fatalf("expected explicit schema version error, got %v", err)
	}
}
