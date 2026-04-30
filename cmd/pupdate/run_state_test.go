package main

import (
	"reflect"
	"testing"
	"time"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/state"
)

func TestStateUpdatesOnlyOnSuccessOutcomes(t *testing.T) {
	now := time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC)
	current := state.FileState{
		Version: state.SchemaVersion,
		Ecosystems: map[string]state.EcosystemState{
			"node":   {LastSuccessAt: "2026-03-01T10:00:00Z", Lockfiles: map[string]string{"bun.lock": "old"}, LockfileMetadata: map[string]state.LockfileMetadata{"bun.lock": {Size: 3}}},
			"python": {LastSuccessAt: "2026-03-01T11:00:00Z", Lockfiles: map[string]string{"requirements.txt": "old"}},
		},
	}

	updated := applyRunOutcomes(now, current, []detection.DetectionResult{{Ecosystem: detection.EcosystemNode}, {Ecosystem: detection.EcosystemPython}}, []ecosystemOutcome{
		{StateKey: "node", Succeeded: true, Lockfiles: map[string]string{"bun.lock": "new"}, LockfileMetadata: map[string]state.LockfileMetadata{"bun.lock": {Size: 4}}},
		{StateKey: "python", Succeeded: false, Lockfiles: map[string]string{"requirements.txt": "new"}},
	})

	expected := map[string]state.EcosystemState{
		"node":   {LastSuccessAt: state.FormatRFC3339UTC(now), Lockfiles: map[string]string{"bun.lock": "new"}, LockfileMetadata: map[string]state.LockfileMetadata{"bun.lock": {Size: 4}}},
		"python": {LastSuccessAt: "2026-03-01T11:00:00Z", Lockfiles: map[string]string{"requirements.txt": "old"}},
	}
	if !reflect.DeepEqual(updated.Ecosystems, expected) {
		t.Fatalf("ecosystem state mismatch\ngot:  %#v\nwant: %#v", updated.Ecosystems, expected)
	}
}

func TestStateUpdatePreservesExistingMetadataWhenOutcomeDoesNotReplaceIt(t *testing.T) {
	now := time.Date(2026, 3, 2, 15, 0, 0, 0, time.UTC)
	current := state.FileState{
		Version: state.SchemaVersion,
		Ecosystems: map[string]state.EcosystemState{
			"node": {
				LastSuccessAt: "2026-03-01T10:00:00Z",
				Lockfiles:     map[string]string{"bun.lock": "old"},
				LockfileMetadata: map[string]state.LockfileMetadata{
					"bun.lock": {Size: 3, ModTimeUnixNano: 1, Mode: "-rw-r--r--"},
				},
			},
		},
	}

	updated := applyRunOutcomes(now, current, []detection.DetectionResult{{Ecosystem: detection.EcosystemNode}}, []ecosystemOutcome{{
		StateKey:  "node",
		Succeeded: true,
		Lockfiles: map[string]string{"bun.lock": "new"},
	}})

	if !reflect.DeepEqual(updated.Ecosystems["node"].LockfileMetadata, current.Ecosystems["node"].LockfileMetadata) {
		t.Fatalf("expected metadata to be preserved when outcome does not provide replacement, got %#v", updated.Ecosystems["node"].LockfileMetadata)
	}
}

func TestStateUpdateDoesNotTouchFailedEcosystem(t *testing.T) {
	now := time.Date(2026, 3, 2, 12, 0, 0, 0, time.UTC)
	current := state.FileState{
		Version: state.SchemaVersion,
		Ecosystems: map[string]state.EcosystemState{
			"go": {LastSuccessAt: "2026-03-01T12:00:00Z", Lockfiles: map[string]string{"go.mod": "old"}},
		},
	}

	updated := applyRunOutcomes(now, current, []detection.DetectionResult{{Ecosystem: detection.EcosystemGo}}, []ecosystemOutcome{
		{StateKey: "go", Succeeded: false, Lockfiles: map[string]string{"go.mod": "new"}},
	})

	got := updated.Ecosystems["go"].LastSuccessAt
	if got != "2026-03-01T12:00:00Z" {
		t.Fatalf("expected failed ecosystem timestamp unchanged, got %q", got)
	}
	if updated.Ecosystems["go"].Lockfiles["go.mod"] != "old" {
		t.Fatalf("expected failed ecosystem lockfile hash unchanged, got %q", updated.Ecosystems["go"].Lockfiles["go.mod"])
	}
}

