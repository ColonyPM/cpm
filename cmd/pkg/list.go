package pkgcmd

import (
	"fmt"
	"os"

	"github.com/ColonyPM/cpm/internal/pkg"
	"github.com/spf13/cobra"
)

func listPackages(cmd *cobra.Command, args []string) error {
	pkgsDir := pkg.GetPackagesDir()

	fmt.Println("Installed packages:")

	entries, err := os.ReadDir(pkgsDir)
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
		Short:   "List locally installed packages",
		RunE:    listPackages,
	}
}
