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


func testremovecmd(t *testing.T) {
	//refer to Install_test.go
	ctx := context.Background()
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)

	//1. Check handling for incorrect path
	tempdir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tempdir,"testpkg"),0o644))
	require.NoError(t,os.WriteFile(filepath.Join(tempdir, "testpkg/Testfile.yaml"),[]byte("Test"),0o644))
	
	err := RemovePkg(cmd,[]string{"Fake"})
	assert.EqualError(t, err,"No package found.: %w")
	assert.NotEmpty(t,filepath.Join(tempdir,"testpkg")) 
	
	//2. Is package removed?
	temptest := filepath.Join(tempdir,"testpkg")
	RemovePkg(cmd,[]string{temptest})
	require.Empty(t,tempdir)

}