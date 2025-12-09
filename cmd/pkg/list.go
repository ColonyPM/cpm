package pkgcmd

import (
	"fmt"
	"os"

	"github.com/ColonyPM/cpm/internal/pkg"
	"github.com/spf13/cobra"
)

func listPackages(cmd *cobra.Command, args []string) error {
	packagesPath, err := pkg.GetOrMakePackagesDirectory()
	if err != nil {
		return fmt.Errorf("getting packages directory: %v\n", err)
	}

	fmt.Println("Installed packages:")

	entries, err := os.ReadDir(packagesPath)
	if err != nil {
		return fmt.Errorf("reading packages directory: %v\n", err)
	}

	for _, entry := range entries {
		fmt.Println("    - " + entry.Name())
	}
	return nil
}

func newPkgListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List locally installedpackages",
		RunE:    listPackages,
	}
}
