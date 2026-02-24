package pkgcmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestInstallPackage_UsesDownloadedArchives(t *testing.T) {
	ctx := context.Background()
	pkgAndVersion := "mypkg@1.0.0"
	srcDir := t.TempDir()
	version := "1.0.0"
	require.NoError(t, os.MkdirAll(filepath.Join(srcDir, version, "templates"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(srcDir, version, "package.yaml"),
		[]byte("name: mypkg@1.0.0\nversion: 1.0.0\n"),
		0o644,
	))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, version, "readme.md"), []byte("# demo\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, version, "values.yaml"), []byte("global:\n"), 0o644))

	buf, err := buildArchive(ctx, filepath.Join(srcDir, version))
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/packages/"+pkgAndVersion+"/download" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf.Bytes())
	}))
	defer server.Close()

	oldBaseURL := baseURL
	oldGetPackagesDir := getPackagesDir
	oldNewRestyClient := newRestyClient
	t.Cleanup(func() {
		baseURL = oldBaseURL
		getPackagesDir = oldGetPackagesDir
		newRestyClient = oldNewRestyClient
	})

	baseURL = server.URL + "/api/" 
	pkgsDir := t.TempDir()
	getPackagesDir = func() string { return pkgsDir }

	cmd := &cobra.Command{}
	cmd.SetContext(ctx)

	err = installPackage(cmd, []string{pkgAndVersion})
	require.NoError(t, err)

	pkgName := "mypkg"

	require.FileExists(t, filepath.Join(pkgsDir, pkgName, version, "package.yaml"))
	require.FileExists(t, filepath.Join(pkgsDir, pkgName, version, "readme.md"))
	require.FileExists(t, filepath.Join(pkgsDir, pkgName, version, "values.yaml"))
	require.DirExists(t, filepath.Join(pkgsDir, pkgName, version, "templates"))

}
