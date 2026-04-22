package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

const configFileName = "config.yaml"
const defaultUserConfigContent = "root_directories: []\n"

type userConfig struct {
	RootDirectories []string `yaml:"root_directories"`
}

var userConfigDir = os.UserConfigDir
var runtimeGOOS = runtime.GOOS

func resolveUserConfigPath() (string, error) {
	configDir, err := resolveUserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "pupdate", configFileName), nil
}

func resolveUserConfigDir() (string, error) {
	if runtimeGOOS == "darwin" {
		xdgConfigHome := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME"))
		if xdgConfigHome != "" {
			return xdgConfigHome, nil
		}

		homeCandidates := homeDirectoryCandidates()
		if len(homeCandidates) == 0 {
			return "", fmt.Errorf("failed to resolve user home directory")
		}

		return filepath.Join(homeCandidates[0], ".config"), nil
	}

	configDir, err := userConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve user config directory: %w", err)
	}

	return configDir, nil
}

func readUserConfig(path string) (userConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return userConfig{}, nil
		}
		return userConfig{}, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var cfg userConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return userConfig{}, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	return cfg, nil
}

func ensureUserConfigExists(path string) error {
	configDir := filepath.Dir(path)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", configDir, err)
	}

	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to stat %s: %w", path, err)
	}

	if err := os.WriteFile(path, []byte(defaultUserConfigContent), 0o644); err != nil {
		return fmt.Errorf("failed to create %s: %w", path, err)
	}

	return nil
}

func resolveUserConfig(cfg userConfig) (userConfig, error) {
	if len(cfg.RootDirectories) > 0 {
		resolvedDirectories := make([]string, 0, len(cfg.RootDirectories))
		for index, configuredRoot := range cfg.RootDirectories {
			resolved, err := expandConfiguredDirectory(configuredRoot)
			if err != nil {
				return userConfig{}, fmt.Errorf("failed to resolve root_directories[%d]: %w", index, err)
			}
			if resolved == "" {
				continue
			}
			resolvedDirectories = append(resolvedDirectories, resolved)
		}
		cfg.RootDirectories = resolvedDirectories
	}

	return cfg, nil
}

func loadUserConfig() (userConfig, error) {
	path, err := resolveUserConfigPath()
	if err != nil {
		return userConfig{}, err
	}

	if err := ensureUserConfigExists(path); err != nil {
		return userConfig{}, err
	}

	cfg, err := readUserConfig(path)
	if err != nil {
		return userConfig{}, err
	}

	return resolveUserConfig(cfg)
}

func expandConfiguredDirectory(path string) (string, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "", nil
	}

	if trimmed == "~" || strings.HasPrefix(trimmed, "~/") {
		homeCandidates := homeDirectoryCandidates()
		if len(homeCandidates) == 0 {
			return "", fmt.Errorf("expand ~: failed to resolve home directory")
		}
		homeDir := homeCandidates[0]
		if trimmed == "~" {
			trimmed = homeDir
		} else {
			trimmed = filepath.Join(homeDir, strings.TrimPrefix(trimmed, "~/"))
		}
	}

	if !filepath.IsAbs(trimmed) {
		absolute, err := filepath.Abs(trimmed)
		if err != nil {
			return "", fmt.Errorf("resolve absolute path: %w", err)
		}
		trimmed = absolute
	}

	return resolveDirectory(trimmed), nil
}

func isTopLevelDirectoryWithinRoot(path string, root string) bool {
	resolvedPath := resolveDirectory(path)
	resolvedRoot := resolveDirectory(root)
	if resolvedRoot == "" || resolvedPath == "" {
		return false
	}

	rel, err := filepath.Rel(resolvedRoot, resolvedPath)
	if err != nil {
		return false
	}

	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}

	return !strings.Contains(rel, string(filepath.Separator))
}
