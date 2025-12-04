package pkg

import (
	"io"

	"gopkg.in/yaml.v3"
)

type Executor struct {
	Name  string `yaml:"name"`
	Image string `yaml:"image"`
}

type DeploymentConfig struct {
	FuncSpecs []string   `yaml:"funcSpecs"`
	Workflows []string   `yaml:"workflows"`
	Executors []Executor `yaml:"executors"`
}

type Manifest struct {
	Name        string           `yaml:"name"`
	Version     string           `yaml:"version"`
	Description string           `yaml:"description"`
	Deprecated  bool             `yaml:"deprecated"`
	Deployments DeploymentConfig `yaml:"deployments"`
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

func WriteManifest(w io.Writer, manifest *Manifest) error {
	enc := yaml.NewEncoder(w)
	defer enc.Close()
	return enc.Encode(manifest)
}

func NewDefaultManifest(pkgName string) Manifest {
	return Manifest{
		Name:        pkgName,
		Version:     "0.0.1",
		Description: "A package",
		Deprecated:  false,
		Deployments: DeploymentConfig{
			FuncSpecs: []string{},
			Workflows: []string{},
			Executors: []Executor{},
		},
	}
}
