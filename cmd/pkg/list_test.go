package pkgcmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList_success(t *testing.T) {
	mockBaseDir := t.TempDir()
	oldGetDir := getPackagesDir
	t.Cleanup(func() { getPackagesDir = oldGetDir })
	getPackagesDir = func() string { return mockBaseDir }

	pkgA := filepath.Join(mockBaseDir, "alpha-pkg")
	os.MkdirAll(filepath.Join(pkgA, "1.0.0"), 0755)
	os.MkdirAll(filepath.Join(pkgA, "1.1.0"), 0755)

	pkgB := filepath.Join(mockBaseDir, "beta-pkg")
	os.MkdirAll(filepath.Join(pkgB, "2.0.0"), 0755)

	os.WriteFile(filepath.Join(mockBaseDir, "ignore-me.txt"), []byte("dummy"), 0644)

	cmd := newPkgListCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	expectedOutput := `alpha-pkg
  ├─ 1.0.0
  └─ 1.1.0
beta-pkg
  └─ 2.0.0`
	actualOutput := strings.TrimSpace(buf.String())
	assert.Equal(t, expectedOutput, actualOutput)
}

func TestList_MissingDirectory(t *testing.T) {
	oldGetDir := getPackagesDir
	t.Cleanup(func() { getPackagesDir = oldGetDir })
	getPackagesDir = func() string { return "/this/path/does/not/exist/" }

	cmd := newPkgListCmd()

	cmd.SetOut(new(bytes.Buffer))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading packages directory")
}
