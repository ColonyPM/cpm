package pkgcmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ColonyPM/cpm/internal/pkg"
	"github.com/spf13/cobra"
)

func createManifest(pkgPath string) error {

	// Manifest file
	path := filepath.Join(pkgPath, "package.yaml")

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("error creating package.yaml: %w", err)
	}
	defer f.Close()

	if err := pkg.WriteManifest(f, pkg.NewDefaultManifest(filepath.Base(pkgPath))); err != nil {
		return fmt.Errorf("writing package manifest: %w", err)
	}

	return nil
}

func initPackage(cmd *cobra.Command, args []string) error {

	// Create directories
	pkgPath := args[0]
	// If arg is empty create package in current directory
	if len(args) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get cwd: %v", err)
		}
		pkgPath = filepath.Base(cwd)
	}
	// Package directory
	err := os.MkdirAll(pkgPath, 0o755)
	if err != nil {
		return fmt.Errorf("creating package directory: %w", err)
	}
	err = os.Mkdir(filepath.Join(pkgPath, "templates"), 0o755)
	if err != nil {
		return fmt.Errorf("creating package templates directory: %w", err)
	}

	// Create manifest file
	err = createManifest(pkgPath)
	if err != nil {
		return fmt.Errorf("creating package manifest: %w", err)
	}

	// Create values.yaml file
	valuesPath := filepath.Join(pkgPath, "values.yaml")
	f, err := os.OpenFile(valuesPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("error creating values.yaml: %w", err)
	}
	defer f.Close()

	_, err = f.WriteString("global:\n")
	if err != nil {
		return fmt.Errorf("error writing to values.yaml: %w", err)
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
