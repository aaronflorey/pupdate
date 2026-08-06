package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	configFileName        = "config.yaml"
	projectConfigFileName = ".pupdate.yaml"
)

type userConfig struct {
	RootDirectories []string `yaml:"root_directories"`
	WorkspaceGlobs  []string `yaml:"workspace_globs"`
	FolderBlacklist []string `yaml:"folder_blacklist"`
	Quiet           *bool    `yaml:"quiet"`
	AllowScripts    *bool    `yaml:"allow_scripts"`
}

type projectConfig struct {
	WorkspaceGlobs  []string `yaml:"workspace_globs"`
	FolderBlacklist []string `yaml:"folder_blacklist"`
}

var userConfigDir = os.UserConfigDir
var runtimeGOOS = runtime.GOOS

func resolveUserConfigPath() (string, error) {
	if env := os.Getenv("PUPDATE_CONFIG"); env != "" {
		path, err := filepath.Abs(env)
		if err != nil {
			return "", fmt.Errorf("failed to resolve PUPDATE_CONFIG path %q: %w", env, err)
		}
		return path, nil
	}

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
	if err := decodeConfigYAML(raw, &cfg); err != nil {
		return userConfig{}, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	return cfg, nil
}

func readProjectConfig(path string) (projectConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return projectConfig{}, nil
		}
		return projectConfig{}, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var cfg projectConfig
	if err := decodeConfigYAML(raw, &cfg); err != nil {
		return projectConfig{}, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return cfg, nil
}

func decodeConfigYAML(raw []byte, destination any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)
	if err := decoder.Decode(destination); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err == nil {
		return fmt.Errorf("multiple YAML documents are not supported")
	} else if !errors.Is(err, io.EOF) {
		return err
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

	if len(cfg.WorkspaceGlobs) > 0 {
		resolvedGlobs := make([]string, 0, len(cfg.WorkspaceGlobs))
		for index, configuredGlob := range cfg.WorkspaceGlobs {
			resolved, err := normalizeWorkspaceGlob(configuredGlob)
			if err != nil {
				return userConfig{}, fmt.Errorf("failed to resolve workspace_globs[%d]: %w", index, err)
			}
			if resolved == "" {
				continue
			}
			resolvedGlobs = append(resolvedGlobs, resolved)
		}
		cfg.WorkspaceGlobs = resolvedGlobs
	}

	if len(cfg.FolderBlacklist) > 0 {
		resolvedEntries := make([]string, 0, len(cfg.FolderBlacklist))
		for index, configuredEntry := range cfg.FolderBlacklist {
			resolved, err := normalizeFolderBlacklistEntry(configuredEntry)
			if err != nil {
				return userConfig{}, fmt.Errorf("failed to resolve folder_blacklist[%d]: %w", index, err)
			}
			if resolved == "" {
				continue
			}
			resolvedEntries = append(resolvedEntries, resolved)
		}
		cfg.FolderBlacklist = resolvedEntries
	}

	return cfg, nil
}

func resolveProjectConfig(cfg projectConfig) (projectConfig, error) {
	resolved, err := resolveUserConfig(userConfig{
		WorkspaceGlobs:  cfg.WorkspaceGlobs,
		FolderBlacklist: cfg.FolderBlacklist,
	})
	if err != nil {
		return projectConfig{}, err
	}
	return projectConfig{
		WorkspaceGlobs:  resolved.WorkspaceGlobs,
		FolderBlacklist: resolved.FolderBlacklist,
	}, nil
}

func mergeUserAndProjectConfig(user userConfig, project projectConfig) userConfig {
	user.WorkspaceGlobs = stableUnion(user.WorkspaceGlobs, project.WorkspaceGlobs)
	user.FolderBlacklist = stableUnion(user.FolderBlacklist, project.FolderBlacklist)
	return user
}

func stableUnion(first []string, second []string) []string {
	result := make([]string, 0, len(first)+len(second))
	seen := make(map[string]struct{}, len(first)+len(second))
	for _, values := range [][]string{first, second} {
		for _, value := range values {
			if _, ok := seen[value]; ok {
				continue
			}
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}
	return result
}

func expandHomeTilde(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		homeCandidates := homeDirectoryCandidates()
		if len(homeCandidates) == 0 {
			return "", fmt.Errorf("expand ~: failed to resolve home directory")
		}
		homeDir := homeCandidates[0]
		if path == "~" {
			return homeDir, nil
		}
		return filepath.Join(homeDir, strings.TrimPrefix(path, "~/")), nil
	}
	return path, nil
}

func expandConfiguredDirectory(path string) (string, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "", nil
	}

	var err error
	trimmed, err = expandHomeTilde(trimmed)
	if err != nil {
		return "", err
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

func normalizeWorkspaceGlob(pattern string) (string, error) {
	trimmed := strings.TrimSpace(pattern)
	if trimmed == "" {
		return "", nil
	}

	if _, err := filepath.Match(trimmed, ""); err != nil {
		return "", fmt.Errorf("invalid glob pattern: %w", err)
	}

	normalized := filepath.Clean(trimmed)
	if filepath.IsAbs(normalized) {
		return "", fmt.Errorf("workspace glob must be relative")
	}
	if normalized == "." {
		return "", fmt.Errorf("workspace glob must not match the repository root")
	}
	if normalized == ".." || strings.HasPrefix(normalized, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("workspace glob must stay within the repository root")
	}

	return normalized, nil
}

func normalizeFolderBlacklistEntry(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", nil
	}

	if strings.ContainsAny(trimmed, `*?[]`) {
		return "", fmt.Errorf("folder blacklist entry must be an exact directory name, not a glob")
	}

	if trimmed == "." || trimmed == ".." || strings.Contains(trimmed, "/") || strings.Contains(trimmed, `\`) {
		return "", fmt.Errorf("folder blacklist entry must be an exact directory name, not a path")
	}

	return trimmed, nil
}

func isTopLevelDirectoryWithinRoot(path string, root string) bool {
	resolvedPath := normalizeDirectoryForComparison(path)
	resolvedRoot := normalizeDirectoryForComparison(root)
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

func normalizeDirectoryForComparison(path string) string {
	resolved := resolveDirectory(path)
	if resolved == "" {
		return ""
	}

	return resolved
}
