package deploycmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	green = "\x1b[32m"
	red   = "\x1b[31m"
	reset = "\x1b[0m"

	greenOK = green + "✔" + reset
	redX    = red + "❌" + reset
)

type SpawnedExecutor struct {
	ExecutorName string
	AnchorName   string
	ContainerID  string
	ImgName      string
}

type deployRepo interface {
	RevisionExists(ctx context.Context, pkgName, version string) (bool, error)
	SaveDeployment(ctx context.Context, pkgName, version string, deployedAt time.Time, executors []SpawnedExecutor) error
}

type deployRuntime interface {
	ApprovedAnchors(ctx context.Context) ([]string, error)
	CreateExecutor(ctx context.Context, anchorName, image string) (SpawnedExecutor, error)
	RemoveExecutor(ctx context.Context, anchorName, containerID string) error
}

type deployer struct {
	repo               deployRepo
	runtime            deployRuntime
	loadExecutorImages func(ref string) ([]string, error)
	now                func() time.Time
	out                io.Writer
}

func (d *deployer) Deploy(ctx context.Context, ref string) error {
	pkgName, version, err := parsePackageRef(ref)
	if err != nil {
		return err
	}

	exists, err := d.repo.RevisionExists(ctx, pkgName, version)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%s has already been deployed", ref)
	}

	images, err := d.loadExecutorImages(ref)
	if err != nil {
		return err
	}

	anchors, err := d.runtime.ApprovedAnchors(ctx)
	if err != nil {
		return err
	}

	spawned := make([]SpawnedExecutor, 0, len(anchors)*len(images))

	for _, anchorName := range anchors {
		d.printf("[+] Spawning executors on anchor '%s'\n", anchorName)

		for _, image := range images {
			spawnedExec, err := d.runtime.CreateExecutor(ctx, anchorName, image)
			if err != nil {
				return d.failAndRollback(ctx, spawned, err)
			}

			spawned = append(spawned, spawnedExec)
			d.printf(" → %s %s\n", greenOK, image)
		}
	}

	nowFn := d.now
	if nowFn == nil {
		nowFn = time.Now
	}

	if err := d.repo.SaveDeployment(ctx, pkgName, version, nowFn(), spawned); err != nil {
		return d.failAndRollback(ctx, spawned, err)
	}

	return nil
}

func parsePackageRef(ref string) (pkgName, version string, err error) {
	pkgName, version, ok := strings.Cut(ref, "@")
	if !ok || pkgName == "" || version == "" {
		return "", "", fmt.Errorf("package must be in the format name@version")
	}
	return pkgName, version, nil
}

func (d *deployer) failAndRollback(ctx context.Context, spawned []SpawnedExecutor, cause error) error {
	cleanupErr := d.rollback(ctx, spawned)
	if cleanupErr != nil {
		return fmt.Errorf("%w (rollback failed: %v)", cause, cleanupErr)
	}
	return cause
}

func (d *deployer) rollback(ctx context.Context, spawned []SpawnedExecutor) error {
	if len(spawned) == 0 {
		return nil
	}

	d.printf("[-] Deployment failed, removing already spawned containers...\n")

	var cleanupErrs []string

	for i := len(spawned) - 1; i >= 0; i-- {
		exec := spawned[i]

		if err := d.runtime.RemoveExecutor(ctx, exec.AnchorName, exec.ContainerID); err != nil {
			d.printf(" → %s failed to remove %s: %v\n", redX, exec.ContainerID, err)
			cleanupErrs = append(cleanupErrs, fmt.Sprintf("remove %s: %v", exec.ContainerID, err))
			continue
		}

		d.printf(" → %s removed %s\n", greenOK, exec.ContainerID)
	}

	if len(cleanupErrs) > 0 {
		return errors.New(strings.Join(cleanupErrs, "; "))
	}

	return nil
}

func (d *deployer) printf(format string, args ...any) {
	if d.out == nil {
		return
	}
	fmt.Fprintf(d.out, format, args...)
}
