package main

import (
	"bytes"
	"strings"
	"testing"
)

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
	if !strings.Contains(out, "pupdate run --quiet") {
		t.Fatalf("expected bash snippet to invoke quiet run, got %q", out)
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
	if !strings.Contains(out, "pupdate run --quiet") {
		t.Fatalf("expected zsh snippet to invoke quiet run, got %q", out)
	}
}

func TestInitUnsupportedShellReturnsActionableError(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := newRootCmd()
	cmd.SetArgs([]string{"init", "--shell", "fish"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected init command to fail for unsupported shell")
	}
	if !strings.Contains(err.Error(), "supported shells: bash, zsh") {
		t.Fatalf("expected actionable supported-shells error, got %q", err.Error())
	}
}
