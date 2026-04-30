package main

import (
	"fmt"
	"maps"
	"time"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/state"
)

type ecosystemOutcome struct {
	StateKey         string
	Succeeded        bool
	RefreshMetadata  bool
	Lockfiles        map[string]string
	LockfileMetadata map[string]state.LockfileMetadata
}

func applyRunOutcomes(now time.Time, current state.FileState, activeResults []detection.DetectionResult, outcomes []ecosystemOutcome) state.FileState {
	activeStateKeys := make(map[string]struct{}, len(activeResults))
	for _, result := range activeResults {
		activeStateKeys[result.StateKey()] = struct{}{}
	}
	for _, outcome := range outcomes {
		activeStateKeys[outcome.StateKey] = struct{}{}
	}

	next := state.FileState{
		Version:    current.Version,
		Ecosystems: make(map[string]state.EcosystemState, len(activeStateKeys)),
	}
	for key, value := range current.Ecosystems {
		if _, ok := activeStateKeys[key]; !ok {
			continue
		}
		next.Ecosystems[key] = value
	}
	if next.Version == 0 {
		next.Version = state.SchemaVersion
	}

	for _, outcome := range outcomes {
		if !outcome.Succeeded && !outcome.RefreshMetadata {
			continue
		}

		existing := next.Ecosystems[outcome.StateKey]
		lockfiles := existing.Lockfiles
		lockfileMetadata := existing.LockfileMetadata
		if len(outcome.Lockfiles) > 0 {
			lockfiles = cloneLockfiles(outcome.Lockfiles)
		}
		if len(outcome.LockfileMetadata) > 0 {
			lockfileMetadata = cloneLockfileMetadata(outcome.LockfileMetadata)
		}
		lastSuccessAt := existing.LastSuccessAt
		if outcome.Succeeded {
			lastSuccessAt = state.FormatRFC3339UTC(now)
		}
		next.Ecosystems[outcome.StateKey] = state.EcosystemState{
			LastSuccessAt:    lastSuccessAt,
			Lockfiles:        lockfiles,
			LockfileMetadata: lockfileMetadata,
		}
	}

	return next
}

func saveRunOutcomes(store state.Store, currentState state.FileState, activeResults []detection.DetectionResult, outcomes []ecosystemOutcome) error {
	activeStateKeys := make(map[string]struct{}, len(activeResults))
	for _, result := range activeResults {
		activeStateKeys[result.StateKey()] = struct{}{}
	}

	hasPersistedChange := len(currentState.Ecosystems) != len(activeStateKeys)
	if !hasPersistedChange {
		for key := range currentState.Ecosystems {
			if _, ok := activeStateKeys[key]; !ok {
				hasPersistedChange = true
				break
			}
		}
	}

	for _, outcome := range outcomes {
		if outcome.Succeeded || outcome.RefreshMetadata {
			hasPersistedChange = true
			break
		}
	}
	if !hasPersistedChange {
		return nil
	}

	updated := applyRunOutcomes(time.Now().UTC(), currentState, activeResults, outcomes)
	if maps.EqualFunc(currentState.Ecosystems, updated.Ecosystems, func(left state.EcosystemState, right state.EcosystemState) bool {
		return maps.Equal(left.Lockfiles, right.Lockfiles) && maps.Equal(left.LockfileMetadata, right.LockfileMetadata) && left.LastSuccessAt == right.LastSuccessAt
	}) && currentState.Version == updated.Version {
		return nil
	}
	if err := store.Save(updated); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}
	return nil
}

func cloneLockfiles(lockfiles map[string]string) map[string]string {
	if len(lockfiles) == 0 {
		return map[string]string{}
	}
	cloned := make(map[string]string, len(lockfiles))
	for key, value := range lockfiles {
		cloned[key] = value
	}
	return cloned
}

func cloneLockfileMetadata(metadata map[string]state.LockfileMetadata) map[string]state.LockfileMetadata {
	if len(metadata) == 0 {
		return map[string]state.LockfileMetadata{}
	}
	cloned := make(map[string]state.LockfileMetadata, len(metadata))
	for key, value := range metadata {
		cloned[key] = value
	}
	return cloned
}
