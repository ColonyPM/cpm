package pkgcmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/mholt/archives"
	"github.com/spf13/cobra"
)

var url = "https://colonypm.xyz/api/packages/upload"

type manifestSchema struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Author      string `yaml:"author"`
}

type uploadResponse struct {
	Url string `json:"url"`
}

type uploadError struct {
	Detail string `json:"detail"`
}

func buildArchive(ctx context.Context, dir string) (*bytes.Buffer, error) {
	dir = filepath.Clean(dir)

	if !strings.HasSuffix(dir, string(os.PathSeparator)) {
		dir += string(os.PathSeparator)
	}

	files, err := archives.FilesFromDisk(ctx, nil, map[string]string{dir: ""})
	if err != nil {
		return nil, fmt.Errorf("collect files: %w", err)
	}

	format := archives.CompressedArchive{
		Compression: archives.Gz{},
		Archival:    archives.Tar{},
	}

	var buf bytes.Buffer
	if err := format.Archive(ctx, &buf, files); err != nil {
		return nil, fmt.Errorf("create archive: %w", err)
	}
	return &buf, nil
}

func getPkgPath(args []string) (string, error) {
	if len(args) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get cwd: %w", err)
		}
		return cwd, nil
	}
	return args[0], nil
}

func validateDir(pkgPath string) error {
	info, err := os.Stat(pkgPath)
	if err != nil {
		return fmt.Errorf("stat %s: %w", pkgPath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", pkgPath)
	}
	return nil
}

func uploadPackage(cmd *cobra.Command, args []string) error {

	ctx := cmd.Context()

	//Handle token
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return fmt.Errorf("read token flag: %w", err)
	}

	var pkgPath string
	// If arg is empty create package in current directory
	getPath, err := getPkgPath(args)
	if err != nil {
		return fmt.Errorf("resolve package path: %w", err)
	}
	pkgPath = getPath

	//Validate package path
	if err := validateDir(pkgPath); err != nil {
		return fmt.Errorf("validate package path: %w", err)
	}

	// build archive
	buf, err := buildArchive(ctx, pkgPath)
	if err != nil {
		return fmt.Errorf("build archive: %w", err)
	}

	client := resty.New().SetTimeout(30 * time.Second)
	resp, err := client.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		SetMultipartField(
			"archive",
			"my-dir.tar.gz",
			"application/gzip",
			bytes.NewReader(buf.Bytes()),
		).
		SetResult(&uploadResponse{}).
		SetError(&uploadError{}).
		Post(url)
	if err != nil {
		return fmt.Errorf("uploading archive: %w", err)
	}

	if resp.IsError() {
		apiErr, ok := resp.Error().(*uploadError)
		if ok && apiErr != nil && apiErr.Detail != "" {
			return fmt.Errorf("upload failed (%d): %s", resp.StatusCode(), apiErr.Detail)
		}
		return fmt.Errorf("upload failed (%d): %s", resp.StatusCode(), resp.String())
	}

	fmt.Println(resp.Result().(*uploadResponse).Url)

	return nil
}

func newPkgUploadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload <dir>",
		Short: "Upload a package to the repository	",
		Args:  cobra.MaximumNArgs(1),
		RunE:  uploadPackage,
	}

	cmd.Flags().StringP("token", "t", "", "upload token")
	cmd.MarkFlagRequired("token")

	return cmd
}
