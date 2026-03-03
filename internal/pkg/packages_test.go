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

func TestReadManifest(t *testing.T) {
	tempDir := t.TempDir()
	manifestPath := filepath.Join(tempDir, "manifest.yaml")
	content := []byte(`
name: mypkg
version: 1.0.0
description: mypkg
author: mypkg
deprecated: false
`)
	require.NoError(t, os.WriteFile(manifestPath, content, 0755))

	f, err := os.Open(manifestPath)
	require.NoError(t, err)
	defer f.Close()

	manifest, err := ReadManifest(f)
	require.NoError(t, err)
	assert.Equal(t, "mypkg", manifest.Name, "Manifest name doesn't match expected value")
	assert.Equal(t, "1.0.0", manifest.Version, "Manifest version doesn't match expected value")
	assert.Equal(t, "mypkg", manifest.Description, "Manifest description doesn't match expected value")
	assert.Equal(t, "mypkg", manifest.Author, "Manifest author doesn't match expected value")
	assert.Equal(t, false, manifest.Deprecated, "Manifest deprecated doesn't match expected value")
}



func TestWriteManifest(t *testing.T) {
	manifest := Manifest{
		Name:        "mypkg",
		Version:     "1.0.0",
		Description: "description",
		Author:      "author",
		Deprecated:  false,
	}

	tempFile, err := os.CreateTemp("", "manifest-*.yaml")
	require.NoError(t, err)
	defer tempFile.Close()

	err = WriteManifest(tempFile, manifest)
	require.NoError(t, err)

	content, err := os.ReadFile(tempFile.Name())
	require.NoError(t, err)

	assert.Contains(t, string(content), "name: mypkg")
	assert.Contains(t, string(content), "version: 1.0.0")
	assert.Contains(t, string(content), "description: description")
	assert.Contains(t, string(content), "author: author")
}

func TestNewDefaultManifest(t *testing.T) {
    pkgName := "mypkg"
    manifest := NewDefaultManifest(pkgName)

    assert.Equal(t, pkgName, manifest.Name, "Manifest name doesn't match expected value")
    assert.Equal(t, "0.0.1", manifest.Version, "Manifest version doesn't match expected value")
    assert.Equal(t, "A package", manifest.Description, "Manifest description doesn't match expected value")
    assert.Equal(t, "Your Name", manifest.Author, "Manifest author doesn't match expected value")
    assert.Equal(t, false, manifest.Deprecated, "Manifest deprecated doesn't match expected value")
    assert.Empty(t, manifest.Deployments.FuncSpecs, "Expected empty FuncSpecs array")
    assert.Empty(t, manifest.Deployments.Workflows, "Expected empty Workflows array")
    assert.Empty(t, manifest.Deployments.Executors, "Expected empty Executors array")
}

func TestGetPackagesDir(t *testing.T) {
    dir := GetPackagesDir()

    assert.Contains(t, dir, "cpm", "Packages directory should contain 'cpm'")
    assert.Contains(t, dir, "packages", "Packages directory should contain 'packages'")
}

func TestEnsurePackagesDir(t *testing.T) {
    dir, err := EnsurePackagesDir()
    assert.NoError(t, err)
    
    assert.DirExists(t, dir)
}

