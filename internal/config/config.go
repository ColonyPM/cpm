package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Colonies ColoniesConfig `yaml:"colonies"`
	S3       S3Config       `yaml:"s3"`
	Minio    MinioConfig    `yaml:"minio"`
}

type ColoniesConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	TLS          bool   `yaml:"tls"`
	ColonyName   string `yaml:"colony_name"`
	ColonyPrvkey string `yaml:"colony_prvkey"`
	Prvkey       string `yaml:"prvkey"`
}

type S3Config struct {
	Endpoint   string `yaml:"endpoint"`
	Accesskey  string `yaml:"accesskey"`
	Secretkey  string `yaml:"secretkey"`
	Region     string `yaml:"region"`
	Bucket     string `yaml:"bucket"`
	TLS        bool   `yaml:"tls"`
	SkipVerify bool   `yaml:"skip_verify"`
}

type MinioConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
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
	// Colonies config
	if c.Colonies.Host == "" {
		return fmt.Errorf("missing required config: colonies.host")
	}
	if c.Colonies.Port == 0 {
		return fmt.Errorf("missing required config: colonies.port")
	}
	if c.Colonies.Port < 1 || c.Colonies.Port > 65535 {
		return fmt.Errorf("invalid config: colonies.port must be between 1-65535, got %d", c.Colonies.Port)
	}
	if c.Colonies.ColonyName == "" {
		return fmt.Errorf("missing required config: colonies.colony_name")
	}
	if c.Colonies.ColonyPrvkey == "" {
		return fmt.Errorf("missing required config: colonies.colony_prvkey")
	}
	if c.Colonies.Prvkey == "" {
		return fmt.Errorf("missing required config: colonies.prvkey")
	}

	// S3 config
	if c.S3.Endpoint == "" {
		return fmt.Errorf("missing required config: s3.endpoint")
	}
	if c.S3.Accesskey == "" {
		return fmt.Errorf("missing required config: s3.accesskey")
	}
	if c.S3.Secretkey == "" {
		return fmt.Errorf("missing required config: s3.secretkey")
	}
	if c.S3.Region == "" {
		return fmt.Errorf("missing required config: s3.region")
	}
	if c.S3.Bucket == "" {
		return fmt.Errorf("missing required config: s3.bucket")
	}

	// Minio config
	if c.Minio.User == "" {
		return fmt.Errorf("missing required config: minio.user")
	}
	if c.Minio.Password == "" {
		return fmt.Errorf("missing required config: minio.password")
	}

	return nil
}
