package freshness

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/state"
)

type Decision string

const (
	DecisionUpdate Decision = "update"
	DecisionSkip   Decision = "skip"
)

type EcosystemDecision struct {
	Ecosystem string
	Decision  Decision
	Reason    string
	MaxMTime  time.Time
}

func Evaluate(dir string, detections []detection.DetectionResult, current state.FileState) ([]EcosystemDecision, error) {
	decisions := make([]EcosystemDecision, 0, len(detections))

	for _, result := range detections {
		maxMTime, err := maxMatchedMTime(dir, result.MatchedFiles)
		if err != nil {
			return nil, err
		}

		ecosystem := string(result.Ecosystem)
		lastRaw, hasState := "", false
		if ecosystemState, ok := current.Ecosystems[ecosystem]; ok {
			lastRaw = ecosystemState.LastSuccessAt
			hasState = true
		}

		decision := EcosystemDecision{
			Ecosystem: ecosystem,
			Decision:  DecisionUpdate,
			Reason:    "missing prior successful run timestamp",
			MaxMTime:  maxMTime.UTC(),
		}

		if !hasState || lastRaw == "" {
			decisions = append(decisions, decision)
			continue
		}

		lastSuccess, err := state.ParseRFC3339UTC(lastRaw)
		if err != nil {
			decision.Reason = "invalid prior successful run timestamp"
			decisions = append(decisions, decision)
			continue
		}

		lastSuccess = lastSuccess.UTC()
		if !lastSuccess.Before(maxMTime.UTC()) {
			decision.Decision = DecisionSkip
			decision.Reason = "no dependency changes since last successful run"
		} else {
			decision.Reason = "dependency files changed since last successful run"
		}

		decisions = append(decisions, decision)
	}

	return decisions, nil
}

func maxMatchedMTime(dir string, matchedFiles []string) (time.Time, error) {
	var max time.Time
	for _, matchedFile := range matchedFiles {
		fullPath := filepath.Join(dir, matchedFile)
		info, err := os.Stat(fullPath)
		if err != nil {
			return time.Time{}, fmt.Errorf("stat matched file %q: %w", matchedFile, err)
		}

		modTime := info.ModTime().UTC()
		if modTime.After(max) {
			max = modTime
		}
	}

	return max.UTC(), nil
}
