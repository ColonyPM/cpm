package pkg

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/colonyos/colonies/pkg/core"
	"gopkg.in/yaml.v3"
)

type ExecutorSpec struct {
	Name  string `yaml:"name"`
	Image string `yaml:"img"`
}

type DeploymentConfig struct {
	FuncSpecs []string       `yaml:"functionSpecs"`
	Workflows []string       `yaml:"workflows"`
	Executors []ExecutorSpec `yaml:"executors"`
	Setup     []string       `yaml:"setup"`
	Teardown  []string       `yaml:"teardown"`
}

type Manifest struct {
	Name        string           `yaml:"name"`
	Version     string           `yaml:"version"`
	Description string           `yaml:"description"`
	Author      string           `yaml:"author"`
	Deprecated  bool             `yaml:"deprecated"`
	Deployments DeploymentConfig `yaml:"deploy"`
}

func ReadManifest(r io.Reader) (*Manifest, error) {
	var manifest Manifest
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	if err := dec.Decode(&manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

func WriteManifest(w io.Writer, manifest Manifest) error {
	enc := yaml.NewEncoder(w)
	defer enc.Close()
	return enc.Encode(manifest)
}

func NewDefaultManifest(pkgName string) Manifest {
	return Manifest{
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
}

func GetPackagesDir() string {
	base, err := os.UserConfigDir()
	if err != nil {
		panic("GetPackagesDir: UserConfigDir returned an error; this environment is unsupported")
	}
	return filepath.Join(base, "cpm", "packages")
}

func EnsurePackagesDir() (string, error) {
	dir := GetPackagesDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating packages directory %q: %w", dir, err)
	}
	return dir, nil
}

func GetPackageDirectory(pkgName string) (string, error) {
	pkgsDir := GetPackagesDir()
	pkgDir := filepath.Join(pkgsDir, pkgName)

	info, err := os.Stat(pkgDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("package %q is not installed", pkgName)
		}
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%q is not a directory", pkgDir)
	}
	return pkgDir, nil
}

func GetPackageManifest(pkgName string) (*Manifest, error) {
	pkgDir, err := GetPackageDirectory(pkgName)
	if err != nil {
		return nil, err
	}

	manifestPath := filepath.Join(pkgDir, "package.yaml")

	f, err := os.Open(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("manifest %q not found for package %q", manifestPath, pkgName)
		}
		return nil, fmt.Errorf("opening manifest %q: %w", manifestPath, err)
	}
	defer f.Close()

	m, err := ReadManifest(f)
	if err != nil {
		return nil, fmt.Errorf("parsing manifest %q: %w", manifestPath, err)
	}

	return m, nil
}

var getPackageDirectory = GetPackageDirectory
var convertJSONToFunctionSpec = core.ConvertJSONToFunctionSpec

func GetFunctionSpec(pkgName, fnSpecName string) (*core.FunctionSpec, error) {
	// Re-use our validated getter.
	// This replaces your first 15 lines of stat/exist checks.
	pkgDir, err := getPackageDirectory(pkgName)
	if err != nil {
		return nil, err
	}

	templatesDir := filepath.Join(pkgDir, "templates")

	// Check templates dir
	tInfo, err := os.Stat(templatesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("templates dir %q not found for package %q", templatesDir, pkgName)
		}
		return nil, fmt.Errorf("stat templates dir %q: %w", templatesDir, err)
	}
	if !tInfo.IsDir() {
		return nil, fmt.Errorf("templates path %q is not a directory", templatesDir)
	}

	// Handle extension
	fileName := fnSpecName
	if filepath.Ext(fileName) == "" {
		fileName += ".json"
	}

	specPath := filepath.Join(templatesDir, fileName)

	data, err := os.ReadFile(specPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("function spec %q not found in %q", fileName, templatesDir)
		}
		return nil, fmt.Errorf("reading function spec %q: %w", specPath, err)
	}

	fs, err := convertJSONToFunctionSpec(string(data))
	if err != nil {
		return nil, fmt.Errorf("parsing function spec %q: %w", specPath, err)
	}

	return fs, nil
}

// In pkg/package.go (or similar)

func GetValuesPath(pkgName string) (string, error) {
	fileRootName, version, _ := strings.Cut(pkgName, "@")

	pkgDir, err := GetPackageDirectory(fileRootName)
	if err != nil {
		return "", err
	}
	pkgVersionDir := filepath.Join(pkgDir, version)

	valuesPath := filepath.Join(pkgVersionDir, "values.yaml")

	info, err := os.Stat(valuesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("values.yaml not found for package %q", pkgName)
		}
		return "", fmt.Errorf("stat values file %q: %w", valuesPath, err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("values path %q is a directory", valuesPath)
	}

	return valuesPath, nil
}
