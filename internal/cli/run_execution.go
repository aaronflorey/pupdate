package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/freshness"
	"github.com/aaronflorey/pupdate/internal/install"
	"github.com/aaronflorey/pupdate/internal/repolock"
	"github.com/aaronflorey/pupdate/internal/state"
	"github.com/aaronflorey/pupdate/internal/userenv"
	"github.com/spf13/cobra"
)

func executeRun(cmd *cobra.Command, quietFlag bool, allowScriptsFlag bool, dryRunFlag bool) (returnErr error) {
	deps := dependenciesFor(cmd)
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to resolve working directory: %w", err)
	}
	preflight, err := collectDetectionPreflight(workingDir, deps)
	if err != nil {
		return err
	}
	options := resolveRunOptions(cmd, preflight.ResolvedConfig, quietFlag, allowScriptsFlag)
	options.DryRun = dryRunFlag
	if skipLine, ok := runSkipStatusLine(preflight.SkipReason); ok {
		printStatus(cmd, options.Quiet, skipLine)
		return nil
	}
	if len(preflight.Results) == 0 {
		return nil
	}

	lock, acquired, err := repolock.Acquire(workingDir)
	if err != nil {
		return err
	}
	if !acquired {
		if cmd != nil && cmd.Name() == "hook" {
			printStatus(cmd, quietFlag, "pupdate: skip repo (another run is active)")
			return nil
		}
		return fmt.Errorf("another pupdate operation is already active in %s", workingDir)
	}
	defer func() {
		if err := lock.Release(); err != nil && returnErr == nil {
			returnErr = err
		}
	}()

	preflight, err = collectStatePreflight(cmd.Context(), preflight, preflightOptions{}, deps)
	if err != nil {
		return err
	}

	for _, warning := range preflight.StateWarnings {
		printStatus(cmd, options.Quiet, "pupdate: "+safeDisplay(warning))
	}

	installDisabled := isInstallDisabled()
	if installDisabled {
		printStatus(cmd, options.Quiet, "pupdate: installs disabled via PUPDATE_SKIP_INSTALL")
	}
	if options.DryRun {
		printStatus(cmd, options.Quiet, "pupdate: --dry-run (no installs, no state changes)")
	}

	outcomes, installErrors := executeRunResults(cmd, workingDir, preflight.Results, preflight.DecisionByEcosystem, options, installDisabled, deps)
	if options.DryRun || installDisabled {
		return nil
	}
	saveErr := saveRunOutcomes(preflight.Store, preflight.CurrentState, preflight.Results, outcomes)
	if len(installErrors) > 0 {
		installErr := aggregateInstallErrors(installErrors)
		if saveErr != nil {
			return fmt.Errorf("%v; %w", installErr, saveErr)
		}
		return installErr
	}
	return saveErr
}

func runSkipStatusLine(reason preflightSkipReason) (string, bool) {
	switch reason {
	case preflightSkipHomeDirectory:
		return "pupdate: skip repo ($HOME)", true
	case preflightSkipOutsideRoots:
		return "pupdate: skip repo (outside configured root_directories)", true
	case preflightSkipPupIgnore:
		return "pupdate: skip repo (.pupignore)", true
	default:
		return "", false
	}
}

func isOutsideConfiguredRootDirectories(workingDir string, config userConfig) bool {
	if len(config.RootDirectories) == 0 {
		return false
	}

	for _, rootDirectory := range config.RootDirectories {
		if userenv.IsTopLevelDirectoryWithinRoot(workingDir, rootDirectory) {
			return false
		}
	}
	return true
}

func resolveRunOptions(cmd *cobra.Command, config userConfig, quietFlag bool, allowScriptsFlag bool) runOptions {
	options := runOptions{}
	if config.Quiet != nil {
		options.Quiet = *config.Quiet
	}
	if config.AllowScripts != nil {
		options.AllowScripts = *config.AllowScripts
	}
	if cmd != nil && cmd.Flags().Changed("quiet") {
		options.Quiet = quietFlag
	}
	if cmd != nil && cmd.Flags().Changed("allow-scripts") {
		options.AllowScripts = allowScriptsFlag
	}
	return options
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
	workingDir string,
	results []detection.DetectionResult,
	decisionByEcosystem map[string]freshness.EcosystemDecision,
	options runOptions,
	installDisabled bool,
	deps dependencies,
) ([]ecosystemOutcome, []error) {
	outcomes := make([]ecosystemOutcome, 0, len(results))
	var installErrors []error
	for _, result := range results {
		if err := cmd.Context().Err(); err != nil {
			installErrors = append(installErrors, err)
			break
		}
		decision, ok := decisionByEcosystem[result.StateKey()]
		if !ok {
			continue
		}

		outcome, ok, err := executeRunResult(cmd, workingDir, result, decision, options, installDisabled, deps)
		if ok {
			outcomes = append(outcomes, outcome)
		}
		if err != nil {
			installErrors = append(installErrors, err)
		}
	}
	return outcomes, installErrors
}

