package main

import (
	"encoding/json"
	"fmt"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/spf13/cobra"
)

type runPayload struct {
	Directory  string                `json:"directory"`
	Ecosystems []runEcosystemPayload `json:"ecosystems"`
	Warnings   []runWarningPayload   `json:"warnings"`
}

type runEcosystemPayload struct {
	Ecosystem    detection.Ecosystem `json:"ecosystem"`
	Directory    string              `json:"directory,omitempty"`
	Managers     []string            `json:"managers"`
	MatchedFiles []string            `json:"matched_files"`
	Warnings     []runWarningPayload `json:"warnings"`
}

type runWarningPayload struct {
	Code    detection.WarningCode `json:"code"`
	Message string                `json:"message"`
}

func writeRunPayload(cmd *cobra.Command, quiet bool, results []detection.DetectionResult) error {
	if quiet {
		return nil
	}

	encoder := json.NewEncoder(cmd.OutOrStdout())
	if err := encoder.Encode(buildRunPayload(results)); err != nil {
		return fmt.Errorf("failed to encode output: %w", err)
	}
	return nil
}

func buildRunPayload(results []detection.DetectionResult) runPayload {
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

		payloadDirectory := result.Directory
		if payloadDirectory == "." {
			payloadDirectory = ""
		}

		payload.Ecosystems = append(payload.Ecosystems, runEcosystemPayload{
			Ecosystem:    result.Ecosystem,
			Directory:    payloadDirectory,
			Managers:     result.Managers,
			MatchedFiles: result.MatchedFiles,
			Warnings:     ecosystemWarnings,
		})
	}

	return payload
}
