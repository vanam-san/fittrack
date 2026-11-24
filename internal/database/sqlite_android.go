package database

import (
	"context"
	"database/sql"

	"github.com/mattn/go-sqlite3"
)

const (
	sqlitePrimaryQuery  = "_foreign_keys=on&_busy_timeout=5000&_synchronous=FULL&_txlock=immediate"
	sqliteReadOnlyQuery = "mode=ro&_busy_timeout=5000"
	sqliteSnapshotQuery = "mode=rw&_synchronous=FULL"
)

func init() {
	// Use the NDK's libc on Android, where raw Linux syscalls can be blocked.
	sql.Register("sqlite", &sqlite3.SQLiteDriver{ConnectHook: func(conn *sqlite3.SQLiteConn) error {
		_, err := conn.Exec("PRAGMA trusted_schema=OFF", nil)
		return err
	}})
}

func copyDatabase(ctx context.Context, conn *sql.Conn, uri string, restore bool) error {
	other, err := sql.Open("sqlite", uri)
	if err != nil {
		return err
	}
	defer other.Close()
	otherConn, err := other.Conn(ctx)
	if err != nil {
		return err
	}
	defer otherConn.Close()

	return conn.Raw(func(primary any) error {
		return otherConn.Raw(func(secondary any) error {
			source, destination := primary.(*sqlite3.SQLiteConn), secondary.(*sqlite3.SQLiteConn)
			if restore {
				source, destination = destination, source
			}
			backup, err := destination.Backup("main", source, "main")
			if err != nil {
				return err
			}
			return runBackup(ctx, func(pages int) (bool, error) {
				remaining := backup.Remaining()
				done, err := backup.Step(pages)
				// This driver reports BUSY and LOCKED as an unfinished step.
				// Fail if no pages were copied so a blocked restore can roll back.
				if !done && err == nil && backup.Remaining() == remaining {
					return false, sqlite3.ErrBusy
				}
				return done, err
			}, backup.Finish)
		})
	})
}