func executeRunResult(
	cmd *cobra.Command,
	workingDir string,
	result detection.DetectionResult,
	decision freshness.EcosystemDecision,
	options runOptions,
	installDisabled bool,
	deps dependencies,
) (ecosystemOutcome, bool, error) {
	target := resultTarget(result)
	action := effectiveActionFor(result, decision, options.AllowScripts, installDisabled, deps.LookPath)
	if action.Kind == effectiveActionUnchanged {
		printSkipDecision(cmd, options.Quiet, result, decision, target)
		refreshMetadata := shouldRefreshMetadata(result, decision)
		return ecosystemOutcome{
			StateKey:         result.StateKey(),
			RefreshMetadata:  refreshMetadata,
			Lockfiles:        decision.Lockfiles,
			LockfileMetadata: decision.LockfileMetadata,
		}, refreshMetadata, nil
	}
	if action.Kind == effectiveActionDisabled {
		return ecosystemOutcome{}, false, nil
	}
	if action.Kind == effectiveActionPolicyBlocked || action.Kind == effectiveActionManagerMissing {
		if action.Reason != "" && !options.Quiet {
			fmt.Fprintf(cmd.ErrOrStderr(), "pupdate: skip %s (%s)\n", safeDisplay(target), safeDisplay(action.Reason))
		}
		return ecosystemOutcome{}, false, nil
	}
	plan := action.Plan

	printStatus(cmd, options.Quiet, formatRunLine(result, plan))
	if options.DryRun {
		printStatus(cmd, options.Quiet, formatDryRunLine(result, plan))
		return ecosystemOutcome{
			StateKey:  result.StateKey(),
			Succeeded: false,
			Lockfiles: decision.Lockfiles,
		}, false, nil
	}
	err := runInstall(cmd, options.Quiet, filepath.Join(workingDir, result.Directory), deps, plan.Manager, plan.Args...)
	if err == nil {
		printStatus(cmd, options.Quiet, formatDoneLine(result, plan))
	}

	lockfiles, lockfileMetadata, snapshotErr := postInstallLockfileState(cmd.Context(), workingDir, result, decision, err == nil, deps.SnapshotFreshness)
	if snapshotErr != nil {
		return ecosystemOutcome{}, false, snapshotErr
	}

	return ecosystemOutcome{
		StateKey:         result.StateKey(),
		Succeeded:        err == nil,
		Lockfiles:        lockfiles,
		LockfileMetadata: lockfileMetadata,
	}, true, installTargetError(result, plan, err)
}

func installTargetError(result detection.DetectionResult, plan install.Plan, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s install failed for %s: %w", plan.Manager, resultTarget(result), err)
}

func aggregateInstallErrors(failures []error) error {
	parts := make([]string, 0, len(failures))
	for _, failure := range failures {
		parts = append(parts, failure.Error())
	}
	return fmt.Errorf("dependency installs failed: %s", strings.Join(parts, "; "))
}

func postInstallLockfileState(
	ctx context.Context,
	workingDir string,
	result detection.DetectionResult,
	decision freshness.EcosystemDecision,
	installSucceeded bool,
	snapshotFreshness func(context.Context, string, []string, state.EcosystemState) (freshness.MatchedFilesSnapshot, error),
) (map[string]string, map[string]state.LockfileMetadata, error) {
	if !installSucceeded {
		return decision.Lockfiles, decision.LockfileMetadata, nil
	}

	snapshot, err := snapshotFreshness(ctx, workingDir, result.MatchedFiles, state.EcosystemState{
		Lockfiles:        decision.Lockfiles,
		LockfileMetadata: decision.LockfileMetadata,
	})
	if err != nil {
		if cancellationErr := ctx.Err(); cancellationErr != nil {
			return nil, nil, cancellationErr
		}
		return decision.Lockfiles, decision.LockfileMetadata, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	return snapshot.Lockfiles, snapshot.LockfileMetadata, nil
}

func shouldRefreshMetadata(result detection.DetectionResult, decision freshness.EcosystemDecision) bool {
	if decision.Decision != freshness.DecisionSkip {
		return false
	}
	if result.Ecosystem == detection.EcosystemGit {
		return false
	}
	return len(decision.Lockfiles) > 0 && len(decision.LockfileMetadata) > 0
}

func printSkipDecision(cmd *cobra.Command, quiet bool, result detection.DetectionResult, decision freshness.EcosystemDecision, target string) {
	if quiet {
		return
	}

	if result.Ecosystem == detection.EcosystemGit && strings.HasPrefix(decision.Reason, "git submodule status failed:") {
		fmt.Fprintf(cmd.ErrOrStderr(), "pupdate: error %s\n", safeDisplay(decision.Reason))
		return
	}

	reason := decision.Reason
	if reason == "" {
		reason = "up-to-date"
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "pupdate: skip %s (%s)\n", safeDisplay(target), safeDisplay(reason))
}

func printStatus(cmd *cobra.Command, quiet bool, line string) {
	if quiet {
		return
	}
	fmt.Fprintln(cmd.ErrOrStderr(), line)
}

func formatRunLine(result detection.DetectionResult, plan install.Plan) string {
	return "pupdate: run " + formatInstallCommand(plan) + formatResultDirectory(result)
}

func formatDoneLine(result detection.DetectionResult, plan install.Plan) string {
	return "pupdate: done " + safeDisplay(plan.Manager) + formatResultDirectory(result)
}

func formatDryRunLine(result detection.DetectionResult, plan install.Plan) string {
	return "pupdate: dry-run " + formatInstallCommand(plan) + formatResultDirectory(result)
}

func formatInstallCommand(plan install.Plan) string {
	args := make([]string, len(plan.Args))
	for index, arg := range plan.Args {
		args[index] = safeDisplay(arg)
	}
	command := safeDisplay(plan.Manager)
	if len(args) > 0 {
		command += " " + strings.Join(args, " ")
	}
	return command
}

func formatResultDirectory(result detection.DetectionResult) string {
	if result.IsRoot() {
		return ""
	}
	return fmt.Sprintf(" (in %s)", safeDisplay(result.Directory))
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
	if result.IsRoot() {
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

func isHomeDirectory(workingDir string, environment userenv.Environment) bool {
	for _, homeDir := range environment.HomeDirectories() {
		if userenv.SameDirectory(workingDir, homeDir) {
			return true
		}
	}
	return false
}
