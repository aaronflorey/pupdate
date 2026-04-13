package main

import (
	"reflect"
	"testing"
	"time"

	"github.com/aaronflorey/pupdate/internal/state"
)

func TestStateUpdatesOnlyOnSuccessOutcomes(t *testing.T) {
	now := time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC)
	current := state.FileState{
		Version: state.SchemaVersion,
		Ecosystems: map[string]state.EcosystemState{
			"node":   {LastSuccessAt: "2026-03-01T10:00:00Z", Lockfiles: map[string]string{"bun.lock": "old"}},
			"python": {LastSuccessAt: "2026-03-01T11:00:00Z", Lockfiles: map[string]string{"requirements.txt": "old"}},
		},
	}

	updated := applySuccessfulOutcomes(now, current, []ecosystemOutcome{
		{StateKey: "node", Succeeded: true, Lockfiles: map[string]string{"bun.lock": "new"}},
		{StateKey: "python", Succeeded: false, Lockfiles: map[string]string{"requirements.txt": "new"}},
	})

	expected := map[string]state.EcosystemState{
		"node":   {LastSuccessAt: state.FormatRFC3339UTC(now), Lockfiles: map[string]string{"bun.lock": "new"}},
		"python": {LastSuccessAt: "2026-03-01T11:00:00Z", Lockfiles: map[string]string{"requirements.txt": "old"}},
	}
	if !reflect.DeepEqual(updated.Ecosystems, expected) {
		t.Fatalf("ecosystem state mismatch\ngot:  %#v\nwant: %#v", updated.Ecosystems, expected)
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

	updated := applySuccessfulOutcomes(now, current, []ecosystemOutcome{
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

	updated := applySuccessfulOutcomes(now, current, []ecosystemOutcome{
		{StateKey: "node", Succeeded: false, Lockfiles: map[string]string{"bun.lock": "new"}},
		{StateKey: "python", Succeeded: false, Lockfiles: map[string]string{"requirements.txt": "new"}},
	})

	if len(updated.Ecosystems) != 0 {
		t.Fatalf("expected no mutations when no outcomes succeed, got %#v", updated.Ecosystems)
	}
}
