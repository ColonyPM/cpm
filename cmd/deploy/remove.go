package deploycmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func remove(cmd *cobra.Command, args []string) error {
	// 1. Stop executors
	// 2. Run teardown scripts*
	// 3. Unregister deployment (db)
	fmt.Println("remove called")

	return nil
}

func newDeployRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove <pkg@version>",
		Aliases: []string{"rm"},
		Short:   "Remove a deployed package version",
		Args:    cobra.ExactArgs(1),
		RunE:    remove,
	}
}
