package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gi8lino/loggo/internal/profile"
	"gopkg.in/yaml.v3"
)

const (
	localConfigName = ".loggo.yaml"
	envConfigPath   = "LOGGO_CONFIG"
	envProfile      = "LOGGO_PROFILE"
)

// EnvLookup resolves environment variable values.
type EnvLookup func(string) string

// Config describes the loggo configuration file.
type Config struct {
	DefaultProfile string                     `yaml:"defaultProfile"`
	Profiles       map[string]profile.Profile `yaml:"profiles"`
}

// Load loads global, local, or explicit loggo configuration.
func Load(explicitPath string, getEnv EnvLookup) (Config, []string, error) {
	cfg := Config{
		DefaultProfile: "auto",
		Profiles:       profile.Builtins(),
	}

	loaded := []string{}

	if globalPath, ok := globalConfigPath(); ok && fileExists(globalPath) {
		if err := mergeFile(&cfg, globalPath); err != nil {
			return Config{}, nil, err
		}

		loaded = append(loaded, globalPath)
	}

	if explicitPath == "" {
		explicitPath = getEnv(envConfigPath)
	}

	if explicitPath != "" {
		if err := mergeFile(&cfg, explicitPath); err != nil {
			return Config{}, nil, err
		}

		loaded = append(loaded, explicitPath)

		return cfg, loaded, cfg.Validate()
	}

	if localPath, ok := findLocalConfig(); ok {
		if err := mergeFile(&cfg, localPath); err != nil {
			return Config{}, nil, err
		}

		loaded = append(loaded, localPath)
	}

	return cfg, loaded, cfg.Validate()
}

// ResolveProfile resolves the active profile using CLI, environment, config, and fallback order.
func (c Config) ResolveProfile(name string, getEnv EnvLookup) (profile.Profile, error) {
	if name == "" {
		name = getEnv(envProfile)
	}
	if name == "" {
		name = c.DefaultProfile
	}
	if name == "" {
		name = "auto"
	}

	p, ok := c.Profiles[name]
	if !ok {
		return profile.Profile{}, fmt.Errorf("unknown profile %q", name)
	}

	return profile.Normalize(name, p), nil
}

// Validate ensures the config has usable defaults.
func (c Config) Validate() error {
	if c.DefaultProfile == "" {
		return errors.New("defaultProfile must not be empty")
	}
	if len(c.Profiles) == 0 {
		return errors.New("profiles must not be empty")
	}

	return nil
}

// mergeFile loads and merges a YAML config file into cfg.
func mergeFile(cfg *Config, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	next := Config{}
	if err := yaml.Unmarshal(content, &next); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}

	if next.DefaultProfile != "" {
		cfg.DefaultProfile = next.DefaultProfile
	}

	if next.Profiles != nil {
		if cfg.Profiles == nil {
			cfg.Profiles = map[string]profile.Profile{}
		}

		for name, p := range next.Profiles {
			cfg.Profiles[name] = profile.Normalize(name, p)
		}
	}

	return nil
}

// globalConfigPath returns the user-level config path.
func globalConfigPath() (string, bool) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", false
	}

	return filepath.Join(configDir, "loggo", "config.yaml"), true
}

// findLocalConfig walks upward from the current directory looking for .loggo.yaml.
func findLocalConfig() (string, bool) {
	dir, err := os.Getwd()
	if err != nil {
		return "", false
	}

	for {
		path := filepath.Join(dir, localConfigName)
		if fileExists(path) {
			return path, true
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}

		dir = parent
	}
}

// fileExists reports whether path exists and is a regular file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !info.IsDir()
}
