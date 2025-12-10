package pkgcmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ColonyPM/cpm/internal/pkg"
	"path/filepath"
	"os"
	"strings"
	//"errors"
)
 
var All bool
func removepkg(cmd *cobra.Command, args []string) error {
	//get packagedir
	pkgsdir,err := pkg.GetOrMakePackagesDirectory()
	if err != nil {	
		return fmt.Errorf("error creating values.yaml: %w", err)
	}
	println(pkgsdir)
	//Split name and version

	pkgsname, version, versionexist:= strings.Cut(args[0],"@")
	//If version is absent, fail for now


	//Move to directory
		pkgspath := filepath.Join(pkgsdir,pkgsname)
		err = os.Chdir(pkgspath)
		if err != nil {
			return fmt.Errorf("No package found: %w", err)
		}
	if versionexist == true {
		pkgsversionpath := filepath.Join(pkgspath,version)
		err = os.RemoveAll(pkgsversionpath)
	}
	All, _:= cmd.Flags().GetBool("All")
	if All == true {
	err = os.RemoveAll(pkgspath)
	}
	
	println("package",args[0],"removed")
	
	return(nil)
}

func newPkgRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove <pkg>",
		Aliases: []string{"rm"},
		Short:   "Remove a package",
		Args:    cobra.ExactArgs(1),
		RunE: removepkg,
	}
	// Local flag: only applies to `serve`.
	cmd.Flags().BoolVar(&All, "All", false, "remove all versions")

	return cmd
}


