package cmd

import (
	"fmt"

	deploycmd "github.com/ColonyPM/cpm/cmd/deploy"
	pkgcmd "github.com/ColonyPM/cpm/cmd/pkg"
	"github.com/ColonyPM/cpm/internal/config"
	store "github.com/ColonyPM/cpm/internal/db"
	"github.com/ColonyPM/cpm/internal/pkg"
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

		_, err := pkg.EnsurePackagesDir()
		if err != nil {
			return fmt.Errorf("initialization failed: %w", err)
		}

		// Load config
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Use this ctx for db init as well
		dbConn, err := store.OpenLocal(ctx)
		if err != nil {
			return err
		}
		q := store.New(dbConn)

		// build Colonies client from config
		// Note: Colonies client uses "insecure" flag (true = HTTP, false = HTTPS)
		// Our config uses "tls" flag (true = HTTPS, false = HTTP), so we invert it
		cc := client.CreateColoniesClient(cfg.Colonies.Host, cfg.Colonies.Port, !cfg.Colonies.TLS, false)

		// attach to ctx and set it back on the root
		ctx = storectx.WithStore(ctx, dbConn, q, cc, cfg)
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
