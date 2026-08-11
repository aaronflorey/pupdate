package userenv

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

type Environment struct {
	GOOS               string
	Getenv             func(string) string
	UserConfigDir      func() (string, error)
	UserHomeDir        func() (string, error)
	CurrentUserHomeDir func() (string, error)
}

func System() Environment {
	return Environment{
		GOOS:          runtime.GOOS,
		Getenv:        os.Getenv,
		UserConfigDir: os.UserConfigDir,
		UserHomeDir:   os.UserHomeDir,
		CurrentUserHomeDir: func() (string, error) {
			current, err := user.Current()
			if err != nil {
				return "", err
			}
			return current.HomeDir, nil
		},
	}
}

func (e Environment) HomeDirectories() []string {
	candidates := make([]string, 0, 2)
	seen := make(map[string]struct{}, 2)
	for _, resolve := range []func() (string, error){e.UserHomeDir, e.CurrentUserHomeDir} {
		if resolve == nil {
			continue
		}
		home, err := resolve()
		if err != nil || strings.TrimSpace(home) == "" {
			continue
		}
		home = ResolveDirectory(home)
		if home == "" {
			continue
		}
		if _, exists := seen[home]; exists {
			continue
		}
		seen[home] = struct{}{}
		candidates = append(candidates, home)
	}
	return candidates
}

func (e Environment) ConfigDirectory() (string, error) {
	getenv := e.Getenv
	if getenv == nil {
		getenv = os.Getenv
	}
	if e.GOOS == "darwin" {
		if xdg := strings.TrimSpace(getenv("XDG_CONFIG_HOME")); xdg != "" {
			return xdg, nil
		}
		homes := e.HomeDirectories()
		if len(homes) == 0 {
			return "", fmt.Errorf("failed to resolve user home directory")
		}
		return filepath.Join(homes[0], ".config"), nil
	}

	resolve := e.UserConfigDir
	if resolve == nil {
		resolve = os.UserConfigDir
	}
	dir, err := resolve()
	if err != nil {
		return "", fmt.Errorf("failed to resolve user config directory: %w", err)
	}
	return dir, nil
}

func (e Environment) ExpandHome(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path, nil
	}
	homes := e.HomeDirectories()
	if len(homes) == 0 {
		return "", fmt.Errorf("expand ~: failed to resolve home directory")
	}
	if path == "~" {
		return homes[0], nil
	}
	return filepath.Join(homes[0], strings.TrimPrefix(path, "~/")), nil
}

func ResolveDirectory(path string) string {
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		path = resolved
	}
	return filepath.Clean(path)
}

func SameDirectory(left string, right string) bool {
	return ResolveDirectory(left) == ResolveDirectory(right)
}

func IsTopLevelDirectoryWithinRoot(path string, root string) bool {
	resolvedPath := ResolveDirectory(path)
	resolvedRoot := ResolveDirectory(root)
	if resolvedRoot == "" || resolvedPath == "" {
		return false
	}

	rel, err := filepath.Rel(resolvedRoot, resolvedPath)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	return !strings.Contains(rel, string(filepath.Separator))
}
