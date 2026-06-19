package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResetDeletesStateFile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	statePath := filepath.Join(dir, ".pupdate")
	if err := os.WriteFile(statePath, []byte(`{"version":1,"ecosystems":{}}`), 0o644); err != nil {
		t.Fatalf("write state file: %v", err)
	}

	var stdout bytes.Buffer
	cmd := newRootCmd()
	cmd.SetArgs([]string{"reset"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("reset command failed: %v", err)
	}

	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Fatalf("expected state file to be removed, stat error = %v", err)
	}

	if !strings.Contains(stdout.String(), "removed") {
		t.Fatalf("expected 'removed' in output, got %q", stdout.String())
	}
}

func TestResetReportsNoFileFound(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	var stdout bytes.Buffer
	cmd := newRootCmd()
	cmd.SetArgs([]string{"reset"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("reset command failed: %v", err)
	}

	if !strings.Contains(stdout.String(), "nothing to reset") {
		t.Fatalf("expected 'nothing to reset' in output, got %q", stdout.String())
	}
}
