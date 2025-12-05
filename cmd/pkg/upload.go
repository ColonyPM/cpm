package pkgcmd

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/mholt/archives"
	"github.com/spf13/cobra"
)

const url = "https://colonypm.xyz/api/packages/upload"

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

var Token string

func uploadPackage(cmd *cobra.Command, args []string) error {

	var pkgPath string
	// If arg is empty create package in current directory
	if len(args) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get cwd: %v", err)
		}
		pkgPath = cwd
	} else {
		pkgPath = args[0]
	}

	files, err := archives.FilesFromDisk(cmd.Context(), nil, map[string]string{
		pkgPath: "",
	})
	if err != nil {
		return err
	}

	format := archives.CompressedArchive{
		Compression: archives.Gz{},
		Archival:    archives.Tar{},
	}

	var archiveBytes bytes.Buffer

	if err := format.Archive(cmd.Context(), &archiveBytes, files); err != nil {
		return err
	}

	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", Token)).
		SetMultipartField(
			"archive",
			"my-dir.tar.gz",
			"application/gzip",
			bytes.NewReader(archiveBytes.Bytes()),
		).
		SetResult(&uploadResponse{}).
		SetError(&uploadError{}).
		Post(url)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return errors.New(resp.Error().(*uploadError).Detail)
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

	cmd.Flags().StringVarP(&Token, "token", "t", "", "upload token")
	cmd.MarkFlagRequired("token")

	return cmd
}
