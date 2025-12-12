package pkgcmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ColonyPM/cpm/internal/pkg"
	"github.com/spf13/cobra"
)

func CopyFile(srcPath, dstPath string) error {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("readfile: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	info, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}

	if err := os.WriteFile(dstPath, data, info.Mode()); err != nil {
		return fmt.Errorf("writefile: %w", err)
	}

	return nil
}

func values(cmd *cobra.Command, args []string) error {
	pkgDir, err := pkg.GetPackageDirectory(args[0])
	if err != nil {
		return err
	}

	valuesPath := filepath.Join(pkgDir, "values.yaml")

	filename := "values." + args[0] + ".yaml"

	var dstPath string
	if len(args) > 1 {
		dstPath = filepath.Join(args[1], filename)
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getwd: %w", err)
		}

		dstPath = filepath.Join(cwd, filename)
	}

	if err := CopyFile(valuesPath, dstPath); err != nil {
		return fmt.Errorf("copyfile: %w", err)
	}

	fmt.Printf("values.yaml copied to %s\n", dstPath)

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
