package pkgcmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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

func extractPackage(cmd *cobra.Command, resp *resty.Response, installRoot string) error {
	format := archives.CompressedArchive{
		Compression: archives.Gz{},
		Extraction:  archives.Tar{},
	}

	err := format.Extract(cmd.Context(), bytes.NewReader(resp.Body()), func(ctx context.Context, f archives.FileInfo) error {
		info, err := f.Stat()
		if err != nil {
			return err
		}

		name := f.NameInArchive
		targetPath := filepath.Join(installRoot, name)

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

	return err
}

func installPackage(cmd *cobra.Command, args []string) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Prefix = fmt.Sprintf("Downloading %s ", args[0])
	s.Start()

	defer s.Stop()

	time.Sleep(1 * time.Second)
	client := newRestyClient()
	resp, err := client.R().
		SetError(&DownloadError{}).
		Get(baseURL + "packages/" + args[0] + "/download")
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to install package: %w", err)
	}

	if resp.IsError() {
		if apiErr, ok := resp.Error().(*DownloadError); ok && apiErr.Detail != "" {
			s.Stop()
			return fmt.Errorf("install failed %s", apiErr.Detail)
		}
		s.Stop()
		return fmt.Errorf("install failed %s", resp.Status())
	}

	fileRootName, version, hasAt := strings.Cut(args[0], "@")
	if !hasAt || version == "" {
		s.Stop()
		return fmt.Errorf("Please specify version (pkg@version or pkg@latest)")
	}

	if strings.Contains(fileRootName, "/") || strings.Contains(fileRootName, "..") {
		s.Stop()
		return fmt.Errorf("invalid package name")
	}

	if strings.Contains(version, "/") || strings.Contains(version, "..") {
		s.Stop()
		return fmt.Errorf("invalid package name")
	}

	pkgsDir := getPackagesDir()
	pkgRoot := filepath.Join(pkgsDir, fileRootName)

	if err := os.MkdirAll(pkgRoot, 0o755); err != nil {
		s.Stop()
		return fmt.Errorf("failed to create package root: %w", err)
	}

	tempDir, err := os.MkdirTemp(pkgRoot, ".install-*")
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = os.RemoveAll(tempDir)
		}
	}()

	if err := extractPackage(cmd, resp, tempDir); err != nil {
		s.Stop()
		return err
	}

	readManifest, err := os.Open(filepath.Join(tempDir, "package.yaml"))
	if err != nil {
		s.Stop()
		return fmt.Errorf("open manifest: %w", err)
	}

	manifest, err := pkg.ReadManifest(readManifest)
	_ = readManifest.Close()

	if err != nil {
		s.Stop()
		return fmt.Errorf("read manifest: %w", err)
	}
	if manifest.Version == "" {
		s.Stop()
		return fmt.Errorf("manifest missing version")
	}

	if version == "latest" {
		version = manifest.Version
	} else if manifest.Version != version {
		s.Stop()
		return fmt.Errorf("manifest version %q does not match desired version %q", manifest.Version, version)
	}

	if strings.Contains(version, "/") || strings.Contains(version, "..") {
		s.Stop()
		return fmt.Errorf("invalid manifest version")
	}

	installRoot := filepath.Join(pkgRoot, version)

	_ = os.RemoveAll(installRoot)
	if err := os.Rename(tempDir, installRoot); err != nil {
		s.Stop()
		return fmt.Errorf("finalize install: %w", err)
	}
	committed = true

	s.Stop()
	fmt.Printf("Installed %s@%s\n", fileRootName, version)
	return nil
}

func newPkgInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install <pkg@version>",
		Short: "Install a package",
		Args:  cobra.ExactArgs(1),
		RunE:  installPackage,
	}
}