func TestStateUpdateNoSuccessNoMutation(t *testing.T) {
	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)
	current := state.FileState{
		Version:    state.SchemaVersion,
		Ecosystems: map[string]state.EcosystemState{},
	}

	updated := applyRunOutcomes(now, current, nil, []ecosystemOutcome{
		{StateKey: "node", Succeeded: false, Lockfiles: map[string]string{"bun.lock": "new"}},
		{StateKey: "python", Succeeded: false, Lockfiles: map[string]string{"requirements.txt": "new"}},
	})

	if len(updated.Ecosystems) != 0 {
		t.Fatalf("expected no mutations when no outcomes succeed, got %#v", updated.Ecosystems)
	}
}

func TestStateMetadataRefreshPreservesLastSuccessTimestamp(t *testing.T) {
	now := time.Date(2026, 3, 2, 16, 0, 0, 0, time.UTC)
	current := state.FileState{
		Version: state.SchemaVersion,
		Ecosystems: map[string]state.EcosystemState{
			"node": {
				LastSuccessAt: "2026-03-01T10:00:00Z",
				Lockfiles:     map[string]string{"bun.lock": "same"},
				LockfileMetadata: map[string]state.LockfileMetadata{
					"bun.lock": {Size: 3},
				},
			},
		},
	}

	updated := applyRunOutcomes(now, current, []detection.DetectionResult{{Ecosystem: detection.EcosystemNode}}, []ecosystemOutcome{{
		StateKey:        "node",
		RefreshMetadata: true,
		Lockfiles:       map[string]string{"bun.lock": "same"},
		LockfileMetadata: map[string]state.LockfileMetadata{
			"bun.lock": {Size: 3, FileID: "1:2", ChangeTimeUnixNano: 3},
		},
	}})

	got := updated.Ecosystems["node"]
	if got.LastSuccessAt != "2026-03-01T10:00:00Z" {
		t.Fatalf("expected metadata refresh to preserve last success timestamp, got %q", got.LastSuccessAt)
	}
	if !reflect.DeepEqual(got.LockfileMetadata, map[string]state.LockfileMetadata{"bun.lock": {Size: 3, FileID: "1:2", ChangeTimeUnixNano: 3}}) {
		t.Fatalf("unexpected refreshed metadata: %#v", got.LockfileMetadata)
	}
}

func TestStateUpdatePrunesUndetectedEcosystems(t *testing.T) {
	now := time.Date(2026, 3, 2, 17, 0, 0, 0, time.UTC)
	current := state.FileState{
		Version: state.SchemaVersion,
		Ecosystems: map[string]state.EcosystemState{
			"node":          {LastSuccessAt: "2026-03-01T10:00:00Z", Lockfiles: map[string]string{"bun.lock": "same"}},
			"node@frontend": {LastSuccessAt: "2026-03-01T11:00:00Z", Lockfiles: map[string]string{"frontend/package-lock.json": "stale"}},
		},
	}

	updated := applyRunOutcomes(now, current, []detection.DetectionResult{{Ecosystem: detection.EcosystemNode}}, nil)

	if _, ok := updated.Ecosystems["node@frontend"]; ok {
		t.Fatalf("expected stale subdirectory ecosystem to be pruned, got %#v", updated.Ecosystems)
	}
	if _, ok := updated.Ecosystems["node"]; !ok {
		t.Fatalf("expected active ecosystem to be retained, got %#v", updated.Ecosystems)
	}
}

func TestStateUpdateRetainsActiveEcosystemWithoutPersistedOutcome(t *testing.T) {
	now := time.Date(2026, 3, 2, 18, 0, 0, 0, time.UTC)
	current := state.FileState{
		Version: state.SchemaVersion,
		Ecosystems: map[string]state.EcosystemState{
			"node": {LastSuccessAt: "2026-03-01T10:00:00Z", Lockfiles: map[string]string{"bun.lock": "same"}},
		},
	}

	updated := applyRunOutcomes(now, current, []detection.DetectionResult{{Ecosystem: detection.EcosystemNode}}, nil)

	if !reflect.DeepEqual(updated.Ecosystems, current.Ecosystems) {
		t.Fatalf("expected active ecosystem without outcome to remain unchanged, got %#v", updated.Ecosystems)
	}
}
