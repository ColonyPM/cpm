package deploycmd

import (
	"fmt"
	"time"

	"github.com/ColonyPM/cpm/internal/db"
	"github.com/ColonyPM/cpm/internal/pkg"
	"github.com/ColonyPM/cpm/internal/storectx"
	"github.com/briandowns/spinner"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/spf13/cobra"
)

const (
	green = "\x1b[32m"
	red   = "\x1b[31m"
	reset = "\x1b[0m"

	greenOK = green + "✔" + reset
	redX    = red + "❌" + reset
)

const MAX_EXEC_TIME = 30

func deploy(cmd *cobra.Command, args []string) error {
	ctx := cmd.Root().Context()

	// 1. Get Package
	manifest, err := pkg.GetPackageManifest(args[0])
	if err != nil {
		return err
	}

	cpmConfig := storectx.GetConfig(ctx)
	cc := storectx.GetColoniesClient(ctx)

	// 1. Get all executors
	allExecutors, err := cc.GetExecutors(cpmConfig.Colony.Name, cpmConfig.User.Prvkey)
	if err != nil {
		return err
	}

	// 2. Filter out non-anchors
	var anchors []*core.Executor
	for _, e := range allExecutors {
		if e.Type == "cpm-anchor" && e.IsApproved() {
			anchors = append(anchors, e)
		}
	}

	// 3. For each anchor, spawn the package's executors
	for _, anchor := range anchors {
		fmt.Printf("[+] Spawning executors on anchor '%s'\n", anchor.Name)

		for _, pkgExecutor := range manifest.Deployments.Executors {
			s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
			s.Prefix = " → "
			s.Suffix = fmt.Sprintf(" %s", pkgExecutor.Image)
			s.Start()

			proc, err := cc.Submit(&core.FunctionSpec{
				FuncName:    "createExecutor",
				Args:        []any{pkgExecutor.Image},
				MaxExecTime: MAX_EXEC_TIME,
				Conditions: core.Conditions{
					ColonyName:    cpmConfig.Colony.Name,
					ExecutorNames: []string{anchor.Name},
					ExecutorType:  "cpm-anchor",
				},
			}, cpmConfig.User.Prvkey)

			if err != nil {
				return err
			}

			pss, err := cc.SubscribeProcess(cpmConfig.Colony.Name, proc.ID, "cpm-anchor", core.SUCCESS, MAX_EXEC_TIME, cpmConfig.User.Prvkey)
			if err != nil {
				return err
			}
			psf, err := cc.SubscribeProcess(cpmConfig.Colony.Name, proc.ID, "cpm-anchor", core.FAILED, MAX_EXEC_TIME, cpmConfig.User.Prvkey)
			if err != nil {
				return err
			}

			select {
			case proc := <-pss.ProcessChan:
				proc, err := cc.GetProcess(proc.ID, cpmConfig.User.Prvkey)
				if err != nil {
					return err
				}
				s.Stop()
				fmt.Printf(" → %s %s\n", greenOK, pkgExecutor.Image)
			case err := <-pss.ErrChan:
				s.Stop()
				fmt.Printf("PSS Subscription error: %v\n", err)
			case prooc := <-psf.ProcessChan:
				s.Stop()
				fmt.Printf("PSF Received process update: %v\n", prooc.Errors)
			case err := <-psf.ErrChan:
				s.Stop()
				fmt.Printf("PSF Subscription error: %v\n", err)
			}
		}
	}

	_, q := storectx.GetDb(ctx)

	_, err = q.CreateDeployment(ctx, db.CreateDeploymentParams{
		PkgName:    "LMAO",
		DeployedAt: time.Now().UTC(),
	})

	if err != nil {
		return err
	}

	return nil
}

func NewDeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy <pkg>",
		Short: "Deploy a package to the colony",
		Args:  cobra.ExactArgs(1),
		RunE:  deploy,
	}

	cmd.AddCommand(newDeployGetCmd())
	cmd.AddCommand(newDeployListCmd())
	cmd.AddCommand(newDeployRemoveCmd())
	cmd.AddCommand(newDeployFunctionCmd())
	cmd.AddCommand(newDeployWorkflowCmd())

	return cmd
}
