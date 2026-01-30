package pkgcmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildArchive(t *testing.T) {
	tempDir := t.TempDir()

	// Create a temporary directory for the package, /tmp/.../testpkg
	pkgDir := filepath.Join(tempDir, "testpkg")
	require.NoError(t, os.Mkdir(pkgDir, 0755))

	// Create a temporary file for the package
	filePath := filepath.Join(pkgDir, "testfile.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("test content"), 0644))

	ctx := context.Background()

	// Build the archive (in memory)
	buf, err := buildArchive(ctx, pkgDir)
	require.NoError(t, err)
	require.NotNil(t, buf)
	require.Greater(t, buf.Len(), 0)

	// Verify the tar.gz contains the file with expected content
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

		// Read the connent of the file

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

	// Act
	got, err := getPkgPath([]string{})

	// Assert
	require.NoError(t, err)
	require.Equal(t, tempDir, got)
}

func TestGetPkgPath_WithArgs_ReturnsFirstArg(t *testing.T) {
	// Arrange
	want := filepath.Join("some", "path")

	// Act
	got, err := getPkgPath([]string{want})

	// Assert
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestGetPkgPath_MultiArgs_IgnorRest(t *testing.T){
	//Act
	got, err := getPkgPath([]string{"arg1", "arg2"})
	
	//Assert
	require.NoError(t, err)
	require.Equal(t, "arg1", got)
		
}

func TestValidateDir(t *testing.T){
	t.Run("nonexisten path", func(t *testing.T){
		err := validateDir("this/path/is/wrong")
		require.Error(t, err)
	})
	t.Run("path is a file", func(t *testing.T){
		file, err := os.CreateTemp(t.TempDir(), "testfile")
		require.NoError(t, err)
		defer os.Remove(file.Name())
		err = validateDir(file.Name())
		require.Error(t, err)
	})
	
	t.Run("path is a directory", func(t *testing.T){
		dir, err := os.MkdirTemp(t.TempDir(), "testdir")
		require.NoError(t, err)
		defer os.RemoveAll(dir)
		err = validateDir(dir)
		require.NoError(t, err)
		
	})
}



func TestUploadPackage(t *testing.T){
	
}