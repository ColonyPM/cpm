package pkgcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/ColonyPM/cpm/internal/pkg"
	"github.com/spf13/cobra"
)

var all bool
var GetPackages = pkg.GetPackagesDir

func RemovePkg(cmd *cobra.Command, args []string) error {
	pkgDir := GetPackages()
	pkgname, version, versionexist := strings.Cut(args[0], "@")
	//Move to directory
	pkgpath := filepath.Join(pkgDir, pkgname)
	
	//Check if package exists.
	_, err := os.Stat(pkgpath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("No package found.: %w", err)
		} else {
			return fmt.Errorf("package error:%w",err)
		}
	}

	if versionexist == true {
		pkgversionpath := filepath.Join(pkgpath,version)
		err = os.RemoveAll(pkgversionpath)
		if err != nil {
			return fmt.Errorf("The version",version,"was not found.", err)
		}
		println(pkgversionpath)
	}	else {
		all, _ := cmd.Flags().GetBool("all")
		if all == true {
			err = os.RemoveAll(pkgpath)
			if err != nil {
				return fmt.Errorf("No package found.: %w", err)
			}
			//println("package", args[0], "removed")
		} else {
			return fmt.Errorf("To delete the entire package, use the flag -all in the command")
		}
	}
	println("package", args[0], "removed")

	return (nil)
}

func newPkgRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove <pkg>",
		Aliases: []string{"rm"},
		Short:   "Removes packages in the default directory",
		Long:`
Removes packages in the default directory as defined in the config. 
Can delete individual versions through the 'pkgname@version' syntax 
as well as the entire package with the --all flag: 'pkgname --all'.`,
		Args:    cobra.ExactArgs(1),
		RunE:    RemovePkg,
	}
	cmd.Flags().BoolVar(&all, "all", false, "Remove ALL versions")
	return cmd
}