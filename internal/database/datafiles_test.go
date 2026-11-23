package database

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func openDataFiles(t *testing.T) (*sql.DB, *DataFiles) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fittrack.sqlite3")
	db, err := Open(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db, NewDataFiles(db, path)
}

func execData(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(t.Context(), query, args...); err != nil {
		t.Fatal(err)
	}
}

func seedData(t *testing.T, db *sql.DB) {
	t.Helper()
	execData(t, db, `
INSERT INTO app_settings VALUES (1, 'Sam', 'lb', 'dark', 2050, 100, 230, 70);
INSERT INTO exercises VALUES (42, 'Custom exercise', 'Back', '', 'Band', 'weight_reps', '', 0);
INSERT INTO workout_routines VALUES (5, 'Strength');
INSERT INTO routine_exercises VALUES (5, 0, 42, 3, 8, 90);
INSERT INTO workout_sessions VALUES (8, 'Strength', '2026-11-23', 1234, 'paused', 0, 0, 1100, 0, 500);
INSERT INTO session_exercises VALUES (9, 8, 0, 42, 'Custom exercise', 'weight_reps', 3, 8, 90);
INSERT INTO workout_sets VALUES (11, 9, 0, 42.25, 8, 0, 0, 1, '');
INSERT INTO foods VALUES (14, 'Yogurt', '100 g', 120, 9, 12, 4);
INSERT INTO food_logs VALUES (15, 14, '2026-11-22', 'snack', 'Yogurt', '100 g', 1.25, 120, 9, 12, 4);
INSERT INTO food_logs VALUES (16, NULL, '2026-11-23', 'drink', 'Tea', '1 cup', 1, 0, 0, 0, 0);
INSERT INTO weight_measurements VALUES (400, '2026-11-23', 81.25);
INSERT INTO weight_goal VALUES (1, 78);`)
	execData(t, db, "UPDATE workout_sets SET notes = ?", "Form, \"good\"\r\nLine 2 ☕ \\r")
}

func assertRecords(t *testing.T, db *sql.DB, want records) {
	t.Helper()
	actual, err := readRecords(t.Context(), db)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actual, want) {
		t.Fatalf("records differ:\n got %#v\nwant %#v", actual, want)
	}
}

func snapshot(t *testing.T, db *sql.DB) records {
	t.Helper()
	data, err := readRecords(t.Context(), db)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestPortableFilesRoundTrip(t *testing.T) {
	source, files := openDataFiles(t)
	seedData(t, source)
	want := snapshot(t, source)

	for _, format := range []string{"csv", "json"} {
		t.Run(format, func(t *testing.T) {
			dst, target := openDataFiles(t)
			path := filepath.Join(t.TempDir(), "FitTrack records."+format)
			if err := files.Export(t.Context(), path, format); err != nil {
				t.Fatal(err)
			}
			checkpoint, err := target.Import(t.Context(), path)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(checkpoint); err != nil {
				t.Fatal(err)
			}
			assertRecords(t, dst, want)
			if err := dst.Close(); err != nil {
				t.Fatal(err)
			}
			dst, err = Open(t.Context(), target.path)
			if err != nil {
				t.Fatal(err)
			}
			defer dst.Close()
			assertRecords(t, dst, want)
			again := filepath.Join(t.TempDir(), "again."+format)
			if err := NewDataFiles(dst, target.path).Export(t.Context(), again, format); err != nil {
				t.Fatal(err)
			}
			a, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			b, err := os.ReadFile(again)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(a, b) {
				t.Fatal("re-export changed archive")
			}
		})
	}
}

func TestMalformedImportRollsBack(t *testing.T) {
	db, files := openDataFiles(t)
	seedData(t, db)
	want := snapshot(t, db)
	broken := snapshot(t, db)
	broken["app_settings"][0]["display_name"] = "Should roll back"
	broken["food_logs"][1]["food_id"] = int64(9999)

	for _, format := range []string{"csv", "json"} {
		t.Run(format, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "invalid."+format)
			var encoded bytes.Buffer
			if err := writeArchive(&encoded, format, broken); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, encoded.Bytes(), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := files.Import(t.Context(), path); err == nil || !strings.Contains(err.Error(), "food_logs") {
				t.Fatalf("expected an invalid reference, got %v", err)
			}
			assertRecords(t, db, want)
		})
	}
}

