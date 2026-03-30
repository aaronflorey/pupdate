package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFileReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	got, warnings, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("Load() warnings = %v, want none", warnings)
	}
	if got.Version != SchemaVersion {
		t.Fatalf("Load() version = %d, want %d", got.Version, SchemaVersion)
	}
	if len(got.Ecosystems) != 0 {
		t.Fatalf("Load() ecosystems len = %d, want 0", len(got.Ecosystems))
	}
}

func TestLoadCorruptFileReturnsWarningAndEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)
	if err := os.WriteFile(path, []byte("{\"version\":"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	store := NewStore(dir)
	got, warnings, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(warnings) != 1 {
		t.Fatalf("Load() warnings len = %d, want 1", len(warnings))
	}
	if warnings[0] != "state file is invalid; treating as empty" {
		t.Fatalf("Load() warning = %q, want %q", warnings[0], "state file is invalid; treating as empty")
	}
	if got.Version != SchemaVersion {
		t.Fatalf("Load() version = %d, want %d", got.Version, SchemaVersion)
	}
	if len(got.Ecosystems) != 0 {
		t.Fatalf("Load() ecosystems len = %d, want 0", len(got.Ecosystems))
	}
}

func TestSaveCreatesPupdateLazily(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf(".pupdate should not exist before Save(), stat error = %v", err)
	}

	store := NewStore(dir)
	if err := store.Save(Empty()); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf(".pupdate should exist after Save(), stat error = %v", err)
	}
}

func TestSaveThenLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	want := FileState{
		Version: SchemaVersion,
		Ecosystems: map[string]EcosystemState{
			"node":   {LastSuccessAt: "2026-03-27T10:00:00Z"},
			"python": {LastSuccessAt: "2026-03-27T11:00:00Z"},
		},
	}

	if err := store.Save(want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, warnings, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("Load() warnings = %v, want none", warnings)
	}
	if got.Version != want.Version {
		t.Fatalf("Load() version = %d, want %d", got.Version, want.Version)
	}
	if got.Ecosystems["node"].LastSuccessAt != want.Ecosystems["node"].LastSuccessAt {
		t.Fatalf("Load() node timestamp = %q, want %q", got.Ecosystems["node"].LastSuccessAt, want.Ecosystems["node"].LastSuccessAt)
	}
	if got.Ecosystems["python"].LastSuccessAt != want.Ecosystems["python"].LastSuccessAt {
		t.Fatalf("Load() python timestamp = %q, want %q", got.Ecosystems["python"].LastSuccessAt, want.Ecosystems["python"].LastSuccessAt)
	}
}
