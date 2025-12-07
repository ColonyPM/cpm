package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// dbPath returns ./cpm.db relative to current working directory.
func dbPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("working dir: %w", err)
	}
	return filepath.Join(wd, "cpm.db"), nil
}

// OpenLocal opens (and initializes, if needed) the SQLite database in the project directory.
func OpenLocal(ctx context.Context) (*sql.DB, error) {
	path, err := dbPath()
	if err != nil {
		return nil, err
	}

	needsInit := false
	if _, err := os.Stat(path); os.IsNotExist(err) {
		needsInit = true
	}

	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", path)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if needsInit {
		if _, err := conn.ExecContext(ctx, Schema); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("apply schema: %w", err)
		}
	}

	return conn, nil
}
