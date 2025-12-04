package pkgcmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPkgListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List packages",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("pkg list called")
		},
	}
}
