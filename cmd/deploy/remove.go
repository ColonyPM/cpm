package deploycmd

import (
	"fmt"
	"strings"

	"github.com/ColonyPM/cpm/internal/db"
	"github.com/ColonyPM/cpm/internal/storectx"
	"github.com/spf13/cobra"
)

func remove(cmd *cobra.Command, args []string) error {
	ctx := cmd.Root().Context()
	_, queries := storectx.GetDb(ctx)

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

	fmt.Printf("Revision: %s@%s (id=%d)\n", revision.PackageName, revision.Version, revision.ID)
	fmt.Println("Executors:")

	if len(executors) == 0 {
		fmt.Println(" - no executors")
		return nil
	}

	for _, exec := range executors {
		fmt.Printf(" - %s (anchor=%s, container=%s, image=%s)\n",
			exec.ExecutorName,
			exec.AnchorName,
			exec.ContainerID,
			exec.ImgName,
		)
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
