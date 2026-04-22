package main

import (
	"errors"
	"path/filepath"
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
