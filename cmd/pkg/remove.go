package pkgcmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPkgRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove <pkg>",
		Aliases: []string{"rm"},
		Short:   "Remove a package",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("pkg remove called")
		},
	}
}
