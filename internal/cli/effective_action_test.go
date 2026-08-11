package cli

import (
	"errors"
	"testing"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/freshness"
)

func TestEffectiveActionPrecedence(t *testing.T) {
	previousLookPath := lookPath
	t.Cleanup(func() { lookPath = previousLookPath })

	node := detection.DetectionResult{
		Ecosystem:    detection.EcosystemNode,
		Managers:     []string{"npm"},
		MatchedFiles: []string{"package-lock.json"},
	}
	update := freshness.EcosystemDecision{Decision: freshness.DecisionUpdate}

	t.Run("unchanged precedes global disablement", func(t *testing.T) {
		action := effectiveActionFor(node, freshness.EcosystemDecision{Decision: freshness.DecisionSkip}, false, true, lookPath)
		if action.Kind != effectiveActionUnchanged {
			t.Fatalf("unexpected action: %q", action.Kind)
		}
	})

	t.Run("global disablement precedes policy and lookup", func(t *testing.T) {
		lookPath = func(string) (string, error) { t.Fatal("lookup must not run"); return "", nil }
		action := effectiveActionFor(detection.DetectionResult{Ecosystem: detection.EcosystemPython, Managers: []string{"uv"}}, update, false, true, lookPath)
		if action.Kind != effectiveActionDisabled {
			t.Fatalf("unexpected action: %q", action.Kind)
		}
	})

	t.Run("policy precedes missing manager", func(t *testing.T) {
		lookPath = func(string) (string, error) { t.Fatal("lookup must not run"); return "", nil }
		action := effectiveActionFor(detection.DetectionResult{Ecosystem: detection.EcosystemPython, Managers: []string{"uv"}}, update, false, false, lookPath)
		if action.Kind != effectiveActionPolicyBlocked {
			t.Fatalf("unexpected action: %q", action.Kind)
		}
	})

	t.Run("missing manager", func(t *testing.T) {
		lookPath = func(string) (string, error) { return "", errors.New("missing") }
		action := effectiveActionFor(node, update, false, false, lookPath)
		if action.Kind != effectiveActionManagerMissing || action.Plan.Manager != "npm" {
			t.Fatalf("unexpected action: %#v", action)
		}
	})

	t.Run("ready", func(t *testing.T) {
		lookPath = func(string) (string, error) { return "/bin/npm", nil }
		action := effectiveActionFor(node, update, false, false, lookPath)
		if action.Kind != effectiveActionReady || action.ManagerPath != "/bin/npm" {
			t.Fatalf("unexpected action: %#v", action)
		}
	})
}
