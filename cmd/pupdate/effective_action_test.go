package main

import (
	"testing"

	"github.com/aaronflorey/pupdate/internal/freshness"
)

func TestClassifyEffectiveActionPrecedence(t *testing.T) {
	plan := managerPlan{Manager: "npm", Args: []string{"ci"}}
	tests := []struct {
		name   string
		inputs effectiveActionInputs
		want   effectiveActionKind
	}{
		{
			name:   "unchanged precedes global disablement",
			inputs: effectiveActionInputs{Decision: freshness.DecisionSkip, InstallDisabled: true, PolicyReason: "blocked", Plan: &plan, ManagerMissing: true},
			want:   effectiveActionUnchanged,
		},
		{
			name:   "global disablement precedes policy",
			inputs: effectiveActionInputs{Decision: freshness.DecisionUpdate, InstallDisabled: true, PolicyReason: "blocked", Plan: &plan, ManagerMissing: true},
			want:   effectiveActionDisabled,
		},
		{
			name:   "policy precedes missing manager",
			inputs: effectiveActionInputs{Decision: freshness.DecisionUpdate, PolicyReason: "blocked", Plan: &plan, ManagerMissing: true},
			want:   effectiveActionPolicyBlocked,
		},
		{
			name:   "missing manager",
			inputs: effectiveActionInputs{Decision: freshness.DecisionUpdate, Plan: &plan, ManagerMissing: true},
			want:   effectiveActionManagerMissing,
		},
		{
			name:   "ready",
			inputs: effectiveActionInputs{Decision: freshness.DecisionUpdate, Plan: &plan, ManagerPath: "/bin/npm"},
			want:   effectiveActionReady,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := classifyEffectiveAction(tt.inputs).Kind; got != tt.want {
				t.Fatalf("unexpected action: got %q want %q", got, tt.want)
			}
		})
	}
}
