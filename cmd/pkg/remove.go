package pkgcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/spf13/cobra"
)

var all bool

var pkgspath string

func pkgexist(pkgspath string) error {
	_, err := os.Stat(pkgspath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("No package found.: %w", err)
		} else {
			return fmt.Errorf("package error:%w",err)
		}
	}
	return nil
}

func RemovePkg(cmd *cobra.Command, args []string) error {
	pkgDir := getPackagesDir() //using getPackagesDir = pkg.GetPackagesDir from install
	pkgname, version, versionexist := strings.Cut(args[0], "@")
	pkgpath := filepath.Join(pkgDir, pkgname)

	if versionexist == true {
		pkgversionpath := filepath.Join(pkgpath, version)
		err := pkgexist(pkgversionpath)
		if err != nil {
			return err
		}
		err = os.RemoveAll(pkgversionpath)
		if err != nil {
			return fmt.Errorf("error")
		}
	} else {
		err := pkgexist(pkgpath)
		if err != nil {
			return err
		}
		all, _ := cmd.Flags().GetBool("all")
		if all == true {
			err = os.RemoveAll(pkgpath)
			if err != nil {
				return fmt.Errorf("Remove command failed: %w", err)
			}
		} else {
			return fmt.Errorf("To delete the entire package, use the flag --all in the command")
		}
	}
	println("package", args[0], "removed")
	err := os.Remove(pkgpath)
	if err == nil {
		println("Package", pkgname, "removed")
	}
	return (nil)
}

func newPkgRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove <pkg>",
		Aliases: []string{"rm"},
		Short:   "Removes packages in the default directory",
		Long: `
Removes packages in the default directory as defined in the config.
Can delete individual versions through the 'pkgname@version' syntax
as well as the entire package with the --all flag: 'pkgname --all'.`,
		Args: cobra.ExactArgs(1),
		RunE: RemovePkg,
	}
	cmd.Flags().BoolVar(&all, "all", false, "Remove ALL versions")
	return cmd
}
