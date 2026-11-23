package settings_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/vanam-san/fittrack/internal/database"
	"github.com/vanam-san/fittrack/internal/settings"
)

func TestPreferencesSurviveRestart(t *testing.T) {
	ctx := t.Context()
	path := filepath.Join(t.TempDir(), "FitTrack # test", "fittrack.sqlite3")
	db, err := database.Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	service := settings.NewService(db)
	defaults, complete, err := service.Load(ctx)
	if err != nil || complete || defaults.WeightUnit != "kg" || defaults.Theme != "system" {
		t.Fatalf("unexpected first-run state: %+v, complete=%t, err=%v", defaults, complete, err)
	}

	input := settings.Preferences{
		DisplayName: "  Sam  ", WeightUnit: "lb", Theme: "dark",
		CalorieGoal: 2100, ProteinGoal: 130, CarbohydrateGoal: 240, FatGoal: 65,
	}
	saved, err := service.Save(ctx, input)
	if err != nil || saved.DisplayName != "Sam" {
		t.Fatalf("unexpected save result: %+v, err=%v", saved, err)
	}
	invalid := saved
	invalid.ProteinGoal = -1
	_, err = service.Save(ctx, invalid)
	var validation settings.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	var initialMigrations int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations").Scan(&initialMigrations); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("database was not created at the selected path: %v", err)
	}

	reopened, err := database.Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { reopened.Close() })
	loaded, complete, err := settings.NewService(reopened).Load(ctx)
	if err != nil || !complete || loaded != saved {
		t.Fatalf("reopened state: got %+v, complete=%t, err=%v, want %+v", loaded, complete, err, saved)
	}
	var migrations int
	if err := reopened.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations").Scan(&migrations); err != nil || migrations != initialMigrations {
		t.Fatalf("reopen should not reapply migration: count=%d, err=%v", migrations, err)
	}
}
