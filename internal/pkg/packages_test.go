package pkg

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestCase[In any] struct {
	Name      string
	Input     In
	Want      any
	WantErr   bool
	ErrSubstr string
}

func RunErrorTests[In any, Out any](t *testing.T, tests []TestCase[In], fn func(In) (Out, error)) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := fn(tt.Input)

			if tt.WantErr {
				require.Error(t, err, "Expected an error but got nil")
				assert.Contains(t, err.Error(), tt.ErrSubstr)

				assert.Empty(t, got, "Expected nil/empty result on error, but got a value")
			} else {
				require.NoError(t, err, "Expected no error but got one")

				if tt.Want != nil {
					assert.Equal(t, tt.Want, got, "Returned value did not match expected Want value")
				}
			}
		})
	}
}

func createMockDir(t *testing.T) string {
	t.Helper()
	originalGetDir := getPackagesDir
	t.Cleanup(func() {
		getPackagesDir = originalGetDir
	})

	mockBaseDir := t.TempDir()
	getPackagesDir = func() string {
		return mockBaseDir
	}
	return mockBaseDir
}

func mockGetPackageDirectory(t *testing.T, mockBehavior func(string) (string, error)) {
	t.Helper()
	old := getPackageDirectory
	t.Cleanup(func() { getPackageDirectory = old })
	getPackageDirectory = mockBehavior
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

	expected := &Manifest{
		Name:        "mypkg",
		Version:     "1.0.0",
		Description: "mypkg",
		Author:      "mypkg",
		Deprecated:  false,
	}
	assert.Equal(t, expected, manifest)
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

	expected := Manifest{
		Name:        pkgName,
		Version:     "0.0.1",
		Description: "A package",
		Author:      "Your Name",
		Deprecated:  false,
		Deployments: DeploymentConfig{
			FuncSpecs: []string{},
			Workflows: []string{},
			Executors: []ExecutorSpec{},
		},
	}

	assert.Equal(t, expected, manifest)
}

func TestGetPackagesDir(t *testing.T) {
	dir := GetPackagesDir()

	assert.Contains(t, dir, "cpm", "Packages directory should contain 'cpm'")
	assert.Contains(t, dir, "packages", "Packages directory should contain 'packages'")
}

func TestEnsurePackagesDir(t *testing.T) {
	originalGetDir := getPackagesDir

	defer func() { getPackagesDir = originalGetDir }()

	mockPath := t.TempDir()
	getPackagesDir = func() string {
		return mockPath
	}

	dir, err := EnsurePackagesDir()

	if err != nil {
		t.Fatalf("EnsurePackagesDir failed: %v", err)
	}
	if dir != mockPath {
		t.Errorf("expected %q, got %q", mockPath, dir)
	}
}

