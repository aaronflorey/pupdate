package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aaronflorey/pupdate/internal/userenv"
	"gopkg.in/yaml.v3"
)

const (
	FileName        = "config.yaml"
	ProjectFileName = ".pupdate.yaml"
)

type User struct {
	RootDirectories []string `yaml:"root_directories"`
	WorkspaceGlobs  []string `yaml:"workspace_globs"`
	FolderBlacklist []string `yaml:"folder_blacklist"`
	Quiet           *bool    `yaml:"quiet"`
	AllowScripts    *bool    `yaml:"allow_scripts"`
}

type Project struct {
	WorkspaceGlobs  []string `yaml:"workspace_globs"`
	FolderBlacklist []string `yaml:"folder_blacklist"`
}

type Source[T any] struct {
	Path     string
	Exists   bool
	Raw      T
	Resolved T
}

type Snapshot struct {
	User      Source[User]
	Project   Source[Project]
	Effective User
}

type Loader struct {
	Environment userenv.Environment
}

func (l Loader) Load(repositoryDir string) (Snapshot, error) {
	userPath, err := l.UserPath(repositoryDir)
	if err != nil {
		return Snapshot{}, err
	}
	projectPath := filepath.Join(repositoryDir, ProjectFileName)

	rawUser, userExists, err := readFile[User](userPath)
	if err != nil {
		return Snapshot{}, err
	}
	resolvedUser, err := ResolveUser(rawUser, repositoryDir, l.Environment)
	if err != nil {
		return Snapshot{}, err
	}

	rawProject, projectExists, err := readFile[Project](projectPath)
	if err != nil {
		return Snapshot{}, err
	}
	resolvedProject, err := ResolveProject(rawProject)
	if err != nil {
		return Snapshot{}, err
	}

	return Snapshot{
		User: Source[User]{
			Path: userPath, Exists: userExists, Raw: rawUser, Resolved: resolvedUser,
		},
		Project: Source[Project]{
			Path: projectPath, Exists: projectExists, Raw: rawProject, Resolved: resolvedProject,
		},
		Effective: Merge(resolvedUser, resolvedProject),
	}, nil
}

func (l Loader) UserPath(repositoryDir string) (string, error) {
	getenv := l.Environment.Getenv
	if getenv == nil {
		getenv = os.Getenv
	}
	if configured := getenv("PUPDATE_CONFIG"); configured != "" {
		if filepath.IsAbs(configured) {
			return filepath.Clean(configured), nil
		}
		return filepath.Abs(filepath.Join(repositoryDir, configured))
	}

	configDir, err := l.Environment.ConfigDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "pupdate", FileName), nil
}

func readFile[T any](path string) (T, bool, error) {
	var value T
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return value, false, nil
		}
		return value, false, fmt.Errorf("failed to read %s: %w", path, err)
	}
	if err := decodeYAML(raw, &value); err != nil {
		return value, true, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return value, true, nil
}

func decodeYAML(raw []byte, destination any) error {
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

func ResolveUser(value User, repositoryDir string, environment userenv.Environment) (User, error) {
	resolvedRoots := make([]string, 0, len(value.RootDirectories))
	for index, configuredRoot := range value.RootDirectories {
		resolved, err := ExpandDirectory(configuredRoot, repositoryDir, environment)
		if err != nil {
			return User{}, fmt.Errorf("failed to resolve root_directories[%d]: %w", index, err)
		}
		if resolved != "" {
			resolvedRoots = append(resolvedRoots, resolved)
		}
	}
	value.RootDirectories = resolvedRoots

	resolvedGlobs, resolvedBlacklist, err := resolveTraversalSettings(value.WorkspaceGlobs, value.FolderBlacklist)
	if err != nil {
		return User{}, err
	}
	value.WorkspaceGlobs = resolvedGlobs
	value.FolderBlacklist = resolvedBlacklist
	return value, nil
}

func ResolveProject(value Project) (Project, error) {
	resolvedGlobs, resolvedBlacklist, err := resolveTraversalSettings(value.WorkspaceGlobs, value.FolderBlacklist)
	if err != nil {
		return Project{}, err
	}
	return Project{WorkspaceGlobs: resolvedGlobs, FolderBlacklist: resolvedBlacklist}, nil
}

func resolveTraversalSettings(workspaceGlobs []string, folderBlacklist []string) ([]string, []string, error) {
	resolvedGlobs := make([]string, 0, len(workspaceGlobs))
	for index, configuredGlob := range workspaceGlobs {
		resolved, err := NormalizeWorkspaceGlob(configuredGlob)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve workspace_globs[%d]: %w", index, err)
		}
		if resolved != "" {
			resolvedGlobs = append(resolvedGlobs, resolved)
		}
	}

	resolvedBlacklist := make([]string, 0, len(folderBlacklist))
	for index, configuredEntry := range folderBlacklist {
		resolved, err := NormalizeFolderBlacklistEntry(configuredEntry)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve folder_blacklist[%d]: %w", index, err)
		}
		if resolved != "" {
			resolvedBlacklist = append(resolvedBlacklist, resolved)
		}
	}
	return resolvedGlobs, resolvedBlacklist, nil
}

func Merge(user User, project Project) User {
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

func ExpandDirectory(path string, repositoryDir string, environment userenv.Environment) (string, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "", nil
	}

	expanded, err := environment.ExpandHome(trimmed)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(expanded) {
		expanded = filepath.Join(repositoryDir, expanded)
	}
	return userenv.ResolveDirectory(expanded), nil
}

func NormalizeWorkspaceGlob(pattern string) (string, error) {
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

func NormalizeFolderBlacklistEntry(name string) (string, error) {
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
