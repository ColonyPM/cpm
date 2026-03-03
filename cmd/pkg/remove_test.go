package pkgcmd

import (
	"context"
	"os"
	"testing"
	"path/filepath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/spf13/cobra"
)
//Tests for the following:
// 1. Package location logic.
// 2. Package properly removed.


func TestRemoveCmd(t *testing.T) {
	//refer to Install_test.go
	ctx := context.Background()
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)

	tempdir := t.TempDir()
	GetPackages = func() string { return tempdir }
	testpkg := "testpkg"

	//1. Check handling for incorrect path
	//tempdir := t.TempDir()
	//pkgDir := tempdir
	require.NoError(t, os.MkdirAll(filepath.Join(tempdir,testpkg),0o755))
	require.NoError(t,os.WriteFile(filepath.Join(tempdir, "testpkg/Testfile.yaml"),[]byte("Test"),0o644))
	
	err := RemovePkg(cmd,[]string{"Fake"})
	assert.Error(t, err)
	assert.NotEmpty(t,filepath.Join(tempdir,testpkg)) 
	
	//2. Is package removed?
	//temptest := filepath.Join(tempdir,testpkg)
	err = RemovePkg(cmd,[]string{testpkg})
	require.NoError(t,err)
	require.Empty(t,testpkg)

}