package pkgcmd

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

func searchPackages(cmd *cobra.Command, args []string) error {
	client := resty.New()

	query := args[0]
	var pkgs []string

	resp, err := client.R().
		SetContext(cmd.Context()).
		SetResult(&pkgs).
		SetQueryParam("q", query).
		Get(baseURL + "packages/packages")

	if err != nil {
		return err
	}

	if resp.IsSuccess() {
		if len(pkgs) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No packages found.")
			return nil
		}

		for _, pkg := range pkgs {
			fmt.Fprintln(cmd.OutOrStdout(), pkg)
		}
	} else {
		return fmt.Errorf("could not search the repository: %s", resp.Status())
	}

	return nil
}

func newPkgSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <pkg>",
		Short: "Search for a package",
		Args:  cobra.ExactArgs(1),
		RunE:  searchPackages,
	}
}
