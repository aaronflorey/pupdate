package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const configFileName = "config.yaml"

type userConfig struct {
	RootDirectory string `yaml:"root_directory"`
}

var userConfigDir = os.UserConfigDir

func resolveUserConfigPath() (string, error) {
	configDir, err := userConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve user config directory: %w", err)
	}

	return filepath.Join(configDir, "pupdate", configFileName), nil
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

func resolveUserConfig(cfg userConfig) (userConfig, error) {

	if cfg.RootDirectory != "" {
		resolved, err := expandConfiguredDirectory(cfg.RootDirectory)
		if err != nil {
			return userConfig{}, fmt.Errorf("failed to resolve root_directory: %w", err)
		}
		cfg.RootDirectory = resolved
	}

	return cfg, nil
}

func loadUserConfig() (userConfig, error) {
	path, err := resolveUserConfigPath()
	if err != nil {
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

func isWithinDirectory(path string, root string) bool {
	resolvedPath := resolveDirectory(path)
	resolvedRoot := resolveDirectory(root)
	if resolvedRoot == "" {
		return true
	}
	if resolvedPath == resolvedRoot {
		return true
	}

	rel, err := filepath.Rel(resolvedRoot, resolvedPath)
	if err != nil {
		return false
	}

	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
