package pkgcmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewPkgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pkg",
		Short: "Package related commands",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("pkg called")
		},
	}

	cmd.AddCommand(newPkgInitCmd())
	cmd.AddCommand(newPkgInstallCmd())
	cmd.AddCommand(newPkgRemoveCmd())
	cmd.AddCommand(newPkgListCmd())
	cmd.AddCommand(newPkgGetCmd())
	cmd.AddCommand(newPkgSearchCmd())
	cmd.AddCommand(newPkgUploadCmd())

	return cmd
}
