package pkgcmd

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValuesCmd(t *testing.T) {
	tempDir := t.TempDir()

	fakeSourcePath := filepath.Join(tempDir, "fake_values.yaml")
	require.NoError(t, os.WriteFile(fakeSourcePath, []byte("data"), 0644))

	originalGetValues := getValuesPath
	defer func() { getValuesPath = originalGetValues }()

	getValuesPath = func(pkgName string) (string, error) {
		if pkgName == "valid-pkg" {
			return fakeSourcePath, nil
		}
		return "", os.ErrNotExist
	}

	tests := []struct {
		name         string
		args         []string
		prepareDest  func() string
		errContains  string
		expectedFile string
	}{
		{
			name: "Success: Explicit File",
			args: []string{"valid-pkg"},
			prepareDest: func() string {
				return filepath.Join(tempDir, "custom.yaml")
			},
		},
		{
			name: "Failure: Destination Exists",
			args: []string{"valid-pkg"},
			prepareDest: func() string {
				p := filepath.Join(tempDir, "exists.yaml")
				require.NoError(t, os.WriteFile(p, []byte("old"), 0644))
				return p
			},
			errContains: "destination already exists",
		},
		{
			name: "Success: Destination is Directory",
			args: []string{"valid-pkg"},
			prepareDest: func() string {
				dir := filepath.Join(tempDir, "configs")
				require.NoError(t, os.Mkdir(dir, 0755))
				return dir
			},
			expectedFile: filepath.Join(tempDir, "configs", "values.valid-pkg.yaml"),
		},
		{
			name: "Success: Create Parent Directories",
			args: []string{"valid-pkg"},
			prepareDest: func() string {
				return filepath.Join(tempDir, "missing", "folder", "values.yaml")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dest := tt.prepareDest()
			cmd := newValuesCmd()
			cmd.SetArgs(append(tt.args, dest))
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			err := cmd.Execute()

			if tt.errContains != "" {
				assert.ErrorContains(t, err, tt.errContains)
			} else {
				assert.NoError(t, err)

				finalPath := dest
				if tt.expectedFile != "" {
					finalPath = tt.expectedFile
				}

				assert.FileExists(t, finalPath)
			}
		})
	}
}
