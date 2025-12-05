package root

import (
	"github.com/spf13/cobra"

	deploycmd "github.com/ColonyPM/cpm/cmd/deploy"
	pkgcmd "github.com/ColonyPM/cpm/cmd/pkg"
	store "github.com/ColonyPM/cpm/internal/db"
	"github.com/ColonyPM/cpm/internal/storectx"
)

var rootCmd = &cobra.Command{
	Use:   "cpm-cli",
	Short: "A brief description of your application",
	Long:  ``,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if storectx.IsInitialized(cmd) {
			return nil
		}

		ctx := cmd.Context()

		dbConn, err := store.OpenLocal(ctx)
		if err != nil {
			return err
		}
		q := store.New(dbConn)

		storectx.AttachToRoot(cmd.Root(), dbConn, q)

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(pkgcmd.NewPkgCmd())
	rootCmd.AddCommand(deploycmd.NewDeployCmd())
}
