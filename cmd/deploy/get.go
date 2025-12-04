package deploycmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDeployGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <pkg@version>",
		Short: "Get deployment details for a package version",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("get called")
		},
	}
}
