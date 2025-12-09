package pkgcmd

import (
	"fmt"
	"os"

	"github.com/ColonyPM/cpm/internal/pkg"
	"github.com/spf13/cobra"
)

func listPackages(cmd *cobra.Command, args []string) {
	packagesPath, err := pkg.GetOrMakePackagesDirectory()
	if err != nil {
		fmt.Printf("Error getting packages directory: %v\n", err)
		return
	}

	fmt.Println("Installed packages:")

	entries, err := os.ReadDir(packagesPath)
	if err != nil {
		fmt.Printf("Error reading packages directory: %v\n", err)
		return
	}

	for _, entry := range entries {
		fmt.Println("    - " + entry.Name())
	}
}

func newPkgListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List locally installedpackages",
		Run:     listPackages,
	}
}
