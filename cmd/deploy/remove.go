package deploycmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDeployRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove <pkg@version>",
		Aliases: []string{"rm"},
		Short:   "Remove a deployed package version",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("remove called")
		},
	}
}
