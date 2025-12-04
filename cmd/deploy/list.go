package deploycmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDeployListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List deployments",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("list called")
		},
	}
}
