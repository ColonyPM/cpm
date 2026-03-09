package deploycmd

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ColonyPM/cpm/internal/db"
	"github.com/ColonyPM/cpm/internal/storectx"
	"github.com/aquasecurity/table"
	"github.com/spf13/cobra"
)

func newListTable(w io.Writer) *table.Table {
	t := table.New(w)

	t.SetHeaderStyle(table.StyleBold)
	t.SetLineStyle(table.StyleBlue)
	t.SetDividers(table.UnicodeRoundedDividers)

	t.SetHeaders("Revision", "Executors", "Deploy Time")
	t.AddHeaders("Package", "Version", "Name", "Anchor", "Container ID", "Image", "Deploy Time")
	t.SetHeaderColSpans(0, 2, 4, 1)

	t.SetAutoMergeHeaders(true)
	t.SetAutoMerge(true)

	return t
}

const zeroWidthSpace = "\u200b"

type deployQ interface {
	ListRevisions(ctx context.Context) ([]db.Revision, error)
	ListExecutorsByRevision(ctx context.Context, revisionID int64) ([]db.Executor, error)
}

func runList(cmd *cobra.Command, args []string, q deployQ) error {
	t := newListTable(cmd.OutOrStdout())

	revisions, err := q.ListRevisions(cmd.Context())
	if err != nil {
		return err
	}

	for i, revision := range revisions {
		rowSuffix := strings.Repeat(zeroWidthSpace, i+1)

		executors, err := q.ListExecutorsByRevision(cmd.Context(), revision.ID)
		if err != nil {
			return err
		}

		if len(executors) == 0 {
			t.AddRow(
				revision.PackageName+rowSuffix,
				revision.Version+rowSuffix,
				"No Executors",
				"",
				"",
				"",
				revision.DeployTime.Format(time.RFC3339)+rowSuffix,
			)
			continue
		}

		for _, executor := range executors {
			t.AddRow(
				revision.PackageName+rowSuffix,
				revision.Version+rowSuffix,
				executor.ExecutorName,
				executor.AnchorName,
				fmt.Sprintf("%.10s...", executor.ContainerID),
				executor.ImgName,
				revision.DeployTime.Format(time.RFC3339)+rowSuffix,
			)
		}
	}

	t.Render()
	return nil
}

func list(cmd *cobra.Command, args []string) error {
	_, q := storectx.GetDb(cmd.Root().Context())
	return runList(cmd, args, q)
}

func newDeployListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List revisions",
		RunE:    list,
	}
}
