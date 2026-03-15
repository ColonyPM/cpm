package deploycmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ColonyPM/cpm/internal/db"
	"github.com/ColonyPM/cpm/internal/pkg"
	"github.com/ColonyPM/cpm/internal/storectx"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/spf13/cobra"
)

const MAX_EXEC_TIME = 30

type sqlDeployRepo struct {
	db      *sql.DB
	queries *db.Queries
}

func (r sqlDeployRepo) RevisionExists(ctx context.Context, pkgName, version string) (bool, error) {
	_, err := r.queries.GetRevisionByPackageAndVersion(ctx, db.GetRevisionByPackageAndVersionParams{
		PackageName: pkgName,
		Version:     version,
	})
	if err == nil {
		return true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return false, err
}

func (r sqlDeployRepo) SaveDeployment(ctx context.Context, pkgName, version string, deployedAt time.Time, executors []SpawnedExecutor) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)

	revision, err := qtx.CreateRevision(ctx, db.CreateRevisionParams{
		PackageName: pkgName,
		Version:     version,
		DeployTime:  deployedAt,
	})
	if err != nil {
		return err
	}

	for _, exec := range executors {
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

	return tx.Commit()
}

type runtimeFuncs struct {
	approvedAnchors func(ctx context.Context) ([]string, error)
	createExecutor  func(ctx context.Context, anchorName, image string) (SpawnedExecutor, error)
	removeExecutor  func(ctx context.Context, anchorName, containerID string) error
}

func (r runtimeFuncs) ApprovedAnchors(ctx context.Context) ([]string, error) {
	return r.approvedAnchors(ctx)
}

func (r runtimeFuncs) CreateExecutor(ctx context.Context, anchorName, image string) (SpawnedExecutor, error) {
	return r.createExecutor(ctx, anchorName, image)
}

func (r runtimeFuncs) RemoveExecutor(ctx context.Context, anchorName, containerID string) error {
	return r.removeExecutor(ctx, anchorName, containerID)
}

