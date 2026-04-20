package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/freshness"
	"github.com/aaronflorey/pupdate/internal/state"
	"github.com/spf13/cobra"
)

type runExecution struct {
	Results             []detection.DetectionResult
	Store               state.Store
	CurrentState        state.FileState
	DecisionByEcosystem map[string]freshness.EcosystemDecision
}

func executeRun(cmd *cobra.Command, options runOptions) error {
	inHomeDir, err := isHomeDirectory()
	if err != nil {
		return err
	}
	if inHomeDir {
		printStatus(cmd, options.Quiet, "pupdate: skip repo ($HOME)")
		return nil
	}

	execution, err := prepareRunExecution(cmd, options)
	if err != nil {
		return err
	}

	ignored, err := hasPupIgnore(".")
	if err != nil {
		return fmt.Errorf("failed to check .pupignore: %w", err)
	}
	if ignored {
		printStatus(cmd, options.Quiet, "pupdate: skip repo (.pupignore)")
		return nil
	}

	installDisabled := isInstallDisabled()
	if installDisabled {
		printStatus(cmd, options.Quiet, "pupdate: installs disabled via PUPDATE_SKIP_INSTALL")
	}

	outcomes := executeRunResults(cmd, execution.Results, execution.DecisionByEcosystem, options, installDisabled)
	return saveSuccessfulRunOutcomes(execution.Store, execution.CurrentState, outcomes)
}

func prepareRunExecution(cmd *cobra.Command, options runOptions) (runExecution, error) {
	results, err := detectFn(".")
	if err != nil {
		return runExecution{}, fmt.Errorf("detection failed: %w", err)
	}

	store := state.NewStore(".")
	currentState, warnings, err := store.Load()
	if err != nil {
		return runExecution{}, fmt.Errorf("failed to load state: %w", err)
	}
	for _, warning := range warnings {
		printStatus(cmd, options.Quiet, "pupdate: "+warning)
	}

	decisions, err := evaluateFreshnessFn(".", results, currentState)
	if err != nil {
		return runExecution{}, fmt.Errorf("failed to evaluate dependency freshness: %w", err)
	}

	return runExecution{
		Results:             results,
		Store:               store,
		CurrentState:        currentState,
		DecisionByEcosystem: indexDecisionsByEcosystem(decisions),
	}, nil
}

func indexDecisionsByEcosystem(decisions []freshness.EcosystemDecision) map[string]freshness.EcosystemDecision {
	indexed := make(map[string]freshness.EcosystemDecision, len(decisions))
	for _, decision := range decisions {
		indexed[decision.StateKey] = decision
	}
	return indexed
}

func executeRunResults(
	cmd *cobra.Command,
	results []detection.DetectionResult,
	decisionByEcosystem map[string]freshness.EcosystemDecision,
	options runOptions,
	installDisabled bool,
) []ecosystemOutcome {
	outcomes := make([]ecosystemOutcome, 0, len(results))
	for _, result := range results {
		decision, ok := decisionByEcosystem[result.StateKey()]
		if !ok {
			continue
		}

		outcome, ok := executeRunResult(cmd, result, decision, options, installDisabled)
		if ok {
			outcomes = append(outcomes, outcome)
		}
	}
	return outcomes
}

func executeRunResult(
	cmd *cobra.Command,
	result detection.DetectionResult,
	decision freshness.EcosystemDecision,
	options runOptions,
	installDisabled bool,
) (ecosystemOutcome, bool) {
	target := resultTarget(result)
	if decision.Decision != freshness.DecisionUpdate {
		printSkipDecision(cmd, options.Quiet, result, decision, target)
		return ecosystemOutcome{}, false
	}
	if installDisabled {
		return ecosystemOutcome{}, false
	}

	plan, ok, reason := selectManagerPlan(result, options.AllowScripts)
	if !ok {
		if reason != "" && !options.Quiet {
			fmt.Fprintln(cmd.ErrOrStderr(), "pupdate:", reason)
		}
		return ecosystemOutcome{}, false
	}

	if _, err := lookPath(plan.Manager); err != nil {
		if !options.Quiet {
			fmt.Fprintf(cmd.ErrOrStderr(), "pupdate: skip %s (%s not found on PATH)\n", target, plan.Manager)
		}
		return ecosystemOutcome{}, false
	}

	fmt.Fprintln(cmd.ErrOrStderr(), formatRunLine(result, plan))
	err := runInstall(cmd, options.Quiet, filepath.Join(".", result.Directory), plan.Manager, plan.Args...)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "pupdate: error %s install failed: %v\n", plan.Manager, err)
	} else {
		fmt.Fprintln(cmd.ErrOrStderr(), formatDoneLine(result, plan))
	}

	return ecosystemOutcome{
		StateKey:  result.StateKey(),
		Succeeded: err == nil,
		Lockfiles: decision.Lockfiles,
	}, true
}

func printSkipDecision(cmd *cobra.Command, quiet bool, result detection.DetectionResult, decision freshness.EcosystemDecision, target string) {
	if quiet {
		return
	}

	if result.Ecosystem == detection.EcosystemGit && strings.HasPrefix(decision.Reason, "git submodule status failed:") {
		fmt.Fprintf(cmd.ErrOrStderr(), "pupdate: error %s\n", decision.Reason)
		return
	}

	reason := decision.Reason
	if reason == "" {
		reason = "up-to-date"
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "pupdate: skip %s (%s)\n", target, reason)
}

func printStatus(cmd *cobra.Command, quiet bool, line string) {
	if quiet {
		return
	}
	fmt.Fprintln(cmd.ErrOrStderr(), line)
}

func formatRunLine(result detection.DetectionResult, plan managerPlan) string {
	runLine := fmt.Sprintf("pupdate: run %s %s", plan.Manager, strings.Join(plan.Args, " "))
	if result.Directory != "" && result.Directory != "." {
		runLine += fmt.Sprintf(" (in %s)", result.Directory)
	}
	return runLine
}

func formatDoneLine(result detection.DetectionResult, plan managerPlan) string {
	doneLine := fmt.Sprintf("pupdate: done %s", plan.Manager)
	if result.Directory != "" && result.Directory != "." {
		doneLine += fmt.Sprintf(" (in %s)", result.Directory)
	}
	return doneLine
}

func hasPupIgnore(dir string) (bool, error) {
	path := filepath.Join(dir, ".pupignore")
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return !info.IsDir(), nil
}

func resultTarget(result detection.DetectionResult) string {
	if result.Directory == "" || result.Directory == "." {
		return string(result.Ecosystem)
	}
	return fmt.Sprintf("%s:%s", result.Ecosystem, result.Directory)
}

func isInstallDisabled() bool {
	value := strings.TrimSpace(os.Getenv("PUPDATE_SKIP_INSTALL"))
	if value == "" {
		return false
	}
	value = strings.ToLower(value)
	return value == "1" || value == "true" || value == "yes"
}

func isHomeDirectory() (bool, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return false, fmt.Errorf("failed to resolve working directory: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, nil
	}

	return sameDirectory(workingDir, homeDir), nil
}

func sameDirectory(left string, right string) bool {
	return resolveDirectory(left) == resolveDirectory(right)
}

func resolveDirectory(path string) string {
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		path = resolved
	}

	return filepath.Clean(path)
}
