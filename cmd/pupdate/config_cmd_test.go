package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestConfigShowsPathAndResolvedRootDirectories(t *testing.T) {
	homeDir := t.TempDir()
	configHome := filepath.Join(homeDir, ".config")
	configPath := filepath.Join(configHome, "pupdate", "config.yaml")
	writeFixtureFiles(t, configHome, filepath.Join("pupdate", "config.yaml"))
	if err := os.WriteFile(configPath, []byte("root_directories:\n  - ~/src\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", configHome)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := newRootCmd()
	cmd.SetArgs([]string{"config"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("config command failed: %v (stderr=%q)", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "path: "+configPath) {
		t.Fatalf("expected config path in output, got %q", out)
	}
	if !strings.Contains(out, "exists: true") {
		t.Fatalf("expected exists=true in output, got %q", out)
	}
	if !strings.Contains(out, "root_directories: ~/src") {
		t.Fatalf("expected raw root_directories in output, got %q", out)
	}
	expectedResolvedRoot, err := expandConfiguredDirectory("~/src")
	if err != nil {
		t.Fatalf("resolve expected root_directories value: %v", err)
	}
	if !strings.Contains(out, "root_directories_resolved: "+expectedResolvedRoot) {
		t.Fatalf("expected resolved root_directories in output, got %q", out)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestConfigShowsUnsetValuesWhenConfigIsMissing(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := newRootCmd()
	cmd.SetArgs([]string{"config"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("config command failed: %v (stderr=%q)", err, stderr.String())
	}

	out := stdout.String()
	expectedPath := filepath.Join(configHome, "pupdate", "config.yaml")
	if !strings.Contains(out, "path: "+expectedPath) {
		t.Fatalf("expected missing config path in output, got %q", out)
	}
	if !strings.Contains(out, "exists: true") {
		t.Fatalf("expected exists=true in output, got %q", out)
	}
	if !strings.Contains(out, "root_directories: (not set)") {
		t.Fatalf("expected unset root_directories in output, got %q", out)
	}
	if !strings.Contains(out, "root_directories_resolved: (not set)") {
		t.Fatalf("expected unset resolved root_directories in output, got %q", out)
	}

	rawConfig, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("expected command to create default config file: %v", err)
	}
	if string(rawConfig) != defaultUserConfigContent {
		t.Fatalf("expected default config content %q, got %q", defaultUserConfigContent, string(rawConfig))
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestConfigReturnsParseErrorWhenYAMLIsInvalid(t *testing.T) {
	homeDir := t.TempDir()
	configHome := filepath.Join(homeDir, ".config")
	configPath := filepath.Join(configHome, "pupdate", "config.yaml")
	writeFixtureFiles(t, configHome, filepath.Join("pupdate", "config.yaml"))
	if err := os.WriteFile(configPath, []byte("root_directories: [oops\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", configHome)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"config"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected config command to fail")
	}
	if !strings.Contains(err.Error(), "failed to parse "+configPath) {
		t.Fatalf("expected parse error with config path, got %q", err.Error())
	}
}

func TestConfigReturnsUserConfigDirResolutionError(t *testing.T) {
	t.Cleanup(func() {
		userConfigDir = os.UserConfigDir
		runtimeGOOS = runtime.GOOS
	})
	runtimeGOOS = "linux"
	userConfigDir = func() (string, error) {
		return "", errors.New("boom")
	}

	cmd := newRootCmd()
	cmd.SetArgs([]string{"config"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected config command to fail")
	}
	if !strings.Contains(err.Error(), "failed to resolve user config directory: boom") {
		t.Fatalf("expected config-dir resolution error, got %q", err.Error())
	}
}

func TestConfigReturnsReadErrorWhenConfigPathIsDirectory(t *testing.T) {
	configHome := t.TempDir()
	configPath := filepath.Join(configHome, "pupdate", "config.yaml")
	if err := os.MkdirAll(configPath, 0o755); err != nil {
		t.Fatalf("mkdir config path: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", configHome)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"config"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected config command to fail")
	}
	if !strings.Contains(err.Error(), "failed to read "+configPath) {
		t.Fatalf("expected read error with config path, got %q", err.Error())
	}
}
