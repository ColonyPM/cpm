package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	_ "modernc.org/sqlite"
)

// dataDir returns the XDG data directory for cpm.
func dataDir() (string, error) {
	var dir string

	switch runtime.GOOS {
	case "windows":
		base := os.Getenv("LOCALAPPDATA")
		if base == "" {
			base = os.Getenv("APPDATA")
		}
		if base == "" {
			return "", fmt.Errorf("LOCALAPPDATA/APPDATA not set")
		}
		dir = filepath.Join(base, "cpm")

	case "darwin":
		base := os.Getenv("XDG_DATA_HOME")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("cannot determine home directory: %w", err)
			}
			base = filepath.Join(home, "Library", "Application Support")
		}
		dir = filepath.Join(base, "cpm")

	default: // linux, freebsd, etc.
		dataHome := os.Getenv("XDG_DATA_HOME")
		if dataHome == "" {
			home := os.Getenv("HOME")
			if home == "" {
				return "", fmt.Errorf("XDG_DATA_HOME and HOME not set")
			}
			dataHome = filepath.Join(home, ".local", "share")
		}
		dir = filepath.Join(dataHome, "cpm")
	}

	return dir, nil
}

// dbPath returns the path to cpm.db in the XDG data directory.
func dbPath() (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "cpm.db"), nil
}

// OpenLocal opens (and initializes, if needed) the SQLite database in the XDG data directory.
func OpenLocal(ctx context.Context) (*sql.DB, error) {
	path, err := dbPath()
	if err != nil {
		return nil, err
	}

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("creating data directory: %w", err)
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
