package pkgcmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPkgUploadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upload <token>",
		Short: "Upload a package",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("upload called")
		},
	}
}
