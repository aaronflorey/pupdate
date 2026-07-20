package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/freshness"
	"github.com/aaronflorey/pupdate/internal/state"
	"github.com/spf13/cobra"
)

var userHomeDir = os.UserHomeDir

var currentUserHomeDir = func() (string, error) {
	current, err := user.Current()
	if err != nil {
		return "", err
	}
	return current.HomeDir, nil
}

func executeRun(cmd *cobra.Command, quietFlag bool, allowScriptsFlag bool, dryRunFlag bool) (returnErr error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to resolve working directory: %w", err)
	}
	lock, acquired, err := acquireRepositoryLock(workingDir)
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
		if err := lock.release(); err != nil && returnErr == nil {
			returnErr = err
		}
	}()

	preflight, err := collectPreflight(preflightOptions{})
	if err != nil {
		return err
	}

	options := resolveRunOptions(cmd, preflight.ResolvedConfig, quietFlag, allowScriptsFlag)
	options.DryRun = dryRunFlag

	if skipLine, ok := runSkipStatusLine(preflight.SkipReason); ok {
		printStatus(cmd, options.Quiet, skipLine)
		return nil
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

	outcomes, installErrors := executeRunResults(cmd, preflight.Results, preflight.DecisionByEcosystem, options, installDisabled)
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

func isOutsideConfiguredRootDirectories(config userConfig) (bool, error) {
	if len(config.RootDirectories) == 0 {
		return false, nil
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return false, fmt.Errorf("failed to resolve working directory: %w", err)
	}

	for _, rootDirectory := range config.RootDirectories {
		if isTopLevelDirectoryWithinRoot(workingDir, rootDirectory) {
			return false, nil
		}
	}

	return true, nil
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
	options.SuppressRoutineStatus = options.Quiet && cmd != nil && cmd.Name() == "hook"
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
	results []detection.DetectionResult,
	decisionByEcosystem map[string]freshness.EcosystemDecision,
	options runOptions,
	installDisabled bool,
) ([]ecosystemOutcome, []error) {
	outcomes := make([]ecosystemOutcome, 0, len(results))
	var installErrors []error
	for _, result := range results {
		decision, ok := decisionByEcosystem[result.StateKey()]
		if !ok {
			continue
		}

		outcome, ok, err := executeRunResult(cmd, result, decision, options, installDisabled)
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
	result detection.DetectionResult,
	decision freshness.EcosystemDecision,
	options runOptions,
	installDisabled bool,
) (ecosystemOutcome, bool, error) {
	target := resultTarget(result)
	action := effectiveActionFor(result, decision, options.AllowScripts, installDisabled)
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

	printStatus(cmd, options.SuppressRoutineStatus, formatRunLine(result, plan))
	if options.DryRun {
		printStatus(cmd, options.SuppressRoutineStatus, formatDryRunLine(result, plan))
		return ecosystemOutcome{
			StateKey:  result.StateKey(),
			Succeeded: false,
			Lockfiles: decision.Lockfiles,
		}, false, nil
	}
	err := runInstall(cmd, options.Quiet, filepath.Join(".", result.Directory), plan.Manager, plan.Args...)
	if err == nil {
		printStatus(cmd, options.SuppressRoutineStatus, formatDoneLine(result, plan))
	}

	lockfiles, lockfileMetadata := postInstallLockfileState(result, decision, err == nil)

	return ecosystemOutcome{
		StateKey:         result.StateKey(),
		Succeeded:        err == nil,
		Lockfiles:        lockfiles,
		LockfileMetadata: lockfileMetadata,
	}, true, installTargetError(result, plan, err)
}

func installTargetError(result detection.DetectionResult, plan managerPlan, err error) error {
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
	result detection.DetectionResult,
	decision freshness.EcosystemDecision,
	installSucceeded bool,
) (map[string]string, map[string]state.LockfileMetadata) {
	if !installSucceeded {
		return decision.Lockfiles, decision.LockfileMetadata
	}

	postInstallDecisions, err := evaluateFreshnessFn(".", []detection.DetectionResult{result}, state.FileState{
		Version: state.SchemaVersion,
		Ecosystems: map[string]state.EcosystemState{
			result.StateKey(): {
				Lockfiles:        cloneLockfiles(decision.Lockfiles),
				LockfileMetadata: cloneLockfileMetadata(decision.LockfileMetadata),
			},
		},
	})
	if err != nil || len(postInstallDecisions) != 1 {
		return decision.Lockfiles, decision.LockfileMetadata
	}

	return postInstallDecisions[0].Lockfiles, postInstallDecisions[0].LockfileMetadata
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

func formatRunLine(result detection.DetectionResult, plan managerPlan) string {
	args := make([]string, len(plan.Args))
	for index, arg := range plan.Args {
		args[index] = safeDisplay(arg)
	}
	runLine := fmt.Sprintf("pupdate: run %s %s", safeDisplay(plan.Manager), strings.Join(args, " "))
	if result.Directory != "" && result.Directory != "." {
		runLine += fmt.Sprintf(" (in %s)", safeDisplay(result.Directory))
	}
	return runLine
}

func formatDoneLine(result detection.DetectionResult, plan managerPlan) string {
	doneLine := fmt.Sprintf("pupdate: done %s", safeDisplay(plan.Manager))
	if result.Directory != "" && result.Directory != "." {
		doneLine += fmt.Sprintf(" (in %s)", safeDisplay(result.Directory))
	}
	return doneLine
}

func formatDryRunLine(result detection.DetectionResult, plan managerPlan) string {
	args := make([]string, len(plan.Args))
	for index, arg := range plan.Args {
		args[index] = safeDisplay(arg)
	}
	line := fmt.Sprintf("pupdate: dry-run %s %s", safeDisplay(plan.Manager), strings.Join(args, " "))
	if result.Directory != "" && result.Directory != "." {
		line += fmt.Sprintf(" (in %s)", safeDisplay(result.Directory))
	}
	return line
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

	for _, homeDir := range homeDirectoryCandidates() {
		if sameDirectory(workingDir, homeDir) {
			return true, nil
		}
	}

	return false, nil
}

func homeDirectoryCandidates() []string {
	candidates := make([]string, 0, 2)
	seen := make(map[string]struct{}, 2)

	if homeDir, err := userHomeDir(); err == nil && strings.TrimSpace(homeDir) != "" {
		cleaned := resolveDirectory(homeDir)
		candidates = appendUniqueDirectory(candidates, seen, cleaned)
	}

	if homeDir, err := currentUserHomeDir(); err == nil && strings.TrimSpace(homeDir) != "" {
		cleaned := resolveDirectory(homeDir)
		candidates = appendUniqueDirectory(candidates, seen, cleaned)
	}

	return candidates
}

func appendUniqueDirectory(candidates []string, seen map[string]struct{}, dir string) []string {
	if dir == "" {
		return candidates
	}
	if _, ok := seen[dir]; ok {
		return candidates
	}
	seen[dir] = struct{}{}
	return append(candidates, dir)
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
