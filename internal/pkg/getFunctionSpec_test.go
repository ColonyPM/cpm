package pkg

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFunctionSpec_ReturnIfPackageDirFails(t *testing.T) {
	old := getPackageDirectory
	t.Cleanup(func() { getPackageDirectory = old })

	wantErr := errors.New("error getting package directory")

	getPackageDirectory = func(string) (string, error) {
		return "", wantErr
	}

	fs, err := GetFunctionSpec("mypkg", "fn")
	require.Nil(t, fs)
	require.ErrorIs(t, err, wantErr)
}

func TestGetFunctionSpec_ErrIfTemplatesDirNotFound(t *testing.T) {
	old := getPackageDirectory
	t.Cleanup(func() { getPackageDirectory = old })

	pkgdir := t.TempDir()
	
	getPackageDirectory = func(string) (string, error) {
		return pkgdir, nil
	}
	
	fs, err := GetFunctionSpec("mypkg", "fn")
	require.Nil(t, fs)
	require.Error(t, err)
	
	assert.Contains(t, err.Error(), "templates dir")
	assert.Contains(t, err.Error(), "not found")
	assert.Contains(t, err.Error(), "for package")
}


func TestGetFunctionSpec_ErrIfTemplatesIsNotADir(t *testing.T) {
	old := getPackageDirectory
	t.Cleanup(func() { getPackageDirectory = old })

	pkgdir := t.TempDir()
	
	require.NoError(t, os.WriteFile(filepath.Join(pkgdir, "templates"), []byte{}, 0755))
	
	getPackageDirectory = func(string) (string, error) {
		return pkgdir, nil
	}
	
	fs, err := GetFunctionSpec("mypkg", "fn")
	require.Nil(t, fs)
	require.Error(t, err)
	
	assert.Contains(t, err.Error(), "is not a dir")

}

func TestGetFunctionSpec_ErrIfspecMissing(t *testing.T) {
	old := getPackageDirectory
	t.Cleanup(func() { getPackageDirectory = old })

	pkgdir := t.TempDir()
	
	require.NoError(t, os.MkdirAll(filepath.Join(pkgdir, "templates"), 0755))
	
	getPackageDirectory = func(string) (string, error) {
		return pkgdir, nil
	}
	
	fs, err := GetFunctionSpec("mypkg", "fn")
	require.Nil(t, fs)
	require.Error(t, err)
	
	assert.Contains(t, err.Error(), `function spec "fn.json" not found in`)
}


func TestGetFunctionSpec_ErrIfConvertFails(t *testing.T) {
	oldGet := getPackageDirectory
	oldConv := convertJSONToFunctionSpec
	t.Cleanup(func() { getPackageDirectory = oldGet; convertJSONToFunctionSpec = oldConv })

	pkgdir := t.TempDir()
	templatesDir := filepath.Join(pkgdir, "templates")
	require.NoError(t, os.MkdirAll(templatesDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(templatesDir, "fn.json"), []byte("{}"), 0755))
	
	getPackageDirectory = func(string) (string, error) {
		return pkgdir, nil
	}
	
	parseErr := errors.New("parse failed")
	convertJSONToFunctionSpec = func(string) (*core.FunctionSpec, error) {
		return nil, parseErr
	}
	
	fs, err := GetFunctionSpec("mypkg", "fn")
	require.Nil(t, fs)
	require.Error(t, err)
	require.ErrorIs(t, err, parseErr)
	assert.Contains(t, err.Error(), "parsing function spec")
}
