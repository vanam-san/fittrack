//go:build !android

package database

import (
	"context"
	"database/sql"
	"fmt"

	"modernc.org/sqlite"
)

const (
	sqlitePrimaryQuery  = "_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)&_pragma=trusted_schema(OFF)&_txlock=immediate"
	sqliteReadOnlyQuery = "mode=ro&_pragma=trusted_schema(OFF)&_pragma=busy_timeout(5000)"
	sqliteSnapshotQuery = "mode=rw&_pragma=synchronous(FULL)"
)

func copyDatabase(ctx context.Context, conn *sql.Conn, uri string, restore bool) error {
	return conn.Raw(func(raw any) error {
		driver, ok := raw.(interface {
			NewBackup(string) (*sqlite.Backup, error)
			NewRestore(string) (*sqlite.Backup, error)
		})
		if !ok {
			return fmt.Errorf("SQLite backup is unavailable")
		}
		var backup *sqlite.Backup
		var err error
		if restore {
			backup, err = driver.NewRestore(uri)
		} else {
			backup, err = driver.NewBackup(uri)
		}
		if err != nil {
			return err
		}
		return runBackup(ctx, func(pages int) (bool, error) {
			more, err := backup.Step(int32(pages))
			return !more, err
		}, backup.Finish)
	})
}
