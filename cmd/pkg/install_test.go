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

	srcDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(srcDir, "templates"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "package.yaml"), []byte("name: demo\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "README.md"), []byte("# demo\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "values.yaml"), []byte("global:\n"), 0o644))

	buf, err := buildArchive(ctx, srcDir)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/packages/mypkg/download" {
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

	err = installPackage(cmd, []string{"mypkg"})
	require.NoError(t, err)

	top := filepath.Base(srcDir)

	require.FileExists(t, filepath.Join(pkgsDir, top, "package.yaml"))
	require.FileExists(t, filepath.Join(pkgsDir, top, "README.md"))
	require.FileExists(t, filepath.Join(pkgsDir, top, "values.yaml"))
	require.DirExists(t, filepath.Join(pkgsDir, top, "templates"))

}

