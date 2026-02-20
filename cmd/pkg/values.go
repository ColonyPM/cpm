package pkgcmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ColonyPM/cpm/internal/pkg"
	"github.com/spf13/cobra"
)

var getValuesPath = pkg.GetValuesPath

func copyFile(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("stat source: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	if _, err := os.Stat(dstPath); err == nil {
		return fmt.Errorf("destination already exists: %s", dstPath)
	}

	dstFile, err := os.OpenFile(dstPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}

	defer func() {
		if cErr := dstFile.Close(); cErr != nil && err == nil {
			err = fmt.Errorf("close destination: %w", cErr)
		}
	}()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy data: %w", err)
	}

	return nil
}

func values(cmd *cobra.Command, args []string) error {
	pkgName := args[0]

	srcPath, err := getValuesPath(pkgName)
	if err != nil {
		return err
	}

	fileName, version, hasAt := strings.Cut(pkgName, "@")
	if !hasAt || version == "" {
		version = "latest"
	}

	defaultName := "values." + fileName + "@" + version + ".yaml"
	var dstPath string

	if len(args) > 1 {
		// Destination provided: copy to the destination directory.
		userPath := args[1]

		if info, err := os.Stat(userPath); err == nil && info.IsDir() {
			dstPath = filepath.Join(userPath, defaultName)
		} else {
			dstPath = userPath
		}
	} else {
		// No destination provided: copy to current directory.
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getwd: %w", err)
		}
		dstPath = filepath.Join(cwd, defaultName)
	}

	if err := copyFile(srcPath, dstPath); err != nil {
		return fmt.Errorf("failed to copy values: %w", err)
	}

	fmt.Printf("copied %s to %s\n", filepath.Base(srcPath), dstPath)
	return nil
}

func newValuesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "values <PKG> [DEST]",
		Short: "Copies <PKG> values.yaml file to [DEST] (or cwd)",
		Args:  cobra.RangeArgs(1, 2),
		RunE:  values,
	}
}
