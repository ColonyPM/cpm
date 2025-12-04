package pkgcmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPkgSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <pkg>",
		Short: "Search for a package",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("search called")
		},
	}
}
