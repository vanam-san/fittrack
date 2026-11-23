package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Open(ctx context.Context, path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, fmt.Errorf("create application data directory: %w", err)
	}

	db, err := sql.Open("sqlite", sqliteURI(path, "_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)&_pragma=trusted_schema(OFF)&_txlock=immediate"))
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := migrate(ctx, db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}
	return db, nil
}

func sqliteURI(path, query string) string {
	uriPath := filepath.ToSlash(path)
	if !strings.HasPrefix(uriPath, "/") {
		uriPath = "/" + uriPath
	}
	uri := url.URL{
		Scheme:   "file",
		Path:     uriPath,
		RawQuery: query,
	}
	return uri.String()
}

func migrate(ctx context.Context, db *sql.DB) error {
	files, err := migrations.ReadDir("migrations")
	if err != nil {
		return err
	}

	// An immediate transaction prevents concurrent launches from selecting
	// the same pending migrations.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
        version INTEGER PRIMARY KEY
    )`); err != nil {
		return err
	}
	var current int
	if err := tx.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&current); err != nil {
		return err
	}
	if current > len(files) {
		return fmt.Errorf("schema version %d requires a newer FitTrack release", current)
	}
	for index, file := range files {
		prefix, _, found := strings.Cut(file.Name(), "_")
		version, err := strconv.Atoi(prefix)
		if !found || err != nil || version != index+1 {
			return fmt.Errorf("unexpected migration name %q", file.Name())
		}
		if version <= current {
			continue
		}
		script, err := migrations.ReadFile("migrations/" + file.Name())
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, string(script)); err != nil {
			return fmt.Errorf("apply migration %d: %w", version, err)
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			return err
		}
	}
	return tx.Commit()
}
