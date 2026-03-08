package deploycmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/ColonyPM/cpm/internal/db"
	"github.com/ColonyPM/cpm/internal/storectx"
	"github.com/briandowns/spinner"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/spf13/cobra"
)

func remove(cmd *cobra.Command, args []string) error {
	ctx := cmd.Root().Context()
	_, queries := storectx.GetDb(ctx)
	cpmConfig := storectx.GetConfig(ctx)
	cc := storectx.GetColoniesClient(ctx)

	pkgName, version, ok := strings.Cut(args[0], "@")
	if !ok || pkgName == "" || version == "" {
		return fmt.Errorf("package must be in the format name@version")
	}

	revision, err := queries.GetRevisionByPackageAndVersion(
		ctx,
		db.GetRevisionByPackageAndVersionParams{
			PackageName: pkgName,
			Version:     version,
		},
	)
	if err != nil {
		return fmt.Errorf("no revision of %s exists", args[0])
	}

	executors, err := queries.ListExecutorsByRevision(ctx, revision.ID)
	if err != nil {
		return err
	}

	for _, exec := range executors {
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Prefix = " → "
		s.Suffix = fmt.Sprintf(" removing %s", exec.ContainerID)
		s.Start()

		proc, err := cc.Submit(&core.FunctionSpec{
			FuncName:    "removeExecutor",
			Args:        []any{exec.ContainerID},
			MaxExecTime: MAX_EXEC_TIME,
			Conditions: core.Conditions{
				ColonyName:    cpmConfig.Colony.Name,
				ExecutorNames: []string{exec.AnchorName},
				ExecutorType:  "cpm-anchor",
			},
		}, cpmConfig.User.Prvkey)
		if err != nil {
			s.Stop()
			fmt.Printf(" → %s failed to remove %s: %v\n", redX, exec.ContainerID, err)
			continue
		}

		pss, err := cc.SubscribeProcess(
			cpmConfig.Colony.Name,
			proc.ID,
			"cpm-anchor",
			core.SUCCESS,
			MAX_EXEC_TIME,
			cpmConfig.User.Prvkey,
		)
		if err != nil {
			s.Stop()
			fmt.Printf(" → %s failed to subscribe remove success for %s: %v\n", redX, exec.ContainerID, err)
			continue
		}

		psf, err := cc.SubscribeProcess(
			cpmConfig.Colony.Name,
			proc.ID,
			"cpm-anchor",
			core.FAILED,
			MAX_EXEC_TIME,
			cpmConfig.User.Prvkey,
		)
		if err != nil {
			_ = pss.Close()
			s.Stop()
			fmt.Printf(" → %s failed to subscribe remove failure for %s: %v\n", redX, exec.ContainerID, err)
			continue
		}

		select {
		case <-pss.ProcessChan:
			s.Stop()
			fmt.Printf(" → %s removed %s\n", greenOK, exec.ContainerID)

		case err := <-pss.ErrChan:
			s.Stop()
			fmt.Printf(" → %s remove success subscription error for %s: %v\n", redX, exec.ContainerID, err)

		case failedProc := <-psf.ProcessChan:
			s.Stop()
			fmt.Printf(" → %s remove failed for %s: %v\n", redX, exec.ContainerID, failedProc.Errors)

		case err := <-psf.ErrChan:
			s.Stop()
			fmt.Printf(" → %s remove failure subscription error for %s: %v\n", redX, exec.ContainerID, err)
		}

		_ = pss.Close()
		_ = psf.Close()
	}

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
