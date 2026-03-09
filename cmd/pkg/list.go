package pkgcmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)



func listPackages(cmd *cobra.Command, args []string) error {
	pkgsDir := getPackagesDir() //using getPackagesDir = pkg.GetPackagesDir from install

	entries, err := os.ReadDir(pkgsDir)
	if err != nil {
		return fmt.Errorf("reading packages directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pkgName := entry.Name()
		pkgPath := filepath.Join(pkgsDir, pkgName)

		versionEntry, err := os.ReadDir(pkgPath)
		if err != nil {
			return fmt.Errorf("reading versions for %s: %v", pkgName, err)
		}

		var versions []string
		for _, version := range versionEntry {
			if version.IsDir() {
				versions = append(versions, version.Name())
			}
		}

		cmd.Println(pkgName)
		for i, version := range versions {
			branch := "├─"
			if i == len(versions)-1 {
				branch = "└─"
			}
			cmd.Printf("  %s %s\n", branch, version)
		}

	}

	return nil
}

func newPkgListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list installed packages",
		RunE:    listPackages,
	}
}
