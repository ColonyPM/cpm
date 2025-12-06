package pkgcmd

import (
	"github.com/spf13/cobra"
)

func NewPkgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pkg",
		Short: "Package related commands",
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
