package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ColonyPM/cpm/internal/config"
	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	cfs "github.com/colonyos/colonies/pkg/fs"
	_ "modernc.org/sqlite"
)

const (
	LABEL   = "/db"
	DB_FILE = "cpm.db"
)

func getLocalDBPath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	appCache := filepath.Join(cacheDir, "cpm", "db")
	err = os.MkdirAll(appCache, 0755)

	return appCache, err
}

func createFSClientWithEnv(cfg *config.Config, cc *client.ColoniesClient) (*cfs.FSClient, error) {
	envMapping := map[string]string{
		"AWS_S3_ENDPOINT":   cfg.Storage.Endpoint,
		"AWS_S3_ACCESSKEY":  cfg.Storage.Accesskey,
		"AWS_S3_SECRETKEY":  cfg.Storage.Secretkey,
		"AWS_S3_REGION":     cfg.Storage.Region,
		"AWS_S3_BUCKET":     cfg.Storage.Bucket,
		"AWS_S3_TLS":        strconv.FormatBool(cfg.Storage.TLS),
		"AWS_S3_SKIPVERIFY": strconv.FormatBool(cfg.Storage.Skipverify),
	}

	originalEnv := make(map[string]string)
	for key := range envMapping {
		if val, exists := os.LookupEnv(key); exists {
			originalEnv[key] = val
		}
	}

	for key, value := range envMapping {
		os.Setenv(key, value)
	}

	defer func() {
		for key := range envMapping {
			if val, exists := originalEnv[key]; exists {
				os.Setenv(key, val)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	return cfs.CreateFSClient(cc, cfg.Colony.Name, cfg.User.Prvkey)
}

func OpenDB(ctx context.Context, cc *client.ColoniesClient, cfg *config.Config) (*sql.DB, error) {
	localDbPath, err := getLocalDBPath()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(localDbPath, 0o755); err != nil {
		return nil, fmt.Errorf("create local db dir: %w", err)
	}

	path := filepath.Join(localDbPath, "cpm.db")

	// Try to get the remote DB file.
	files, err := cc.GetLatestFileByName(cfg.Colony.Name, LABEL, DB_FILE, cfg.User.Prvkey)
	if err != nil {
		var coloniesErr *core.ColoniesError
		if !(errors.As(err, &coloniesErr) &&
			coloniesErr.Status == 404 &&
			coloniesErr.Message == "Failed to get file") {
			return nil, err
		}
		// Exact "missing remote file" case:
		// do nothing here, we'll create/init locally below.
	} else {
		// Remote DB exists, download it to localDbPath.
		fsClient, err := createFSClientWithEnv(cfg, cc)
		if err != nil {
			return nil, err
		}
		fsClient.Quiet = true

		if len(files) == 0 {
			return nil, fmt.Errorf("remote file lookup succeeded but returned no files")
		}

		if err := fsClient.Download(cfg.Colony.Name, files[0].ID, localDbPath); err != nil {
			return nil, fmt.Errorf("download db: %w", err)
		}
	}

	needsInit := false
	if _, err := os.Stat(path); os.IsNotExist(err) {
		needsInit = true
	} else if err != nil {
		return nil, fmt.Errorf("stat db path: %w", err)
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

func CloseDB(ctx context.Context, cc *client.ColoniesClient, cfg *config.Config) error {
	localDbPath, err := getLocalDBPath()
	if err != nil {
		return err
	}

	fsClient, err := createFSClientWithEnv(cfg, cc)
	if err != nil {
		return err
	}
	fsClient.Quiet = true

	sp, err := fsClient.CalcSyncPlan(localDbPath, LABEL, true)
	if err != nil {
		return err
	}

	if err := fsClient.ApplySyncPlan(sp); err != nil {
		return err
	}

	return nil
}
