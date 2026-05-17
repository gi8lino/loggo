package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// Load loads global, local, and explicit loggo configuration overlays.
func Load(explicitPaths []string, getEnv EnvLookup) (Config, []string, error) {
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

	if len(explicitPaths) == 0 {
		envPath := strings.TrimSpace(getEnv(envConfigPath))
		if envPath != "" {
			explicitPaths = []string{envPath}
		}
	}

	if len(explicitPaths) > 0 {
		for _, explicitPath := range explicitPaths {
			explicitPath = strings.TrimSpace(explicitPath)
			if explicitPath == "" {
				continue
			}

			if err := mergeFile(&cfg, explicitPath); err != nil {
				return Config{}, nil, err
			}

			loaded = append(loaded, explicitPath)
		}

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
			cfg.Profiles[name] = mergeProfile(name, cfg.Profiles[name], p)
		}
	}

	return nil
}

func mergeProfile(name string, base profile.Profile, overlay profile.Profile) profile.Profile {
	merged := base

	if overlay.Parser != "" {
		merged.Parser = overlay.Parser
	}
	if overlay.Regex != "" {
		merged.Regex = overlay.Regex
	}
	if overlay.TimestampField != "" {
		merged.TimestampField = overlay.TimestampField
	}
	if overlay.LevelField != "" {
		merged.LevelField = overlay.LevelField
	}
	if overlay.MessageField != "" {
		merged.MessageField = overlay.MessageField
	}
	if overlay.Fields != nil {
		merged.Fields = append([]string(nil), overlay.Fields...)
	}
	if overlay.FixedFields {
		merged.FixedFields = true
	}
	if overlay.HiddenFields != nil {
		merged.HiddenFields = append([]string(nil), overlay.HiddenFields...)
	}
	if overlay.Format != "" {
		merged.Format = overlay.Format
	}
	if overlay.Split.Delimiter != "" {
		merged.Split.Delimiter = overlay.Split.Delimiter
	}
	if overlay.Split.Fields != nil {
		merged.Split.Fields = append([]string(nil), overlay.Split.Fields...)
	}
	if overlay.Filters.Include != nil {
		merged.Filters.Include = append([]profile.Rule(nil), overlay.Filters.Include...)
	}
	if overlay.Filters.Exclude != nil {
		merged.Filters.Exclude = append([]profile.Rule(nil), overlay.Filters.Exclude...)
	}
	if overlay.Colors.Levels != nil {
		merged.Colors.Levels = cloneStringMap(overlay.Colors.Levels)
	}
	if overlay.Colors.Fields != nil {
		merged.Colors.Fields = cloneStringMap(overlay.Colors.Fields)
	}
	if overlay.Colors.Timestamp != "" {
		merged.Colors.Timestamp = overlay.Colors.Timestamp
	}
	if overlay.Colors.Message != "" {
		merged.Colors.Message = overlay.Colors.Message
	}

	return profile.Normalize(name, merged)
}

func cloneStringMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}

	cloned := make(map[string]string, len(input))
	for key, value := range input {
		cloned[key] = value
	}

	return cloned
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
