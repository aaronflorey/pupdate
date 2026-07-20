package main

import (
	"fmt"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/freshness"
)

type effectiveActionKind string

const (
	effectiveActionUnchanged      effectiveActionKind = "unchanged"
	effectiveActionDisabled       effectiveActionKind = "disabled"
	effectiveActionPolicyBlocked  effectiveActionKind = "policy_blocked"
	effectiveActionManagerMissing effectiveActionKind = "manager_missing"
	effectiveActionReady          effectiveActionKind = "ready"
)

type effectiveActionInputs struct {
	Decision        freshness.Decision
	InstallDisabled bool
	PolicyReason    string
	Plan            *managerPlan
	ManagerPath     string
	ManagerMissing  bool
}

type effectiveAction struct {
	Kind        effectiveActionKind
	Reason      string
	Plan        managerPlan
	ManagerPath string
}

func classifyEffectiveAction(inputs effectiveActionInputs) effectiveAction {
	if inputs.Decision != freshness.DecisionUpdate {
		return effectiveAction{Kind: effectiveActionUnchanged}
	}
	if inputs.InstallDisabled {
		return effectiveAction{Kind: effectiveActionDisabled, Reason: "installs disabled via PUPDATE_SKIP_INSTALL"}
	}
	if inputs.PolicyReason != "" {
		return effectiveAction{Kind: effectiveActionPolicyBlocked, Reason: inputs.PolicyReason}
	}
	if inputs.ManagerMissing && inputs.Plan != nil {
		return effectiveAction{
			Kind:   effectiveActionManagerMissing,
			Reason: fmt.Sprintf("%s not found on PATH", inputs.Plan.Manager),
			Plan:   *inputs.Plan,
		}
	}
	if inputs.Plan != nil {
		return effectiveAction{
			Kind:        effectiveActionReady,
			Plan:        *inputs.Plan,
			ManagerPath: inputs.ManagerPath,
		}
	}
	return effectiveAction{Kind: effectiveActionPolicyBlocked, Reason: "no safe install plan is available"}
}

func effectiveActionFor(result detection.DetectionResult, decision freshness.EcosystemDecision, allowScripts bool, installDisabled bool) effectiveAction {
	inputs := effectiveActionInputs{Decision: decision.Decision, InstallDisabled: installDisabled}
	action := classifyEffectiveAction(inputs)
	if action.Kind == effectiveActionUnchanged || action.Kind == effectiveActionDisabled {
		return action
	}

	plan, ok, reason := selectManagerPlan(result, allowScripts)
	if !ok {
		inputs.PolicyReason = reason
		return classifyEffectiveAction(inputs)
	}
	inputs.Plan = &plan
	managerPath, err := lookPath(plan.Manager)
	if err != nil {
		inputs.ManagerMissing = true
		return classifyEffectiveAction(inputs)
	}
	inputs.ManagerPath = managerPath
	return classifyEffectiveAction(inputs)
}
