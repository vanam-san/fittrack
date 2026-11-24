package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type DataFiles struct {
	db        *sql.DB
	path      string
	backupDir string
}

func NewDataFiles(db *sql.DB, path, backupDir string) *DataFiles {
	return &DataFiles{db: db, path: path, backupDir: backupDir}
}

func (d *DataFiles) Export(ctx context.Context, path, format string) error {
	if err := d.checkPath(path); err != nil {
		return err
	}
	// Read linked tables from one snapshot, then write the file outside the transaction.
	tx, err := d.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	data, err := readRecords(ctx, tx)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return writeFile(path, func(file *os.File) error {
		if err := writeArchive(file, format, data); err != nil {
			return err
		}
		return file.Sync()
	})
}

func (d *DataFiles) Import(ctx context.Context, path string) (string, error) {
	if err := d.checkPath(path); err != nil {
		return "", err
	}
	if err := checkInput(path); err != nil {
		return "", err
	}
	format := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	data, err := readArchive(file, format)
	file.Close()
	if err != nil {
		return "", fmt.Errorf("read import: %w", err)
	}

	conn, err := d.db.Conn(ctx)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	// Keep the same connection through the checkpoint and replacement so other
	// operations in this instance cannot write between them.
	checkpoint, err := d.checkpoint(ctx, conn, "pre-import")
	if err != nil {
		return "", fmt.Errorf("save recovery backup: %w", err)
	}
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return checkpoint, err
	}
	defer tx.Rollback()
	if err := replaceRecords(ctx, tx, data); err != nil {
		return checkpoint, fmt.Errorf("invalid import: %w", err)
	}
	return checkpoint, tx.Commit()
}

func (d *DataFiles) Backup(ctx context.Context, path string) error {
	if err := d.checkPath(path); err != nil {
		return err
	}
	conn, err := d.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	return backupFile(ctx, conn, path)
}

func (d *DataFiles) Restore(ctx context.Context, path string) (string, error) {
	if err := d.checkPath(path); err != nil {
		return "", err
	}
	if err := checkInput(path); err != nil {
		return "", err
	}

	// Migrate and validate a staged copy so the original backup stays reusable.
	stage, err := os.CreateTemp(filepath.Dir(d.path), ".restore-*.sqlite3")
	if err != nil {
		return "", err
	}
	stagePath := stage.Name()
	defer os.Remove(stagePath)
	if err := stage.Close(); err != nil {
		return "", err
	}

	candidate, err := stageBackup(ctx, path, stagePath)
	if err != nil {
		return "", fmt.Errorf("invalid backup: %w", err)
	}
	defer candidate.Close()
	conn, err := d.db.Conn(ctx)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	if err := validateBackup(ctx, candidate, conn); err != nil {
		return "", fmt.Errorf("invalid backup: %w", err)
	}
	if err := candidate.Close(); err != nil {
		return "", err
	}
	checkpoint, err := d.checkpoint(ctx, conn, "pre-restore")
	if err != nil {
		return "", fmt.Errorf("save recovery backup: %w", err)
	}
	return checkpoint, copyDatabase(ctx, conn, sqliteURI(stagePath, "mode=ro"), true)
}

func stageBackup(ctx context.Context, path, stage string) (*sql.DB, error) {
	source, err := sql.Open("sqlite", sqliteURI(path, sqliteReadOnlyQuery))
	if err != nil {
		return nil, err
	}
	defer source.Close()
	conn, err := source.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	var count, minVersion, maxVersion int
	if err := conn.QueryRowContext(ctx, "SELECT COUNT(*), COALESCE(MIN(version),0), COALESCE(MAX(version),0) FROM schema_migrations").Scan(&count, &minVersion, &maxVersion); err != nil {
		return nil, err
	}
	if minVersion != 1 || count != maxVersion {
		return nil, fmt.Errorf("invalid FitTrack migration history")
	}
	if err := copyDatabase(ctx, conn, sqliteURI(stage, "mode=rw"), false); err != nil {
		return nil, err
	}
	return Open(ctx, stage)
}

func validateBackup(ctx context.Context, candidate *sql.DB, live *sql.Conn) error {
	var integrity string
	if err := candidate.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&integrity); err != nil {
		return err
	}
	if integrity != "ok" {
		return fmt.Errorf("SQLite integrity check failed")
	}

	actual, err := schema(ctx, candidate)
	if err != nil {
		return err
	}
	expected, err := schema(ctx, live)
	if err != nil {
		return err
	}
	if !slices.Equal(actual, expected) {
		return fmt.Errorf("the backup schema does not match FitTrack")
	}

	rows, err := candidate.QueryContext(ctx, "PRAGMA foreign_key_check")
	if err != nil {
		return err
	}
	hasOrphans := rows.Next()
	rowsErr := rows.Err()
	rows.Close()
	if rowsErr != nil {
		return rowsErr
	}
	if hasOrphans {
		return fmt.Errorf("the backup contains records with missing references")
	}
	_, err = readRecords(ctx, candidate)
	return err
}

func schema(ctx context.Context, db queryer) ([][4]string, error) {
	rows, err := db.QueryContext(ctx, "SELECT type, name, tbl_name, sql FROM sqlite_schema WHERE name NOT GLOB 'sqlite_*' ORDER BY type, name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result [][4]string
	for rows.Next() {
		var entry [4]string
		if err := rows.Scan(&entry[0], &entry[1], &entry[2], &entry[3]); err != nil {
			return nil, err
		}
		result = append(result, entry)
	}
	return result, rows.Err()
}

func (d *DataFiles) checkpoint(ctx context.Context, conn *sql.Conn, prefix string) (string, error) {
	dir := d.backupDir
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	file, err := os.CreateTemp(dir, prefix+"-"+time.Now().Format("20060102-150405")+"-*.sqlite3")
	if err != nil {
		return "", err
	}
	if err := file.Close(); err != nil {
		os.Remove(file.Name())
		return "", err
	}
	if err := backupFile(ctx, conn, file.Name()); err != nil {
		os.Remove(file.Name())
		return "", err
	}
	return file.Name(), nil
}

func backupFile(ctx context.Context, conn *sql.Conn, path string) error {
	return writeFile(path, func(file *os.File) error {
		stage := file.Name()
		if err := file.Close(); err != nil {
			return err
		}
		return copyDatabase(ctx, conn, sqliteURI(stage, sqliteSnapshotQuery), false)
	})
}

func runBackup(ctx context.Context, step func(int) (bool, error), finish func() error) (result error) {
	defer func() { result = errors.Join(result, finish()) }()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		done, err := step(256)
		if err != nil || done {
			return err
		}
	}
}

func writeFile(path string, write func(*os.File) error) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".fittrack-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if err := write(file); err != nil {
		return err
	}
	if err := file.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
		return err
	}
	return os.Rename(file.Name(), path)
}

func checkInput(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("choose a regular file")
	}
	return nil
}

func (d *DataFiles) checkPath(path string) error {
	if !filepath.IsAbs(path) {
		return fmt.Errorf("choose an absolute file location")
	}
	selected, err := os.Stat(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	for _, suffix := range []string{"", "-journal", "-wal", "-shm"} {
		activePath := d.path + suffix
		active, err := os.Stat(activePath)
		if filepath.Clean(path) == filepath.Clean(activePath) ||
			(err == nil && selected != nil && os.SameFile(selected, active)) {
			return fmt.Errorf("choose a file outside the active database and its sidecar files")
		}
	}
	if selected != nil && !selected.Mode().IsRegular() {
		return fmt.Errorf("choose a regular file")
	}
	return nil
}
