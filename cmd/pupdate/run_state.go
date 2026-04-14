package main

import (
	"fmt"
	"time"

	"github.com/aaronflorey/pupdate/internal/state"
)

type ecosystemOutcome struct {
	StateKey  string
	Succeeded bool
	Lockfiles map[string]string
}

func applySuccessfulOutcomes(now time.Time, current state.FileState, outcomes []ecosystemOutcome) state.FileState {
	next := state.FileState{
		Version:    current.Version,
		Ecosystems: make(map[string]state.EcosystemState, len(current.Ecosystems)),
	}
	for key, value := range current.Ecosystems {
		next.Ecosystems[key] = value
	}
	if next.Version == 0 {
		next.Version = state.SchemaVersion
	}

	for _, outcome := range outcomes {
		if !outcome.Succeeded {
			continue
		}
		existing := next.Ecosystems[outcome.StateKey]
		lockfiles := existing.Lockfiles
		if len(outcome.Lockfiles) > 0 {
			lockfiles = cloneLockfiles(outcome.Lockfiles)
		}
		next.Ecosystems[outcome.StateKey] = state.EcosystemState{
			LastSuccessAt: state.FormatRFC3339UTC(now),
			Lockfiles:     lockfiles,
		}
	}

	return next
}

func saveSuccessfulRunOutcomes(store state.Store, currentState state.FileState, outcomes []ecosystemOutcome) error {
	hasSuccess := false
	for _, outcome := range outcomes {
		if outcome.Succeeded {
			hasSuccess = true
			break
		}
	}
	if !hasSuccess {
		return nil
	}

	updated := applySuccessfulOutcomes(time.Now().UTC(), currentState, outcomes)
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
