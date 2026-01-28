package pkgcmd

import (
	"fmt"

	"os"
	"path/filepath"
	"strings"

	"github.com/ColonyPM/cpm/internal/pkg"
	"github.com/spf13/cobra"
	//"errors"
)

var All bool
var pkgPath string

func removepkg(cmd *cobra.Command, args []string) error {
	pkgsDir := pkg.GetPackagesDir()
	pkgsname, version, versionexist := strings.Cut(args[0], "@")

	//Move to directory
	pkgspath := filepath.Join(pkgsDir, pkgsname)
	err := os.Chdir(pkgspath)
	if err != nil {
		return fmt.Errorf("No package found. If intending to remove the entire directory, use the '-all' flag: %w", err)
	}
	if versionexist == true {
		pkgsversionpath := filepath.Join(pkgspath, version)
		err := os.RemoveAll(pkgsversionpath)
		if err != nil {
			return fmt.Errorf("Version", version,"not found: %w")
		}
	}
	All, _ := cmd.Flags().GetBool("All")
	if All == true {
		err = os.RemoveAll(pkgspath)
	}

	println("package", args[0], "removed")

	return (nil)
}

func newPkgRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove <pkg>",
		Aliases: []string{"rm"},
		Short:   "Remove a package",
		Args:    cobra.ExactArgs(1),
		RunE:    removepkg,
	}
	// Local flag: only applies to `serve`.
	cmd.Flags().BoolVar(&All, "All", false, "remove ALL versions")
	//cmd.Flags().StringVar(&Dir, "Dir", "", "package directory")

	return cmd
}
