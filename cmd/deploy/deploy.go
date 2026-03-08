package deploycmd

import (
	"fmt"
	"strings"
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

type SpawnedExecutor struct {
	ExecutorName string
	AnchorName   string
	ContainerID  string
	ImgName      string
}

func deploy(cmd *cobra.Command, args []string) error {
	ctx := cmd.Root().Context()

	manifest, err := pkg.GetPackageManifest(args[0])
	if err != nil {
		return err
	}

	manifest.Deployments.Executors[1].Image = "RIP"

	cpmConfig := storectx.GetConfig(ctx)
	cc := storectx.GetColoniesClient(ctx)

	allExecutors, err := cc.GetExecutors(cpmConfig.Colony.Name, cpmConfig.User.Prvkey)
	if err != nil {
		return err
	}

	var anchors []*core.Executor
	for _, e := range allExecutors {
		if e.Type == "cpm-anchor" && e.IsApproved() {
			anchors = append(anchors, e)
		}
	}

	var spawnedExecutors []SpawnedExecutor

	cleanupSpawned := func() error {
		if len(spawnedExecutors) == 0 {
			return nil
		}

		fmt.Println("[-] Deployment failed, removing already spawned containers...")
		var cleanupErrs []string

		for i := len(spawnedExecutors) - 1; i >= 0; i-- {
			spawned := spawnedExecutors[i]

			s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
			s.Prefix = " → "
			s.Suffix = fmt.Sprintf(" removing %s", spawned.ContainerID)
			s.Start()

			proc, err := cc.Submit(&core.FunctionSpec{
				FuncName:    "removeExecutor",
				Args:        []any{spawned.ContainerID},
				MaxExecTime: MAX_EXEC_TIME,
				Conditions: core.Conditions{
					ColonyName:    cpmConfig.Colony.Name,
					ExecutorNames: []string{spawned.AnchorName},
					ExecutorType:  "cpm-anchor",
				},
			}, cpmConfig.User.Prvkey)
			if err != nil {
				s.Stop()
				fmt.Printf(" → %s failed to remove %s: %v\n", redX, spawned.ContainerID, err)
				cleanupErrs = append(cleanupErrs, fmt.Sprintf("remove %s: submit failed: %v", spawned.ContainerID, err))
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
				fmt.Printf(" → %s failed to subscribe remove success for %s: %v\n", redX, spawned.ContainerID, err)
				cleanupErrs = append(cleanupErrs, fmt.Sprintf("remove %s: success subscribe failed: %v", spawned.ContainerID, err))
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
				fmt.Printf(" → %s failed to subscribe remove failure for %s: %v\n", redX, spawned.ContainerID, err)
				cleanupErrs = append(cleanupErrs, fmt.Sprintf("remove %s: failure subscribe failed: %v", spawned.ContainerID, err))
				continue
			}

			select {
			case <-pss.ProcessChan:
				s.Stop()
				fmt.Printf(" → %s removed %s\n", greenOK, spawned.ContainerID)

			case err := <-pss.ErrChan:
				s.Stop()
				fmt.Printf(" → %s remove success subscription error for %s: %v\n", redX, spawned.ContainerID, err)
				cleanupErrs = append(cleanupErrs, fmt.Sprintf("remove %s: success subscription error: %v", spawned.ContainerID, err))

			case failedProc := <-psf.ProcessChan:
				s.Stop()
				fmt.Printf(" → %s remove failed for %s: %v\n", redX, spawned.ContainerID, failedProc.Errors)
				cleanupErrs = append(cleanupErrs, fmt.Sprintf("remove %s: process failed: %v", spawned.ContainerID, failedProc.Errors))

			case err := <-psf.ErrChan:
				s.Stop()
				fmt.Printf(" → %s remove failure subscription error for %s: %v\n", redX, spawned.ContainerID, err)
				cleanupErrs = append(cleanupErrs, fmt.Sprintf("remove %s: failure subscription error: %v", spawned.ContainerID, err))
			}

			_ = pss.Close()
			_ = psf.Close()
		}

		if len(cleanupErrs) > 0 {
			return fmt.Errorf(strings.Join(cleanupErrs, "; "))
		}

		return nil
	}

	failAndCleanup := func(cause error) error {
		cleanupErr := cleanupSpawned()
		if cleanupErr != nil {
			return fmt.Errorf("%w (rollback failed: %v)", cause, cleanupErr)
		}
		return cause
	}

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
				s.Stop()
				return failAndCleanup(fmt.Errorf("failed to submit createExecutor for %s on %s: %w", pkgExecutor.Image, anchor.Name, err))
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
				return failAndCleanup(fmt.Errorf("failed to subscribe success for %s on %s: %w", pkgExecutor.Image, anchor.Name, err))
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
				return failAndCleanup(fmt.Errorf("failed to subscribe failure for %s on %s: %w", pkgExecutor.Image, anchor.Name, err))
			}

			select {
			case procUpdate := <-pss.ProcessChan:
				procInfo, err := cc.GetProcess(procUpdate.ID, cpmConfig.User.Prvkey)
				if err != nil {
					_ = pss.Close()
					_ = psf.Close()
					s.Stop()
					return failAndCleanup(fmt.Errorf("failed to fetch process output for %s on %s: %w", pkgExecutor.Image, anchor.Name, err))
				}

				if len(procInfo.Output) < 2 {
					_ = pss.Close()
					_ = psf.Close()
					s.Stop()
					return failAndCleanup(fmt.Errorf("createExecutor returned invalid output for %s on %s", pkgExecutor.Image, anchor.Name))
				}

				executorName, ok1 := procInfo.Output[0].(string)
				containerID, ok2 := procInfo.Output[1].(string)
				if !ok1 || !ok2 {
					_ = pss.Close()
					_ = psf.Close()
					s.Stop()
					return failAndCleanup(fmt.Errorf("createExecutor returned unexpected output types for %s on %s", pkgExecutor.Image, anchor.Name))
				}

				s.Stop()
				fmt.Printf(" → %s %s\n", greenOK, pkgExecutor.Image)

				spawnedExecutors = append(spawnedExecutors, SpawnedExecutor{
					ExecutorName: executorName,
					AnchorName:   anchor.Name,
					ContainerID:  containerID,
					ImgName:      pkgExecutor.Image,
				})

			case err := <-pss.ErrChan:
				_ = pss.Close()
				_ = psf.Close()
				s.Stop()
				return failAndCleanup(fmt.Errorf("success subscription error for %s on %s: %w", pkgExecutor.Image, anchor.Name, err))

			case failedProc := <-psf.ProcessChan:
				_ = pss.Close()
				_ = psf.Close()
				s.Stop()
				return failAndCleanup(fmt.Errorf("createExecutor failed for %s on %s: %v", pkgExecutor.Image, anchor.Name, failedProc.Errors[0]))

			case err := <-psf.ErrChan:
				_ = pss.Close()
				_ = psf.Close()
				s.Stop()
				return failAndCleanup(fmt.Errorf("failure subscription error for %s on %s: %w", pkgExecutor.Image, anchor.Name, err))
			}

			_ = pss.Close()
			_ = psf.Close()
		}
	}

	database, queries := storectx.GetDb(ctx)

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := queries.WithTx(tx)

	pkgName, version, _ := strings.Cut(args[0], "@")

	revision, err := qtx.CreateRevision(ctx, db.CreateRevisionParams{
		PackageName: pkgName,
		Version:     version,
		DeployTime:  time.Now(),
	})
	if err != nil {
		return err
	}

	for _, exec := range spawnedExecutors {
		_, err := qtx.CreateExecutor(ctx, db.CreateExecutorParams{
			RevisionID:   revision.ID,
			ExecutorName: exec.ExecutorName,
			AnchorName:   exec.AnchorName,
			ContainerID:  exec.ContainerID,
			ImgName:      exec.ImgName,
		})
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
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
