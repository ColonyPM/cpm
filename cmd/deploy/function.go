package deploycmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDeployFunctionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "function <pkg@version> <fn-spec>",
		Aliases: []string{"func", "fn"},
		Short:   "Invoke or manage a function in a deployed package",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("function called")
		},
	}
}