func TestGetPackageDirectory(t *testing.T) {
	mockBaseDir := createMockDir(t)

	validPkgPath := filepath.Join(mockBaseDir, "good-pkg", "1.0.0")
	if err := os.MkdirAll(validPkgPath, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	badPkgPath := filepath.Join(mockBaseDir, "file-pkg", "2.0.0")
	if err := os.MkdirAll(filepath.Dir(badPkgPath), 0755); err != nil {
		t.Fatalf("Failed to create base directory for file: %v", err)
	}
	if err := os.WriteFile(badPkgPath, []byte("dummy content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []TestCase[string]{
		{
			Name:    "Valid installed package",
			Input:   "good-pkg@1.0.0",
			Want:    validPkgPath,
			WantErr: false,
		},
		{
			Name:      "Missing @ symbol",
			Input:     "good-pkg",
			WantErr:   true,
			ErrSubstr: "please specify a version",
		},
		{
			Name:      "Empty version",
			Input:     "good-pkg@",
			WantErr:   true,
			ErrSubstr: "please specify a version",
		},
		{
			Name:      "Disallow 'latest'",
			Input:     "good-pkg@latest",
			WantErr:   true,
			ErrSubstr: "'latest' is not a valid version",
		},
		{
			Name:      "Package not installed",
			Input:     "missing-pkg@1.0.0",
			WantErr:   true,
			ErrSubstr: "is not installed",
		},
		{
			Name:      "Path is a file, not a directory",
			Input:     "file-pkg@2.0.0",
			WantErr:   true,
			ErrSubstr: "exists but is not a directory",
		},
	}

	RunErrorTests(t, tests, GetPackageDirectory)
}

func TestGetPackageManifest(t *testing.T) {
	mockBaseDir := createMockDir(t)

	validPkgDir := filepath.Join(mockBaseDir, "valid-pkg", "1.0.0")
	os.MkdirAll(validPkgDir, 0755)
	validYaml := []byte(`name: valid-pkg
version: 1.0.0
description: A perfectly fine package`)
	os.WriteFile(filepath.Join(validPkgDir, "package.yaml"), validYaml, 0644)

	missingYamlDir := filepath.Join(mockBaseDir, "missing-yaml-pkg", "1.0.0")
	os.MkdirAll(missingYamlDir, 0755)

	invalidYamlDir := filepath.Join(mockBaseDir, "invalid-yaml-pkg", "1.0.0")
	os.MkdirAll(invalidYamlDir, 0755)
	invalidYaml := []byte(`name: [this is completely broken yaml {}{}`)
	os.WriteFile(filepath.Join(invalidYamlDir, "package.yaml"), invalidYaml, 0644)

	tests := []TestCase[string]{
		{
			Name:    "Successful manifest read",
			Input:   "valid-pkg@1.0.0",
			WantErr: false,
		},
		{
			Name:      "GetPackageDirectory fails (missing @ version)",
			Input:     "valid-pkg",
			WantErr:   true,
			ErrSubstr: "please specify a version",
		},
		{
			Name:      "Package directory not found",
			Input:     "non-existent-pkg@1.0.0",
			WantErr:   true,
			ErrSubstr: "is not installed",
		},
		{
			Name:      "Manifest file missing",
			Input:     "missing-yaml-pkg@1.0.0",
			WantErr:   true,
			ErrSubstr: "not found for package",
		},
		{
			Name:      "Manifest parsing fails (invalid YAML)",
			Input:     "invalid-yaml-pkg@1.0.0",
			WantErr:   true,
			ErrSubstr: "parsing manifest",
		},
	}

	RunErrorTests(t, tests, GetPackageManifest)
}

type funcSpecInput struct {
	PkgName    string
	FnSpecName string
}

func TestGetFunctionSpec(t *testing.T) {
	mockBaseDir := t.TempDir()
	oldConv := convertJSONToFunctionSpec
	t.Cleanup(func() { convertJSONToFunctionSpec = oldConv })
	mockGetPackageDirectory(t, func(pkgName string) (string, error) {
		if pkgName == "dir-fail" {
			return "", errors.New("error getting package directory")
		}
		return filepath.Join(mockBaseDir, pkgName), nil
	})
	templatesDir := filepath.Join(mockBaseDir, "valid-pkg", "templates")
	os.MkdirAll(templatesDir, 0755)
	os.WriteFile(filepath.Join(templatesDir, "fn.json"), []byte("{}"), 0755)

	badTemplatesDir := filepath.Join(mockBaseDir, "bad-templates-pkg")
	os.MkdirAll(badTemplatesDir, 0755)
	os.WriteFile(filepath.Join(badTemplatesDir, "templates"), []byte{}, 0755)

	convertJSONToFunctionSpec = func(string) (*core.FunctionSpec, error) {
		return &core.FunctionSpec{}, nil
	}

	tests := []TestCase[funcSpecInput]{
		{
			Name:      "Fails on package directory error",
			Input:     funcSpecInput{"dir-fail", "fn"},
			WantErr:   true,
			ErrSubstr: "error getting package directory",
		},
		{
			Name:      "Fails if templates dir missing",
			Input:     funcSpecInput{"missing-templates-pkg", "fn"},
			WantErr:   true,
			ErrSubstr: "templates dir",
		},
		{
			Name:      "Fails if templates is a file",
			Input:     funcSpecInput{"bad-templates-pkg", "fn"},
			WantErr:   true,
			ErrSubstr: "is not a directory",
		},
		{
			Name:      "Fails if spec missing",
			Input:     funcSpecInput{"valid-pkg", "missing-fn"},
			WantErr:   true,
			ErrSubstr: "not found in",
		},
		{
			Name:    "Success path",
			Input:   funcSpecInput{"valid-pkg", "fn"},
			WantErr: false,
		},
	}

	RunErrorTests(t, tests, func(in funcSpecInput) (*core.FunctionSpec, error) {
		return GetFunctionSpec(in.PkgName, in.FnSpecName)
	})
}

func TestGetValuesPath(t *testing.T) {
	mockBaseDir := createMockDir(t)

	mockGetPackageDirectory(t, func(pkgName string) (string, error) {
		if pkgName == "fail-pkg@1.0.0" {
			return "", errors.New("forced directory error")
		}
		name, version, _ := strings.Cut(pkgName, "@")
		return filepath.Join(mockBaseDir, name, version), nil
	})

	validPkgDir := filepath.Join(mockBaseDir, "valid-pkg", "1.0.0")
	os.MkdirAll(validPkgDir, 0755)
	os.WriteFile(filepath.Join(validPkgDir, "values.yaml"), []byte("key: value"), 0644)

	noValuesPkgDir := filepath.Join(mockBaseDir, "no-values-pkg", "1.0.0")
	os.MkdirAll(noValuesPkgDir, 0755)

	dirValuesPkgDir := filepath.Join(mockBaseDir, "dir-values-pkg", "1.0.0")
	os.MkdirAll(filepath.Join(dirValuesPkgDir, "values.yaml"), 0755)

	tests := []TestCase[string]{
		{
			Name:    "Successful values.yaml path",
			Input:   "valid-pkg@1.0.0",
			WantErr: false,
		},
		{
			Name:      "getPackageDirectory fails",
			Input:     "fail-pkg@1.0.0",
			WantErr:   true,
			ErrSubstr: "forced directory error",
		},
		{
			Name:      "values.yaml missing",
			Input:     "no-values-pkg@1.0.0",
			WantErr:   true,
			ErrSubstr: "not found for package",
		},
		{
			Name:      "values.yaml is a directory",
			Input:     "dir-values-pkg@1.0.0",
			WantErr:   true,
			ErrSubstr: "is a directory",
		},
	}

	RunErrorTests(t, tests, GetValuesPath)
}
