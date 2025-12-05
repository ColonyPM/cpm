package pkgcmd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

// /api/packages/{name}/download
const baseURL = "https://colonypm.xyz/api"

func installPackage(cmd *cobra.Command, args []string) error {
	client := resty.New()
	resp, err := client.R().
		Get(filepath.Join(baseURL, args[0], "download"))
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status())
	}

	// Save to file
	err = os.WriteFile("my-dir.tar.gz", resp.Body(), 0o644)
	if err != nil {
		return err
	}

	fmt.Println("Downloaded archive to my-dir.tar.gz")

	return nil
}

func newPkgInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install <pkg>",
		Short: "Install a package",
		Args:  cobra.ExactArgs(1),
		RunE:  installPackage,
	}
}
