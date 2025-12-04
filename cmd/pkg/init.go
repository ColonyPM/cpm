package pkgcmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPkgInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [dir]",
		Short: "Initialize a package directory",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("pkg init called")
		},
	}
}
