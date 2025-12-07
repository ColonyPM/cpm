package cmd

import (
	"os"
	"strconv"

	deploycmd "github.com/ColonyPM/cpm/cmd/deploy"
	pkgcmd "github.com/ColonyPM/cpm/cmd/pkg"
	store "github.com/ColonyPM/cpm/internal/db"
	"github.com/ColonyPM/cpm/internal/storectx"
	"github.com/colonyos/colonies/pkg/client"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cpm",
	Short: "A brief description of your application",
	Long:  ``,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Decide whether you want to use the root ctx or the command ctx as the base
		root := cmd.Root()
		ctx := root.Context()

		if storectx.IsInitialized(ctx) {
			return nil
		}

		// Use this ctx for db init as well
		dbConn, err := store.OpenLocal(ctx)
		if err != nil {
			return err
		}
		q := store.New(dbConn)

		// build Colonies client
		host := os.Getenv("COLONIES_SERVER_HOST")
		portStr := os.Getenv("COLONIES_SERVER_PORT")
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return err
		}

		cc := client.CreateColoniesClient(host, port, true, false)

		// attach to ctx and set it back on the root
		ctx = storectx.WithStore(ctx, dbConn, q, cc)
		root.SetContext(ctx)

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.SilenceUsage = true

	rootCmd.AddCommand(pkgcmd.NewPkgCmd())
	rootCmd.AddCommand(deploycmd.NewDeployCmd())
}
