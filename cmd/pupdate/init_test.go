package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

// expectedHookInvocation returns the expected hook command prefix used in
// generated snippets, i.e. the shell-quoted resolved executable path followed
// by "hook".
func expectedHookInvocation(t *testing.T) string {
	t.Helper()
	return shellQuote(resolvedExecutablePath()) + " hook"
}

func TestInitBashSnippetIncludesHookAndQuietRun(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := newRootCmd()
	cmd.SetArgs([]string{"init", "--shell", "bash"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command failed: %v (stderr=%q)", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "PROMPT_COMMAND") {
		t.Fatalf("expected bash snippet to reference PROMPT_COMMAND, got %q", out)
	}
	if !strings.Contains(out, expectedHookInvocation(t)+" --quiet") {
		t.Fatalf("expected bash snippet to invoke quiet hook with resolved path, got %q", out)
	}
	if !strings.Contains(out, expectedHookInvocation(t)+" --quiet --async") {
		t.Fatalf("expected default bash snippet to use async mode, got %q", out)
	}
	if !strings.Contains(out, "[ \"$PWD\" != \"$HOME\" ]") {
		t.Fatalf("expected bash snippet to skip $HOME, got %q", out)
	}
	if strings.Contains(out, "2>/dev/null") {
		t.Fatalf("expected bash snippet to preserve stderr status output, got %q", out)
	}
}

func TestInitZshSnippetIncludesHooksAndQuietRun(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := newRootCmd()
	cmd.SetArgs([]string{"init", "--shell", "zsh"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command failed: %v (stderr=%q)", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "add-zsh-hook") {
		t.Fatalf("expected zsh snippet to add zsh hooks, got %q", out)
	}
	if !strings.Contains(out, expectedHookInvocation(t)+" --quiet") {
		t.Fatalf("expected zsh snippet to invoke quiet hook with resolved path, got %q", out)
	}
	if !strings.Contains(out, expectedHookInvocation(t)+" --quiet --async") {
		t.Fatalf("expected default zsh snippet to use async mode, got %q", out)
	}
	if !strings.Contains(out, "[ \"$PWD\" != \"$HOME\" ]") {
		t.Fatalf("expected zsh snippet to skip $HOME, got %q", out)
	}
	if strings.Contains(out, "2>/dev/null") {
		t.Fatalf("expected zsh snippet to preserve stderr status output, got %q", out)
	}
}

func TestInitFishSnippetIncludesHookAndQuietRun(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := newRootCmd()
	cmd.SetArgs([]string{"init", "--shell", "fish"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command failed: %v (stderr=%q)", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "--on-variable PWD") {
		t.Fatalf("expected fish snippet to use PWD variable hook, got %q", out)
	}
	if !strings.Contains(out, expectedHookInvocation(t)+" --quiet") {
		t.Fatalf("expected fish snippet to invoke quiet hook with resolved path, got %q", out)
	}
	if !strings.Contains(out, expectedHookInvocation(t)+" --quiet --async") {
		t.Fatalf("expected default fish snippet to use async mode, got %q", out)
	}
	if !strings.Contains(out, "test \"$PWD\" != \"$HOME\"") {
		t.Fatalf("expected fish snippet to skip $HOME, got %q", out)
	}
	if strings.Contains(out, "2>/dev/null") {
		t.Fatalf("expected fish snippet to preserve stderr status output, got %q", out)
	}
}

func TestInitAsyncModeUsesBackgroundHook(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := newRootCmd()
	cmd.SetArgs([]string{"init", "--shell", "bash", "--mode", "async"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command failed: %v (stderr=%q)", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, expectedHookInvocation(t)+" --quiet --async") {
		t.Fatalf("expected async init mode to launch background hook, got %q", out)
	}
}

func TestInitForegroundModeUsesSynchronousHook(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := newRootCmd()
	cmd.SetArgs([]string{"init", "--shell", "bash", "--mode", "foreground"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command failed: %v (stderr=%q)", err, stderr.String())
	}

	out := stdout.String()
	if strings.Contains(out, "--async") {
		t.Fatalf("expected foreground init mode to use synchronous hook, got %q", out)
	}
}

func TestInitUnsupportedShellReturnsActionableError(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := newRootCmd()
	cmd.SetArgs([]string{"init", "--shell", "tcsh"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected init command to fail for unsupported shell")
	}
	if !strings.Contains(err.Error(), "supported shells: bash, zsh, fish") {
		t.Fatalf("expected actionable supported-shells error, got %q", err.Error())
	}
}

func TestInitDefaultsToFishWhenShellEnvIsFish(t *testing.T) {
	t.Setenv("SHELL", "/usr/bin/fish")

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := newRootCmd()
	cmd.SetArgs([]string{"init"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command failed: %v (stderr=%q)", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "--on-variable PWD") {
		t.Fatalf("expected default snippet to be fish when SHELL is fish, got %q", out)
	}
	if !strings.Contains(out, expectedHookInvocation(t)+" --quiet") {
		t.Fatalf("expected fish default snippet to invoke quiet hook, got %q", out)
	}
	if !strings.Contains(out, expectedHookInvocation(t)+" --quiet --async") {
		t.Fatalf("expected fish default snippet to use async mode, got %q", out)
	}
}

func TestInitDefaultsToBashWhenShellEnvIsEmptyOrUnknown(t *testing.T) {
	tests := []struct {
		name     string
		shellEnv string
	}{
		{name: "empty shell env", shellEnv: ""},
		{name: "unknown shell env", shellEnv: "/usr/bin/tcsh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SHELL", tt.shellEnv)

			var stdout bytes.Buffer
			var stderr bytes.Buffer

			cmd := newRootCmd()
			cmd.SetArgs([]string{"init"})
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("init command failed: %v (stderr=%q)", err, stderr.String())
			}

			out := stdout.String()
			if !strings.Contains(out, "PROMPT_COMMAND") {
				t.Fatalf("expected default snippet to be bash, got %q", out)
			}
			if !strings.Contains(out, expectedHookInvocation(t)+" --quiet") {
				t.Fatalf("expected bash default snippet to invoke quiet hook, got %q", out)
			}
			if !strings.Contains(out, expectedHookInvocation(t)+" --quiet --async") {
				t.Fatalf("expected bash default snippet to use async mode, got %q", out)
			}
		})
	}
}

func TestInitRejectsUnsupportedHookMode(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"init", "--mode", "daemon"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected init command to fail for unsupported hook mode")
	}
	if !strings.Contains(err.Error(), "supported modes: foreground, async") {
		t.Fatalf("expected actionable hook-mode error, got %q", err.Error())
	}
}

func TestShellQuoteWrapsInSingleQuotes(t *testing.T) {
	got := shellQuote("/usr/local/bin/pupdate")
	want := "'/usr/local/bin/pupdate'"
	if got != want {
		t.Fatalf("shellQuote: got %q, want %q", got, want)
	}
}

func TestShellQuoteHandlesPathWithSpaces(t *testing.T) {
	got := shellQuote("/opt/my tools/pupdate")
	want := "'/opt/my tools/pupdate'"
	if got != want {
		t.Fatalf("shellQuote with spaces: got %q, want %q", got, want)
	}
}

func TestShellQuoteEscapesSingleQuotes(t *testing.T) {
	got := shellQuote("/opt/pat's tools/pupdate")
	want := "'/opt/pat'\\''s tools/pupdate'"
	if got != want {
		t.Fatalf("shellQuote with single quote: got %q, want %q", got, want)
	}
}

func TestResolvedExecutablePathReturnsNonEmpty(t *testing.T) {
	path := resolvedExecutablePath()
	if path == "" {
		t.Fatal("resolvedExecutablePath returned empty string")
	}
	if !filepath.IsAbs(path) {
		t.Fatalf("resolvedExecutablePath returned non-absolute path: %q", path)
	}
}
