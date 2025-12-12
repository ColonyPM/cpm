package pkg

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

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

func GetOrMakePackagesDirectory() (string, error) {
	var dir string

	switch runtime.GOOS {
	case "windows":
		base := os.Getenv("LOCALAPPDATA")
		if base == "" {
			base = os.Getenv("APPDATA")
		}
		if base == "" {
			return "", fmt.Errorf("LOCALAPPDATA/APPDATA not set")
		}
		dir = filepath.Join(base, "cpm", "packages")

	case "darwin":
		// Prefer XDG if the user explicitly set it
		base := os.Getenv("XDG_DATA_HOME")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("cannot determine home directory: %w", err)
			}
			// Standard macOS convention
			base = filepath.Join(home, "Library", "Application Support")
		}
		dir = filepath.Join(base, "cpm", "packages")

	case "linux", "freebsd", "openbsd", "netbsd", "dragonfly", "solaris":
		dataHome := os.Getenv("XDG_DATA_HOME")
		if dataHome == "" {
			home := os.Getenv("HOME")
			if home == "" {
				return "", fmt.Errorf("XDG_DATA_HOME and HOME not set")
			}
			dataHome = filepath.Join(home, ".local", "share")
		}
		dir = filepath.Join(dataHome, "cpm", "packages")

	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating packages directory %q: %w", dir, err)
	}

	return dir, nil
}

func GetPackageDirectory(pkgName string) (string, error) {
	pkgsDir, err := GetOrMakePackagesDirectory()
	if err != nil {
		return "", fmt.Errorf("packages directory: %w", err)
	}

	pkgDir := filepath.Join(pkgsDir, pkgName)

	info, err := os.Stat(pkgDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("package %q does not exist", pkgName)
		}
		return "", fmt.Errorf("stat package dir %q: %w", pkgDir, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("package path %q is not a directory", pkgDir)
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

func GetFunctionSpec(pkgName, fnSpecName string) (*core.FunctionSpec, error) {
	pkgsDir, err := GetOrMakePackagesDirectory()
	if err != nil {
		return nil, err
	}

	pkgDir := filepath.Join(pkgsDir, pkgName)

	info, err := os.Stat(pkgDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("package %q not installed (missing dir %s)", pkgName, pkgDir)
		}
		return nil, fmt.Errorf("stat package dir %q: %w", pkgDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("package path %q is not a directory", pkgDir)
	}

	templatesDir := filepath.Join(pkgDir, "templates")
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

	// Allow "fn" or "fn.json"
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

	fs, err := core.ConvertJSONToFunctionSpec(string(data))
	if err != nil {
		return nil, fmt.Errorf("parsing function spec %q: %w", specPath, err)
	}

	return fs, nil
}
