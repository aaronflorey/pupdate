package main

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveUserConfigPathUsesDotConfigOnDarwin(t *testing.T) {
	previousGOOS := runtimeGOOS
	previousUserHomeDir := userHomeDir
	previousCurrentUserHomeDir := currentUserHomeDir

	t.Cleanup(func() {
		runtimeGOOS = previousGOOS
		userHomeDir = previousUserHomeDir
		currentUserHomeDir = previousCurrentUserHomeDir
	})

	runtimeGOOS = "darwin"
	t.Setenv("XDG_CONFIG_HOME", "")
	userHomeDir = func() (string, error) {
		return filepath.Join(string(filepath.Separator), "Users", "tester"), nil
	}
	currentUserHomeDir = func() (string, error) {
		return "", errors.New("not available")
	}

	path, err := resolveUserConfigPath()
	if err != nil {
		t.Fatalf("resolve user config path: %v", err)
	}

	expected := filepath.Join(string(filepath.Separator), "Users", "tester", ".config", "pupdate", "config.yaml")
	if path != expected {
		t.Fatalf("expected path %q, got %q", expected, path)
	}
}

func TestResolveUserConfigPathUsesXDGConfigHomeOnDarwin(t *testing.T) {
	previousGOOS := runtimeGOOS

	t.Cleanup(func() {
		runtimeGOOS = previousGOOS
	})

	runtimeGOOS = "darwin"
	xdgConfigHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdgConfigHome)

	path, err := resolveUserConfigPath()
	if err != nil {
		t.Fatalf("resolve user config path: %v", err)
	}

	expected := filepath.Join(xdgConfigHome, "pupdate", "config.yaml")
	if path != expected {
		t.Fatalf("expected path %q, got %q", expected, path)
	}
}

func TestResolveUserConfigResolvesRootDirectories(t *testing.T) {
	baseDir := t.TempDir()
	withChdir(t, baseDir)
	homeDir := filepath.Join(baseDir, "home")
	t.Setenv("HOME", homeDir)

	configured := userConfig{
		RootDirectories: []string{"~/code", "./workspace"},
	}

	resolved, err := resolveUserConfig(configured)
	if err != nil {
		t.Fatalf("resolve user config: %v", err)
	}

	if len(resolved.RootDirectories) != 2 {
		t.Fatalf("expected 2 resolved roots, got %d", len(resolved.RootDirectories))
	}

	expectedHomeRoot := resolveDirectory(filepath.Join(homeDir, "code"))
	if resolved.RootDirectories[0] != expectedHomeRoot {
		t.Fatalf("expected first root %q, got %q", expectedHomeRoot, resolved.RootDirectories[0])
	}

	expectedWorkRoot := resolveDirectory(filepath.Join(baseDir, "workspace"))
	if resolved.RootDirectories[1] != expectedWorkRoot {
		t.Fatalf("expected second root %q, got %q", expectedWorkRoot, resolved.RootDirectories[1])
	}
}

func TestResolveUserConfigReturnsIndexErrorForInvalidRootDirectoriesEntry(t *testing.T) {
	previousUserHomeDir := userHomeDir
	previousCurrentUserHomeDir := currentUserHomeDir

	t.Cleanup(func() {
		userHomeDir = previousUserHomeDir
		currentUserHomeDir = previousCurrentUserHomeDir
	})

	userHomeDir = func() (string, error) {
		return "", errors.New("boom")
	}
	currentUserHomeDir = func() (string, error) {
		return "", errors.New("boom")
	}

	configured := userConfig{RootDirectories: []string{"/tmp/code", "~/projects"}}

	_, err := resolveUserConfig(configured)
	if err == nil {
		t.Fatalf("expected resolve user config to fail")
	}

	if !strings.Contains(err.Error(), "failed to resolve root_directories[1]") {
		t.Fatalf("expected indexed root_directories resolution error, got %q", err.Error())
	}
}

func TestIsTopLevelDirectoryWithinRoot(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "tmp", "code")
	topLevelProject := filepath.Join(root, "my-project")
	nestedProjectPath := filepath.Join(root, "my-project", "nested")
	outsidePath := filepath.Join(string(filepath.Separator), "tmp", "test", "code")

	if !isTopLevelDirectoryWithinRoot(topLevelProject, root) {
		t.Fatalf("expected top-level project to match configured root")
	}
	if isTopLevelDirectoryWithinRoot(root, root) {
		t.Fatalf("expected configured root directory itself to be rejected")
	}
	if isTopLevelDirectoryWithinRoot(nestedProjectPath, root) {
		t.Fatalf("expected nested project path to be rejected")
	}
	if isTopLevelDirectoryWithinRoot(outsidePath, root) {
		t.Fatalf("expected outside path to be rejected")
	}
}
