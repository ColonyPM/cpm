package pkgcmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildArchive(t *testing.T) {
	tempDir := t.TempDir()

	pkgDir := filepath.Join(tempDir, "testpkg")
	require.NoError(t, os.Mkdir(pkgDir, 0755))

	filePath := filepath.Join(pkgDir, "testfile.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("test content"), 0644))

	ctx := context.Background()

	buf, err := buildArchive(ctx, pkgDir)
	require.NoError(t, err)
	require.NotNil(t, buf)
	require.Greater(t, buf.Len(), 0)

	got := readTarGzFromBuffer(t, buf.Bytes(), "testfile.txt")
	require.Equal(t, "test content", got)
}

func readTarGzFromBuffer(t *testing.T, data []byte, wantName string) string {
	t.Helper()

	gzr, err := gzip.NewReader(bytes.NewReader(data))
	require.NoError(t, err)
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		h, err := tr.Next()
		if err == io.EOF {
			require.FailNow(t, "file not found in archive", wantName)
		}
		require.NoError(t, err)

		if h.Name == wantName || filepath.Base(h.Name) == wantName {
			b, err := io.ReadAll(tr)
			require.NoError(t, err)
			return string(b)
		}
	}
}

func TestGetPackage_NoDir(t *testing.T) {
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(oldCwd) })

	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	got, err := getPkgPath([]string{})

	require.NoError(t, err)
	require.Equal(t, tempDir, got)
}

func TestGetPkgPath_WithArgs_ReturnsFirstArg(t *testing.T) {
	want := filepath.Join("some", "path")

	got, err := getPkgPath([]string{want})

	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestGetPkgPath_MultiArgs_IgnorRest(t *testing.T) {
	got, err := getPkgPath([]string{"arg1", "arg2"})

	require.NoError(t, err)
	require.Equal(t, "arg1", got)

}

func TestValidateDir(t *testing.T) {
	t.Run("nonexisten path", func(t *testing.T) {
		err := validateDir("this/path/is/wrong")
		require.Error(t, err)
	})
	t.Run("path is a file", func(t *testing.T) {
		file, err := os.CreateTemp(t.TempDir(), "testfile")
		require.NoError(t, err)
		defer os.Remove(file.Name())
		err = validateDir(file.Name())
		require.Error(t, err)
	})

	t.Run("path is a directory", func(t *testing.T) {
		dir, err := os.MkdirTemp(t.TempDir(), "testdir")
		require.NoError(t, err)
		defer os.RemoveAll(dir)
		err = validateDir(dir)
		require.NoError(t, err)

	})
}

func TestUploadCommand_Success(t *testing.T) {
	pkgDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "testfile.txt"), []byte("test content"), 0o644))

	type receivedRequest struct {
		Method          string
		Path            string
		Authorization   string
		Filename        string
		PartContentType string
		Archive         []byte
	}

	var got receivedRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.Method = r.Method
		got.Path = r.URL.Path
		got.Authorization = r.Header.Get("Authorization")

		file, header, err := r.FormFile("archive")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		got.Filename = header.Filename
		got.PartContentType = header.Header.Get("Content-Type")

		got.Archive, err = io.ReadAll(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"url":"https://colonypm.xyz/packages/testpkg"}`))
	}))
	defer srv.Close()

	oldURL := url
	url = srv.URL + "/api/packages/upload"
	t.Cleanup(func() { url = oldURL })

	out, err := executeUploadCommand(t, "--token", "secret-token", pkgDir)
	require.NoError(t, err)

	require.Equal(t, "https://colonypm.xyz/packages/testpkg\n", out)
	require.Equal(t, http.MethodPost, got.Method)
	require.Equal(t, "/api/packages/upload", got.Path)
	require.Equal(t, "Bearer secret-token", got.Authorization)
	require.Equal(t, "my-dir.tar.gz", got.Filename)
	require.Equal(t, "application/gzip", got.PartContentType)
	require.Equal(t, "test content", readTarGzFromBuffer(t, got.Archive, "testfile.txt"))
}

func TestUploadCommand_UsesCurrentDirectoryWhenNoArg(t *testing.T) {
	pkgDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "cwd.txt"), []byte("cwd content"), 0o644))

	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(pkgDir))
	t.Cleanup(func() { _ = os.Chdir(oldCwd) })

	var uploaded []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, _, err := r.FormFile("archive")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		uploaded, err = io.ReadAll(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"url":"https://colonypm.xyz/packages/from-cwd"}`))
	}))
	defer srv.Close()

	oldURL := url
	url = srv.URL + "/api/packages/upload"
	t.Cleanup(func() { url = oldURL })

	out, err := executeUploadCommand(t, "--token", "secret-token")
	require.NoError(t, err)
	require.Equal(t, "https://colonypm.xyz/packages/from-cwd\n", out)
	require.Equal(t, "cwd content", readTarGzFromBuffer(t, uploaded, "cwd.txt"))
}

func TestUploadCommand_APIErrorDetail(t *testing.T) {
	pkgDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "testfile.txt"), []byte("bad content"), 0o644))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"detail":"bad package"}`))
	}))
	defer srv.Close()

	oldURL := url
	url = srv.URL + "/api/packages/upload"
	t.Cleanup(func() { url = oldURL })

	out, err := executeUploadCommand(t, "--token", "secret-token", pkgDir)
	require.Error(t, err)
	require.Empty(t, out)
	require.ErrorContains(t, err, "upload failed (400): bad package")
}

func TestUploadCommand_InvalidDir_DoesNotHitServer(t *testing.T) {
	var calls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	oldURL := url
	url = srv.URL + "/api/packages/upload"
	t.Cleanup(func() { url = oldURL })

	missingDir := filepath.Join(t.TempDir(), "does-not-exist")

	out, err := executeUploadCommand(t, "--token", "secret-token", missingDir)
	require.Error(t, err)
	require.Empty(t, out)
	require.ErrorContains(t, err, "validate package path")
	require.EqualValues(t, 0, calls.Load())
}

func TestUploadCommand_RequiresToken(t *testing.T) {
	pkgDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "testfile.txt"), []byte("test content"), 0o644))

	out, err := executeUploadCommand(t, pkgDir)
	require.Error(t, err)
	require.Empty(t, out)
	require.ErrorContains(t, err, `required flag(s) "token" not set`)
}

func executeUploadCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()

	cmd := newPkgUploadCmd()
	cmd.SetArgs(args)

	return captureStdout(t, func() error {
		return cmd.ExecuteContext(context.Background())
	})
}

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	runErr := fn()

	require.NoError(t, w.Close())

	out, err := io.ReadAll(r)
	require.NoError(t, err)

	return string(out), runErr
}
