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
			"node":   {LastSuccessAt: "2026-03-01T10:00:00Z"},
			"python": {LastSuccessAt: "2026-03-01T11:00:00Z"},
		},
	}

	updated := applySuccessfulOutcomes(now, current, []ecosystemOutcome{
		{Ecosystem: "node", Succeeded: true},
		{Ecosystem: "python", Succeeded: false},
	})

	expected := map[string]state.EcosystemState{
		"node":   {LastSuccessAt: state.FormatRFC3339UTC(now)},
		"python": {LastSuccessAt: "2026-03-01T11:00:00Z"},
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
			"go": {LastSuccessAt: "2026-03-01T12:00:00Z"},
		},
	}

	updated := applySuccessfulOutcomes(now, current, []ecosystemOutcome{
		{Ecosystem: "go", Succeeded: false},
	})

	got := updated.Ecosystems["go"].LastSuccessAt
	if got != "2026-03-01T12:00:00Z" {
		t.Fatalf("expected failed ecosystem timestamp unchanged, got %q", got)
	}
}

func TestStateUpdateNoSuccessNoMutation(t *testing.T) {
	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)
	current := state.FileState{
		Version:    state.SchemaVersion,
		Ecosystems: map[string]state.EcosystemState{},
	}

	updated := applySuccessfulOutcomes(now, current, []ecosystemOutcome{
		{Ecosystem: "node", Succeeded: false},
		{Ecosystem: "python", Succeeded: false},
	})

	if len(updated.Ecosystems) != 0 {
		t.Fatalf("expected no mutations when no outcomes succeed, got %#v", updated.Ecosystems)
	}
}
