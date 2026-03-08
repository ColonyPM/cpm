package deploycmd

import (
	"github.com/ColonyPM/cpm/internal/pkg"
	"github.com/ColonyPM/cpm/internal/storectx"
	"github.com/spf13/cobra"
)

func deployFunction(cmd *cobra.Command, args []string) error {
	ctx := cmd.Root().Context()
	cc := storectx.GetColoniesClient(ctx)
	cfg := storectx.GetConfig(ctx)

	fnSpec, err := pkg.GetFunctionSpec(args[0], args[1])
	if err != nil {
		return err
	}

	if _, err := cc.Submit(fnSpec, cfg.User.Prvkey); err != nil {
		return err
	}
	return nil
}

func newDeployFunctionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "function <pkg> <fn-spec>",
		Aliases: []string{"func", "fn"},
		Short:   "Submit a function spec in a deployed package",
		Args:    cobra.ExactArgs(2),
		RunE:    deployFunction,
	}

	return cmd
}
