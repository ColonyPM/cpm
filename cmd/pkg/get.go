package pkgcmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPkgGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <pkg>",
		Short: "Get package details",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("pkg get called")
		},
	}
}
