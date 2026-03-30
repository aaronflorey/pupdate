package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/state"
	"github.com/spf13/cobra"
)

type runPayload struct {
	Directory  string                `json:"directory"`
	Ecosystems []runEcosystemPayload `json:"ecosystems"`
	Warnings   []runWarningPayload   `json:"warnings"`
}

type runEcosystemPayload struct {
	Ecosystem    detection.Ecosystem `json:"ecosystem"`
	Managers     []string            `json:"managers"`
	MatchedFiles []string            `json:"matched_files"`
	Warnings     []runWarningPayload `json:"warnings"`
}

type runWarningPayload struct {
	Code    detection.WarningCode `json:"code"`
	Message string                `json:"message"`
}

type ecosystemOutcome struct {
	Ecosystem string
	Succeeded bool
}

var detectFn = detection.Detect

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
		next.Ecosystems[outcome.Ecosystem] = state.EcosystemState{
			LastSuccessAt: state.FormatRFC3339UTC(now),
		}
	}

	return next
}

func newRunCmd() *cobra.Command {
	return &cobra.Command{
		Use: "run",
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := detectFn(".")
			if err != nil {
				return fmt.Errorf("detection failed: %w", err)
			}

			store := state.NewStore(".")
			currentState, _, err := store.Load()
			if err != nil {
				return fmt.Errorf("failed to load state: %w", err)
			}

			outcomes := []ecosystemOutcome{}

			payload := runPayload{
				Directory:  ".",
				Ecosystems: make([]runEcosystemPayload, 0, len(results)),
				Warnings:   make([]runWarningPayload, 0),
			}

			for _, result := range results {
				ecosystemWarnings := make([]runWarningPayload, 0, len(result.Warnings))
				for _, warning := range result.Warnings {
					wp := runWarningPayload{Code: warning.Code, Message: warning.Message}
					ecosystemWarnings = append(ecosystemWarnings, wp)
					payload.Warnings = append(payload.Warnings, wp)
				}

				payload.Ecosystems = append(payload.Ecosystems, runEcosystemPayload{
					Ecosystem:    result.Ecosystem,
					Managers:     result.Managers,
					MatchedFiles: result.MatchedFiles,
					Warnings:     ecosystemWarnings,
				})
			}

			encoder := json.NewEncoder(cmd.OutOrStdout())
			if err := encoder.Encode(payload); err != nil {
				return fmt.Errorf("failed to encode output: %w", err)
			}

			hasSuccess := false
			for _, outcome := range outcomes {
				if outcome.Succeeded {
					hasSuccess = true
					break
				}
			}
			if hasSuccess {
				updated := applySuccessfulOutcomes(time.Now().UTC(), currentState, outcomes)
				if err := store.Save(updated); err != nil {
					return fmt.Errorf("failed to save state: %w", err)
				}
			}

			return nil
		},
	}
}
