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

func executeRun(cmd *cobra.Command, quietFlag bool, allowScriptsFlag bool) error {
	preflight, err := collectPreflight(preflightOptions{})
	if err != nil {
		return err
	}

	options := resolveRunOptions(cmd, preflight.ResolvedConfig, quietFlag, allowScriptsFlag)

	if skipLine, ok := runSkipStatusLine(preflight.SkipReason); ok {
		printStatus(cmd, options.Quiet, skipLine)
		return nil
	}

	for _, warning := range preflight.StateWarnings {
		printStatus(cmd, options.Quiet, "pupdate: "+warning)
	}

	installDisabled := isInstallDisabled()
	if installDisabled {
		printStatus(cmd, options.Quiet, "pupdate: installs disabled via PUPDATE_SKIP_INSTALL")
	}

	outcomes := executeRunResults(cmd, preflight.Results, preflight.DecisionByEcosystem, options, installDisabled)
	return saveRunOutcomes(preflight.Store, preflight.CurrentState, preflight.Results, outcomes)
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
		refreshMetadata := shouldRefreshMetadata(result, decision)
		return ecosystemOutcome{
			StateKey:         result.StateKey(),
			RefreshMetadata:  refreshMetadata,
			Lockfiles:        decision.Lockfiles,
			LockfileMetadata: decision.LockfileMetadata,
		}, refreshMetadata
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

	lockfiles, lockfileMetadata := postInstallLockfileState(result, decision, err == nil)

	return ecosystemOutcome{
		StateKey:         result.StateKey(),
		Succeeded:        err == nil,
		Lockfiles:        lockfiles,
		LockfileMetadata: lockfileMetadata,
	}, true
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
