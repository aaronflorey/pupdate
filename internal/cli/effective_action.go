package cli

import (
	"fmt"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/freshness"
	"github.com/aaronflorey/pupdate/internal/install"
)

type effectiveActionKind string

const (
	effectiveActionUnchanged      effectiveActionKind = "unchanged"
	effectiveActionDisabled       effectiveActionKind = "disabled"
	effectiveActionPolicyBlocked  effectiveActionKind = "policy_blocked"
	effectiveActionManagerMissing effectiveActionKind = "manager_missing"
	effectiveActionReady          effectiveActionKind = "ready"
)

type effectiveAction struct {
	Kind        effectiveActionKind
	Reason      string
	Plan        install.Plan
	ManagerPath string
}

func effectiveActionFor(
	result detection.DetectionResult,
	decision freshness.EcosystemDecision,
	allowScripts bool,
	installDisabled bool,
	lookPath func(string) (string, error),
) effectiveAction {
	if decision.Decision != freshness.DecisionUpdate {
		return effectiveAction{Kind: effectiveActionUnchanged}
	}
	if installDisabled {
		return effectiveAction{Kind: effectiveActionDisabled, Reason: "installs disabled via PUPDATE_SKIP_INSTALL"}
	}

	plan, ok, reason := install.SelectPlan(result, allowScripts)
	if !ok {
		return effectiveAction{Kind: effectiveActionPolicyBlocked, Reason: reason}
	}
	managerPath, err := lookPath(plan.Manager)
	if err != nil {
		return effectiveAction{
			Kind:   effectiveActionManagerMissing,
			Reason: fmt.Sprintf("%s not found on PATH", plan.Manager),
			Plan:   plan,
		}
	}
	return effectiveAction{Kind: effectiveActionReady, Plan: plan, ManagerPath: managerPath}
}