func TestBackupRestoreKeepsRecoveryCopy(t *testing.T) {
	db, files := openDataFiles(t)
	seedData(t, db)
	want := snapshot(t, db)
	// The full backup preserves SQLite's sequence even after a row is deleted.
	execData(t, db, "UPDATE sqlite_sequence SET seq = 999 WHERE name = 'weight_measurements'")
	path := filepath.Join(t.TempDir(), "user backup #1.sqlite3")
	if err := files.Backup(t.Context(), path); err != nil {
		t.Fatal(err)
	}
	if err := files.Backup(t.Context(), path); err != nil {
		t.Fatal(err)
	}
	execData(t, db, "UPDATE app_settings SET display_name = 'Newer'")
	newer := snapshot(t, db)
	checkpoint, err := files.Restore(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	assertRecords(t, db, want)
	result, err := db.ExecContext(t.Context(), "INSERT INTO weight_measurements (measured_on, weight_kg) VALUES ('2026-11-24', 80)")
	if err != nil {
		t.Fatal(err)
	}
	if id, err := result.LastInsertId(); err != nil || id != 1000 {
		t.Fatalf("restored sequence produced ID %d: %v", id, err)
	}
	recovery, err := Open(t.Context(), checkpoint)
	if err != nil {
		t.Fatal(err)
	}
	defer recovery.Close()
	assertRecords(t, recovery, newer)

	fresh, target := openDataFiles(t)
	if _, err := target.Restore(t.Context(), path); err != nil {
		t.Fatal(err)
	}
	if err := fresh.Close(); err != nil {
		t.Fatal(err)
	}
	fresh, err = Open(t.Context(), target.path)
	if err != nil {
		t.Fatal(err)
	}
	defer fresh.Close()
	assertRecords(t, fresh, want)
}

func TestRestoreMigratesOlderBackup(t *testing.T) {
	old, files := openDataFiles(t)
	seedData(t, old)
	execData(t, old, "DROP TABLE weight_measurements; DROP TABLE weight_goal; DROP INDEX completed_workout_date; DELETE FROM schema_migrations WHERE version = 4")
	backup := filepath.Join(t.TempDir(), "v3.sqlite3")
	if err := files.Backup(t.Context(), backup); err != nil {
		t.Fatal(err)
	}
	db, target := openDataFiles(t)
	if _, err := target.Restore(t.Context(), backup); err != nil {
		t.Fatal(err)
	}
	data := snapshot(t, db)
	if len(data["weight_measurements"]) != 0 || data["app_settings"][0]["display_name"] != "Sam" {
		t.Fatal("restore did not migrate the original records")
	}
}

func TestInvalidAndBlockedRestorePreserveExistingData(t *testing.T) {
	db, files := openDataFiles(t)
	seedData(t, db)
	want := snapshot(t, db)
	invalid := filepath.Join(t.TempDir(), "broken.sqlite3")
	if err := os.WriteFile(invalid, []byte("invalid SQLite data"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := files.Restore(t.Context(), invalid); err == nil {
		t.Fatal("accepted an invalid backup")
	}
	assertRecords(t, db, want)
	if err := files.Backup(t.Context(), files.path); err == nil {
		t.Fatal("accepted the active database as a backup target")
	}

	source, other := openDataFiles(t)
	execData(t, source, "INSERT INTO app_settings VALUES (1, 'Different', 'kg', 'system', 0, 0, 0, 0)")
	valid := filepath.Join(t.TempDir(), "valid.sqlite3")
	if err := other.Backup(t.Context(), valid); err != nil {
		t.Fatal(err)
	}
	// An external read transaction prevents SQLite from committing the restore.
	// The app retains its original database and the pre-restore backup.
	reader, err := sql.Open("sqlite", sqliteURI(files.path, "mode=ro"))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	execData(t, db, "PRAGMA busy_timeout = 20")
	lock, err := reader.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Rollback()
	var name string
	if err := lock.QueryRowContext(t.Context(), "SELECT display_name FROM app_settings").Scan(&name); err != nil {
		t.Fatal(err)
	}
	checkpoint, err := files.Restore(t.Context(), valid)
	if err == nil || checkpoint == "" {
		t.Fatalf("expected restore to fail with a recovery backup, got %q, %v", checkpoint, err)
	}
	if err := lock.Rollback(); err != nil {
		t.Fatal(err)
	}
	assertRecords(t, db, want)
	recovery, err := Open(t.Context(), checkpoint)
	if err != nil {
		t.Fatal(err)
	}
	defer recovery.Close()
	assertRecords(t, recovery, want)
}
