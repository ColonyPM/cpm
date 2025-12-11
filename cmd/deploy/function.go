package deploycmd

import (
	"fmt"
	"time"

	"github.com/ColonyPM/cpm/internal/pkg"
	"github.com/ColonyPM/cpm/internal/storectx"
	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/spf13/cobra"
)

var Follow bool

// TODO: Attribute colonies cli
func follow(client *client.ColoniesClient, env map[string]string, process *core.Process) error {
	var lastTimestamp int64
	lastTimestamp = 0
	Count := 1 // What even is this argument?
	for {
		logs, err := client.GetLogsByProcessSince(env["COLONIES_COLONY_NAME"], process.ID, Count, lastTimestamp, env["COLONIES_PRVKEY"])
		if err != nil {
			return err
		}

		process, err := client.GetProcess(process.ID, env["COLONIES_PRVKEY"])
		if err != nil {
			return err
		}

		if len(logs) == 0 {
			time.Sleep(500 * time.Millisecond)
			if process.State == core.SUCCESS {
				fmt.Println("Process finished successfully")
				return nil
			}
			if process.State == core.FAILED {
				return fmt.Errorf("Process failed")
			}
			continue
		} else {
			for _, log := range logs {
				fmt.Print(log.Message)
			}
			lastTimestamp = logs[len(logs)-1].Timestamp
		}
	}
}

func deployFunction(cmd *cobra.Command, args []string) error {
	ctx := cmd.Root().Context()
	cc := storectx.GetColoniesClient(ctx)
	cfg := storectx.GetConfig(ctx)

	fnSpec, err := pkg.GetFunctionSpec(args[0], args[1])
	if err != nil {
		return err
	}

	proc, err := cc.Submit(fnSpec, cfg.Colonies.Prvkey)
	if err != nil {
		return err
	}

	if Follow {
		env := map[string]string{
			"COLONIES_COLONY_NAME": cfg.Colonies.ColonyName,
			"COLONIES_PRVKEY":      cfg.Colonies.Prvkey,
		}
		follow(cc, env, proc)
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

	cmd.Flags().BoolVar(&Follow, "follow", false, "follow process logs")

	return cmd
}
