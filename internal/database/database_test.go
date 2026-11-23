package database

import (
	"path/filepath"
	"strings"
	"testing"
)

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
