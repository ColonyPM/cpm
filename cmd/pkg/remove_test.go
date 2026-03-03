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
	
	//0. Setup
	ctx := context.Background()
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	

	tempdir := t.TempDir()
	GetPackages = func() string { return tempdir }
	testpkg := "testpkg"

	require.NoError(t, os.MkdirAll(filepath.Join(tempdir,testpkg),0o755))
	require.NoError(t,os.WriteFile(filepath.Join(tempdir, "testpkg/Testfile.yaml"),[]byte("Test"),0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(tempdir,"testpkg/1.2.3"),0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(tempdir,"testpkg/3.2.1"),0o755))

	//1. Check handling for incorrect and correct path

	//1.2 fake path check
	err := RemovePkg(cmd,[]string{"Fake"})
	assert.Error(t, err)
	assert.NotEmpty(t,filepath.Join(tempdir,testpkg)) 
	

	//2.1 Version removal test
	err = RemovePkg(cmd,[]string{"testpkg@15"})
	assert.Error(t,err)
	err = RemovePkg(cmd,[]string{"testpkg@1.2.3"})
	assert.NoError(t,err)
	assert.DirExists(t,filepath.Join(tempdir,testpkg,"3.2.1"))
	assert.NoDirExists(t,filepath.Join(tempdir,testpkg,"1.2.3"))


	//2.2. Is package removed?
	require.Error(t,RemovePkg(cmd,[]string{testpkg}))
	//below i set the flag from default false to default true.
	cmd.Flags().BoolVar(&all, "all", true, "Remove ALL versions")
	err = RemovePkg(cmd,[]string{testpkg})
	require.NoError(t,err)
	require.Error(t,pkgexist(filepath.Join(tempdir,testpkg)))
}