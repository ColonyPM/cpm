package deploycmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDeployWorkflowCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "workflow <pkg@version> <workflow>",
		Aliases: []string{"wf"},
		Short:   "Manage a workflow for a deployed package",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("workflow called")
		},
	}
}