func deploy(cmd *cobra.Command, args []string) error {
	ctx := cmd.Root().Context()

	database, queries := storectx.GetDb(ctx)
	cfg := storectx.GetConfig(ctx)
	cc := storectx.GetColoniesClient(ctx)

	loadExecutorImages := func(ref string) ([]string, error) {
		manifest, err := pkg.GetPackageManifest(ref)
		if err != nil {
			return nil, err
		}

		images := make([]string, 0, len(manifest.Deployments.Executors))
		for _, exec := range manifest.Deployments.Executors {
			images = append(images, exec.Image)
		}
		return images, nil
	}

	runtime := runtimeFuncs{
		approvedAnchors: func(ctx context.Context) ([]string, error) {
			allExecutors, err := cc.GetExecutors(cfg.Colony.Name, cfg.User.Prvkey)
			if err != nil {
				return nil, err
			}

			anchors := make([]string, 0)
			for _, e := range allExecutors {
				if e.Type == "cpm-anchor" && e.IsApproved() {
					anchors = append(anchors, e.Name)
				}
			}

			return anchors, nil
		},

		createExecutor: func(ctx context.Context, anchorName, image string) (SpawnedExecutor, error) {
			proc, err := cc.Submit(&core.FunctionSpec{
				FuncName:    "createExecutor",
				Args:        []any{image},
				MaxExecTime: MAX_EXEC_TIME,
				Conditions: core.Conditions{
					ColonyName:    cfg.Colony.Name,
					ExecutorNames: []string{anchorName},
					ExecutorType:  "cpm-anchor",
				},
			}, cfg.User.Prvkey)
			if err != nil {
				return SpawnedExecutor{}, fmt.Errorf("failed to submit createExecutor for %s on %s: %w", image, anchorName, err)
			}

			pss, err := cc.SubscribeProcess(
				cfg.Colony.Name,
				proc.ID,
				"cpm-anchor",
				core.SUCCESS,
				MAX_EXEC_TIME,
				cfg.User.Prvkey,
			)
			if err != nil {
				return SpawnedExecutor{}, fmt.Errorf("failed to subscribe success for %s on %s: %w", image, anchorName, err)
			}
			defer pss.Close()

			psf, err := cc.SubscribeProcess(
				cfg.Colony.Name,
				proc.ID,
				"cpm-anchor",
				core.FAILED,
				MAX_EXEC_TIME,
				cfg.User.Prvkey,
			)
			if err != nil {
				return SpawnedExecutor{}, fmt.Errorf("failed to subscribe failure for %s on %s: %w", image, anchorName, err)
			}
			defer psf.Close()

			select {
			case procUpdate := <-pss.ProcessChan:
				procInfo, err := cc.GetProcess(procUpdate.ID, cfg.User.Prvkey)
				if err != nil {
					return SpawnedExecutor{}, fmt.Errorf("failed to fetch process output for %s on %s: %w", image, anchorName, err)
				}

				if len(procInfo.Output) < 2 {
					return SpawnedExecutor{}, fmt.Errorf("createExecutor returned invalid output for %s on %s", image, anchorName)
				}

				executorName, ok1 := procInfo.Output[0].(string)
				containerID, ok2 := procInfo.Output[1].(string)
				if !ok1 || !ok2 {
					return SpawnedExecutor{}, fmt.Errorf("createExecutor returned unexpected output types for %s on %s", image, anchorName)
				}

				return SpawnedExecutor{
					ExecutorName: executorName,
					AnchorName:   anchorName,
					ContainerID:  containerID,
					ImgName:      image,
				}, nil

			case err := <-pss.ErrChan:
				return SpawnedExecutor{}, fmt.Errorf("success subscription error for %s on %s: %w", image, anchorName, err)

			case failedProc := <-psf.ProcessChan:
				return SpawnedExecutor{}, fmt.Errorf("createExecutor failed for %s on %s: %v", image, anchorName, failedProc.Errors)

			case err := <-psf.ErrChan:
				return SpawnedExecutor{}, fmt.Errorf("failure subscription error for %s on %s: %w", image, anchorName, err)
			}
		},

		removeExecutor: func(ctx context.Context, anchorName, containerID string) error {
			proc, err := cc.Submit(&core.FunctionSpec{
				FuncName:    "removeExecutor",
				Args:        []any{containerID},
				MaxExecTime: MAX_EXEC_TIME,
				Conditions: core.Conditions{
					ColonyName:    cfg.Colony.Name,
					ExecutorNames: []string{anchorName},
					ExecutorType:  "cpm-anchor",
				},
			}, cfg.User.Prvkey)
			if err != nil {
				return fmt.Errorf("submit failed: %w", err)
			}

			pss, err := cc.SubscribeProcess(
				cfg.Colony.Name,
				proc.ID,
				"cpm-anchor",
				core.SUCCESS,
				MAX_EXEC_TIME,
				cfg.User.Prvkey,
			)
			if err != nil {
				return fmt.Errorf("success subscribe failed: %w", err)
			}
			defer pss.Close()

			psf, err := cc.SubscribeProcess(
				cfg.Colony.Name,
				proc.ID,
				"cpm-anchor",
				core.FAILED,
				MAX_EXEC_TIME,
				cfg.User.Prvkey,
			)
			if err != nil {
				return fmt.Errorf("failure subscribe failed: %w", err)
			}
			defer psf.Close()

			select {
			case <-pss.ProcessChan:
				return nil

			case err := <-pss.ErrChan:
				return fmt.Errorf("success subscription error: %w", err)

			case failedProc := <-psf.ProcessChan:
				return fmt.Errorf("process failed: %v", failedProc.Errors)

			case err := <-psf.ErrChan:
				return fmt.Errorf("failure subscription error: %w", err)
			}
		},
	}

	svc := &deployer{
		repo:               sqlDeployRepo{db: database, queries: queries},
		runtime:            runtime,
		loadExecutorImages: loadExecutorImages,
		now:                time.Now,
		out:                cmd.OutOrStdout(),
	}

	return svc.Deploy(ctx, args[0])
}

func NewDeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy <pkg>",
		Short: "Deploy a package to the colony",
		Args:  cobra.ExactArgs(1),
		RunE:  deploy,
	}

	cmd.AddCommand(newDeployListCmd())
	cmd.AddCommand(newDeployRemoveCmd())
	cmd.AddCommand(newDeployFunctionCmd())

	return cmd
}
