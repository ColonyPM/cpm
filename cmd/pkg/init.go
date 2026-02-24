package pkgcmd

import (
	"fmt"
	"os"
	"path/filepath"

	"io"

	"github.com/ColonyPM/cpm/internal/pkg"
	"github.com/spf13/cobra"
)

func createFile(pkgPath string, fileName string) error {
	path := filepath.Join(pkgPath, fileName)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("error creating %s: %w", fileName, err)
	}
	defer f.Close()

	pkgName := filepath.Base(pkgPath)

	var write func() error

	switch fileName {
	case "package.yaml":
		write = func() error {
			return pkg.WriteManifest(f, pkg.NewDefaultManifest(pkgName))
		}

	case "readme.md":
		write = func() error {
			if _, err := fmt.Fprintf(f, "# %s\n", pkgName); err != nil {
				return fmt.Errorf("error writing readme: %w", err)
			}
			return nil
		}

	case "values.yaml":
		write = func() error {
			if _, err := io.WriteString(f, "global:\n"); err != nil {
				return fmt.Errorf("error writing to values.yaml: %w", err)
			}
			return nil
		}

	default:
		return fmt.Errorf("unknown file: %s", fileName)
	}

	if err := write(); err != nil {
		return fmt.Errorf("writing %s: %w", fileName, err)
	}
	return nil
}

func initPackage(cmd *cobra.Command, args []string) error {

	// Create directories
	var pkgPath string
	// If arg is empty create package in current directory
	if len(args) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get cwd: %v", err)
		}
		pkgPath = cwd
	} else {
		pkgPath = args[0]
		// Package directory
		err := os.MkdirAll(pkgPath, 0o755)
		if err != nil {
			return fmt.Errorf("creating package directory: %w", err)
		}
	}

	err := os.Mkdir(filepath.Join(pkgPath, "templates"), 0o755)
	if err != nil {
		return fmt.Errorf("creating package templates directory: %w", err)
	}

	// Create manifest.yaml, README.md, values.yaml
	files := []string{"package.yaml", "readme.md", "values.yaml"}

	for _, name := range files {
		if err := createFile(pkgPath, name); err != nil {
			return err
		}
	}

	return nil

}

func newPkgInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a package directory",
		Args:  cobra.MaximumNArgs(1),
		RunE:  initPackage,
	}
}
