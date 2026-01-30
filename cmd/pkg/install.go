package pkgcmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ColonyPM/cpm/internal/pkg"
	"github.com/briandowns/spinner"
	"github.com/go-resty/resty/v2"
	"github.com/mholt/archives"
	"github.com/spf13/cobra"
)

// /api/packages/{name}/download
var baseURL = "https://colonypm.xyz/api/"
var getPackagesDir = pkg.GetPackagesDir
var newRestyClient = resty.New

type DownloadError struct {
	Detail string `json:"detail"`
}

func installPackage(cmd *cobra.Command, args []string) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Prefix = fmt.Sprintf("Downloading %s ", args[0])
	s.Start()
	
	defer s.Stop()

	client := newRestyClient()
	resp, err := client.R().
		SetError(&DownloadError{}).
		Get(baseURL + "packages/" + args[0] + "/download")
	if err != nil {
		return fmt.Errorf("failed to download package: %w", err)
	}

	if resp.IsError() {
		if apiErr, ok := resp.Error().(*DownloadError); ok && apiErr.Detail != "" {
			return fmt.Errorf("download failed %s", apiErr.Detail)
		}
		return fmt.Errorf("download failed %s", resp.Status())
	}

	pkgsDir := getPackagesDir() 

	format := archives.CompressedArchive{
		Compression: archives.Gz{},
		Extraction:  archives.Tar{},
	}

	err = format.Extract(cmd.Context(), bytes.NewReader(resp.Body()), func(ctx context.Context, f archives.FileInfo) error {
		info, err := f.Stat()
		if err != nil {
			return err
		}

		name := f.NameInArchive
		targetPath := filepath.Join(pkgsDir, name)

		if info.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}

		src, err := f.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		dst, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer dst.Close()

		_, err = io.Copy(dst, src)
		return err
	})

	if err != nil {
		s.Stop()
		return err
	}

	s.Stop()
	fmt.Printf("Downloaded %s\n", args[0])

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
