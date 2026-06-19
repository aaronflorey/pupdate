package main

import (
	"errors"
	"os"
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

func TestResolveUserConfigPathUsesPupdateConfigEnv(t *testing.T) {
	customPath := filepath.Join(string(filepath.Separator), "etc", "pupdate", "custom.yaml")
	t.Setenv("PUPDATE_CONFIG", customPath)

	path, err := resolveUserConfigPath()
	if err != nil {
		t.Fatalf("resolve user config path: %v", err)
	}

	if path != customPath {
		t.Fatalf("expected path %q, got %q", customPath, path)
	}
}

func TestResolveUserConfigPathResolvesRelativePupdateConfigEnv(t *testing.T) {
	baseDir := t.TempDir()
	withChdir(t, baseDir)
	t.Setenv("PUPDATE_CONFIG", "my-config.yaml")

	path, err := resolveUserConfigPath()
	if err != nil {
		t.Fatalf("resolve user config path: %v", err)
	}

	expected := filepath.Join(baseDir, "my-config.yaml")
	if path != expected {
		t.Fatalf("expected path %q, got %q", expected, path)
	}
}

func TestResolveUserConfigPathIgnoresEmptyPupdateConfigEnv(t *testing.T) {
	previousGOOS := runtimeGOOS
	previousUserConfigDir := userConfigDir

	t.Cleanup(func() {
		runtimeGOOS = previousGOOS
		userConfigDir = previousUserConfigDir
	})

	runtimeGOOS = "linux"
	t.Setenv("PUPDATE_CONFIG", "")
	customDir := filepath.Join(string(filepath.Separator), "home", "tester", ".config")
	userConfigDir = func() (string, error) {
		return customDir, nil
	}

	path, err := resolveUserConfigPath()
	if err != nil {
		t.Fatalf("resolve user config path: %v", err)
	}

	expected := filepath.Join(customDir, "pupdate", "config.yaml")
	if path != expected {
		t.Fatalf("expected path %q, got %q", expected, path)
	}
}

func TestResolveUserConfigResolvesRootDirectories(t *testing.T) {
	baseDir := t.TempDir()
	withChdir(t, baseDir)
	homeDir := filepath.Join(baseDir, "home")
	t.Setenv("HOME", homeDir)
	if err := os.MkdirAll(filepath.Join(baseDir, "workspace"), 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}

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

func TestResolveUserConfigResolvesWorkspaceGlobs(t *testing.T) {
	configured := userConfig{
		WorkspaceGlobs: []string{" apps/* ", "./services/*", ""},
	}

	resolved, err := resolveUserConfig(configured)
	if err != nil {
		t.Fatalf("resolve user config: %v", err)
	}

	expected := []string{"apps/*", "services/*"}
	if len(resolved.WorkspaceGlobs) != len(expected) {
		t.Fatalf("expected %d resolved workspace globs, got %d", len(expected), len(resolved.WorkspaceGlobs))
	}
	for i := range expected {
		if resolved.WorkspaceGlobs[i] != expected[i] {
			t.Fatalf("expected workspace_globs[%d] = %q, got %q", i, expected[i], resolved.WorkspaceGlobs[i])
		}
	}
}

func TestResolveUserConfigLeavesWorkspaceGlobsUnsetWhenMissing(t *testing.T) {
	resolved, err := resolveUserConfig(userConfig{})
	if err != nil {
		t.Fatalf("resolve user config: %v", err)
	}
	if len(resolved.WorkspaceGlobs) != 0 {
		t.Fatalf("expected workspace globs to remain unset, got %#v", resolved.WorkspaceGlobs)
	}
}

func TestResolveUserConfigResolvesFolderBlacklist(t *testing.T) {
	configured := userConfig{
		FolderBlacklist: []string{" node_modules ", "vendor", ""},
	}

	resolved, err := resolveUserConfig(configured)
	if err != nil {
		t.Fatalf("resolve user config: %v", err)
	}

	expected := []string{"node_modules", "vendor"}
	if len(resolved.FolderBlacklist) != len(expected) {
		t.Fatalf("expected %d resolved folder blacklist entries, got %d", len(expected), len(resolved.FolderBlacklist))
	}
	for i := range expected {
		if resolved.FolderBlacklist[i] != expected[i] {
			t.Fatalf("expected folder_blacklist[%d] = %q, got %q", i, expected[i], resolved.FolderBlacklist[i])
		}
	}
}

func TestResolveUserConfigLeavesFolderBlacklistUnsetWhenMissing(t *testing.T) {
	resolved, err := resolveUserConfig(userConfig{})
	if err != nil {
		t.Fatalf("resolve user config: %v", err)
	}
	if len(resolved.FolderBlacklist) != 0 {
		t.Fatalf("expected folder blacklist to remain unset, got %#v", resolved.FolderBlacklist)
	}
}

func TestResolveUserConfigReturnsIndexErrorForInvalidWorkspaceGlobsEntry(t *testing.T) {
	configured := userConfig{WorkspaceGlobs: []string{"apps/*", "../services/*"}}

	_, err := resolveUserConfig(configured)
	if err == nil {
		t.Fatalf("expected resolve user config to fail")
	}

	if !strings.Contains(err.Error(), "failed to resolve workspace_globs[1]") {
		t.Fatalf("expected indexed workspace_globs resolution error, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "must stay within the repository root") {
		t.Fatalf("expected workspace_globs validation reason, got %q", err.Error())
	}
}

func TestResolveUserConfigReturnsErrorForInvalidWorkspaceGlobPattern(t *testing.T) {
	configured := userConfig{WorkspaceGlobs: []string{"apps/["}}

	_, err := resolveUserConfig(configured)
	if err == nil {
		t.Fatalf("expected resolve user config to fail")
	}

	if !strings.Contains(err.Error(), "invalid glob pattern") {
		t.Fatalf("expected invalid glob pattern error, got %q", err.Error())
	}
}

func TestResolveUserConfigReturnsIndexErrorForInvalidFolderBlacklistEntry(t *testing.T) {
	configured := userConfig{FolderBlacklist: []string{"vendor", "foo/bar"}}

	_, err := resolveUserConfig(configured)
	if err == nil {
		t.Fatalf("expected resolve user config to fail")
	}

	if !strings.Contains(err.Error(), "failed to resolve folder_blacklist[1]") {
		t.Fatalf("expected indexed folder_blacklist resolution error, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "exact directory name, not a path") {
		t.Fatalf("expected folder_blacklist path validation reason, got %q", err.Error())
	}
}

func TestResolveUserConfigReturnsErrorForBackslashSeparatedFolderBlacklistEntry(t *testing.T) {
	configured := userConfig{FolderBlacklist: []string{`foo\bar`}}

	_, err := resolveUserConfig(configured)
	if err == nil {
		t.Fatalf("expected resolve user config to fail")
	}

	if !strings.Contains(err.Error(), "exact directory name, not a path") {
		t.Fatalf("expected folder_blacklist backslash path validation error, got %q", err.Error())
	}
}

func TestResolveUserConfigReturnsErrorForGlobFolderBlacklistEntry(t *testing.T) {
	configured := userConfig{FolderBlacklist: []string{"build*"}}

	_, err := resolveUserConfig(configured)
	if err == nil {
		t.Fatalf("expected resolve user config to fail")
	}

	if !strings.Contains(err.Error(), "exact directory name, not a glob") {
		t.Fatalf("expected folder_blacklist glob validation error, got %q", err.Error())
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

func TestResolveRunOptionsUsesConfigDefaults(t *testing.T) {
	quiet := true
	allowScripts := true

	options := resolveRunOptions(nil, userConfig{Quiet: &quiet, AllowScripts: &allowScripts}, false, false)

	if !options.Quiet {
		t.Fatal("expected quiet to default from config")
	}
	if !options.AllowScripts {
		t.Fatal("expected allow-scripts to default from config")
	}
}

func TestResolveRunOptionsPrefersExplicitFlags(t *testing.T) {
	quiet := true
	allowScripts := true

	cmd := newRunCmd()
	if err := cmd.ParseFlags([]string{"--quiet=false", "--allow-scripts=false"}); err != nil {
		t.Fatalf("parse flags: %v", err)
	}

	options := resolveRunOptions(cmd, userConfig{Quiet: &quiet, AllowScripts: &allowScripts}, false, false)

	if options.Quiet {
		t.Fatal("expected explicit --quiet=false to override config")
	}
	if options.AllowScripts {
		t.Fatal("expected explicit --allow-scripts=false to override config")
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

func TestIsTopLevelDirectoryWithinRootRequiresMatchingCase(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "tmp", "code")
	topLevelProject := filepath.Join(root, "my-project")
	configuredRoot := filepath.Join(string(filepath.Separator), "TMP", "CODE")

	if isTopLevelDirectoryWithinRoot(topLevelProject, configuredRoot) {
		t.Fatalf("expected top-level project not to match configured root with different casing")
	}
}
