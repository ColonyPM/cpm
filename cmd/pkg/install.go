package pkgcmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPkgInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install <pkg>",
		Short: "Install a package",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("install called")
		},
	}
}
