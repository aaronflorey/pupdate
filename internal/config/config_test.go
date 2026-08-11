package config

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/aaronflorey/pupdate/internal/userenv"
)

func TestReadFileRejectsUnknownFieldsAndTrailingDocuments(t *testing.T) {
	tests := []struct {
		name    string
		content string
		project bool
	}{
		{name: "unknown user field", content: "allow_script: true\n"},
		{name: "trailing user document", content: "quiet: true\n---\nquiet: false\n"},
		{name: "forbidden project policy", content: "allow_scripts: true\n", project: true},
		{name: "forbidden project roots", content: "root_directories: [~/src]\n", project: true},
		{name: "trailing project document", content: "workspace_globs: [apps/*]\n---\nfolder_blacklist: [vendor]\n", project: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(path, []byte(tt.content), 0o600); err != nil {
				t.Fatalf("write config: %v", err)
			}
			var err error
			if tt.project {
				_, _, err = readFile[Project](path)
			} else {
				_, _, err = readFile[User](path)
			}
			if err == nil {
				t.Fatal("expected strict YAML error")
			}
		})
	}
}

func TestLoaderReadsAndMergesEachConfigSource(t *testing.T) {
	dir := t.TempDir()
	userPath := filepath.Join(dir, "user.yaml")
	if err := os.WriteFile(userPath, []byte("workspace_globs: [apps/*, shared/*]\nfolder_blacklist: [vendor]\nquiet: true\n"), 0o600); err != nil {
		t.Fatalf("write user config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ProjectFileName), []byte("workspace_globs: [shared/*, packages/*]\nfolder_blacklist: [dist]\n"), 0o600); err != nil {
		t.Fatalf("write project config: %v", err)
	}

	loader := Loader{Environment: userenv.Environment{Getenv: func(key string) string {
		if key == "PUPDATE_CONFIG" {
			return userPath
		}
		return ""
	}}}
	snapshot, err := loader.Load(dir)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !snapshot.User.Exists || !snapshot.Project.Exists {
		t.Fatalf("expected both config sources to exist: %#v", snapshot)
	}
	if !slices.Equal(snapshot.Effective.WorkspaceGlobs, []string{"apps/*", "shared/*", "packages/*"}) {
		t.Fatalf("unexpected workspace globs: %#v", snapshot.Effective.WorkspaceGlobs)
	}
	if !slices.Equal(snapshot.Effective.FolderBlacklist, []string{"vendor", "dist"}) {
		t.Fatalf("unexpected folder blacklist: %#v", snapshot.Effective.FolderBlacklist)
	}
	if snapshot.Effective.Quiet == nil || !*snapshot.Effective.Quiet {
		t.Fatalf("expected user policy to remain effective: %#v", snapshot.Effective)
	}
}

func TestLoaderReportsMissingSources(t *testing.T) {
	dir := t.TempDir()
	missingUser := filepath.Join(dir, "missing.yaml")
	loader := Loader{Environment: userenv.Environment{Getenv: func(key string) string {
		if key == "PUPDATE_CONFIG" {
			return missingUser
		}
		return ""
	}}}

	snapshot, err := loader.Load(dir)
	if err != nil {
		t.Fatalf("load missing config: %v", err)
	}
	if snapshot.User.Exists || snapshot.Project.Exists {
		t.Fatalf("expected missing sources, got %#v", snapshot)
	}
}
