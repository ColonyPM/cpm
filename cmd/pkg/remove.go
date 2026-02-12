package pkgcmd

import (
	"fmt"
	"os"
	"path/filepath"
	//"strings"
	"github.com/ColonyPM/cpm/internal/pkg"
	"github.com/spf13/cobra"
)

var All bool
var GetPackages = pkg.GetPackagesDir

func RemovePkg(cmd *cobra.Command, args []string) error {
	pkgsDir := GetPackages()
	//pkgsname, version, versionexist := strings.Cut(args[0], "@")	
	pkgsname := args[0]
	//Move to directory
	pkgspath := filepath.Join(pkgsDir, pkgsname)
	err := os.Chdir(pkgspath)
	if err != nil {
		return fmt.Errorf("No package found.: %w", err)
	}
	err = os.RemoveAll(pkgspath)


/* Old version implementation, might be removed later.
	if versionexist == true {
		pkgsversionpath := filepath.Join(pkgspath, version)
		err := os.RemoveAll(pkgsversionpath)
		if err != nil {
			return fmt.Errorf("Version not found: %w",err)
		}
	}
	All, _ := cmd.Flags().GetBool("All")
	if All == true {
		err = os.RemoveAll(pkgspath)
	}
*/
	println("package", args[0], "removed")

	return (nil)
}

func newPkgRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove <pkg>",
		Aliases: []string{"rm"},
		Short:   "Remove a package",
		Args:    cobra.ExactArgs(1),
		RunE:    RemovePkg,
	}
	// Local flag: only applies to `serve`.
	cmd.Flags().BoolVar(&All, "All", false, "remove ALL versions")
	
	return cmd
}