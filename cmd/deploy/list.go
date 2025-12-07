package deploycmd

import (
	"os"
	"strings"
	"time"

	"github.com/ColonyPM/cpm/internal/storectx"
	"github.com/aquasecurity/table"
	"github.com/spf13/cobra"
)

func newListTable() *table.Table {
	t := table.New(os.Stdout)

	t.SetHeaderStyle(table.StyleBold)
	t.SetLineStyle(table.StyleBlue)
	t.SetDividers(table.UnicodeRoundedDividers)

	t.SetHeaders("Package", "Executors", "Created At")
	t.AddHeaders("Package", "Name", "Container ID", "Created At")
	t.SetHeaderColSpans(0, 1, 2, 1)

	t.SetAutoMergeHeaders(true)
	t.SetAutoMerge(true)

	return t
}

const zeroWidthSpace = "\u200b"

func list(cmd *cobra.Command, args []string) error {
	t := newListTable()

	_, q := storectx.GetDb(cmd.Root().Context())

	deployments, err := q.ListDeployments(cmd.Context())
	if err != nil {
		return err
	}
	for i, deployment := range deployments {
		rowSuffix := strings.Repeat(zeroWidthSpace, i+1)
		executors, err := q.ListExecutorsByDeployment(cmd.Context(), deployment.ID)
		if err != nil {
			return err
		}

		if len(executors) == 0 {
			t.AddRow(deployment.PkgName+rowSuffix, "No Executors", "", deployment.DeployedAt.Format(time.RFC3339)+rowSuffix)
			continue
		}

		for _, executor := range executors {
			pkgName := deployment.PkgName + rowSuffix
			deployedAt := deployment.DeployedAt.Format(time.RFC3339) + rowSuffix
			t.AddRow(pkgName, executor.ExecutorName, executor.ContainerID, deployedAt)
		}
	}

	t.Render()

	return nil
}

func newDeployListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List deployments",
		RunE:    list,
	}
}
