package pkg

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"gopkg.in/yaml.v3"

	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ExecutorSpec struct {
	Name  string `yaml:"name"`
	Image string `yaml:"img"`
}

type DeploymentConfig struct {
	FuncSpecs []string       `yaml:"funcSpecs"`
	Workflows []string       `yaml:"workflows"`
	Executors []ExecutorSpec `yaml:"executors"`
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
