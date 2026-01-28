package pkgcmd

import (
	"bytes"
	"fmt"
	"os"
	"time"

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


func uploadPackage(cmd *cobra.Command, args []string) error {
	
	//Handle token
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return fmt.Errorf("read token flag: %w", err)
	}
	
	
	var pkgPath string
	// If arg is empty create package in current directory
	if len(args) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get cwd: %w", err)
		}
		pkgPath = cwd
	} else {
		pkgPath = args[0]
	}
	
	//Handle pkgpath errors
	info, err := os.Stat(pkgPath)
	if err != nil {
	    return fmt.Errorf("stat %s: %w", pkgPath, err)
	}
	if !info.IsDir() {
	    return fmt.Errorf("%s is not a directory", pkgPath)
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
		return fmt.Errorf("creating archive: %w", err)
	}

	client := resty.New().SetTimeout(30 * time.Second)
	resp, err := client.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
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
