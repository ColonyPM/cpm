package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Colony  ColonyConfig  `yaml:"colony"`
	User    UserConfig    `yaml:"user"`
	Storage StorageConfig `yaml:"storage"`
}

type ServerConfig struct {
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
	TLS    bool   `yaml:"tls"`
	Prvkey string `yaml:"prvkey"`
}

type ColonyConfig struct {
	Name string `yaml:"name"`
}

type UserConfig struct {
	Prvkey string `yaml:"prvkey"`
}

type StorageConfig struct {
	Endpoint   string `yaml:"endpoint"`
	Accesskey  string `yaml:"accesskey"`
	Secretkey  string `yaml:"secretkey"`
	Region     string `yaml:"region"`
	Bucket     string `yaml:"bucket"`
	TLS        bool   `yaml:"tls"`
	Skipverify bool   `yaml:"skipverify"`
}

// Path returns the path to the config file based on XDG_CONFIG_HOME or platform defaults.
func Path() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "cpm", "config.yaml")
	}

	switch runtime.GOOS {
	case "windows":
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return filepath.Join(appdata, "cpm", "config.yaml")
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "AppData", "Roaming", "cpm", "config.yaml")
	default:
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "cpm", "config.yaml")
	}
}

// Load reads the config file and returns a validated Config.
func Load() (*Config, error) {
	path := Path()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found at %s", path)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid YAML in config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks that all required fields are present and valid.
func (c *Config) Validate() error {
	//Server
	if c.Server.Host == "" {
	 	return fmt.Errorf("missing required config: server.host")
	}
	if c.Server.Port == 0 {
		return fmt.Errorf("missing required config: server.port")
	}
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid config: server.port must be between 1-65535, got %d", c.Server.Port)
	}

	if c.Server.Prvkey == "" {
		return fmt.Errorf("missing required config: server.prvkey")
	}
	//Colony
	if c.Colony.Name == "" {
		return fmt.Errorf("missing required config: colony.name")
	}
	//User
	if c.User.Prvkey == "" {
		return fmt.Errorf("missing required config: user.prvkey")
	}
	//Storage
	if c.Storage.Endpoint == "" {
		return fmt.Errorf("missing required config: storage.endpoint")
	}
	if c.Storage.Accesskey == "" {
		return fmt.Errorf("missing required config: storage.accesskey")
	}
	if c.Storage.Secretkey == "" {
		return fmt.Errorf("missing required config: storage.secretkey")
	}
	if c.Storage.Bucket == "" {
		return fmt.Errorf("missing required config: storage.bucket")
	}

	return nil
}
