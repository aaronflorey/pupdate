package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/freshness"
	"github.com/spf13/cobra"
)

type statusSnapshot struct {
	preflightResult
	RunStatus       string
	RunReason       string
	RunGuidance     []string
	RunOptions      runOptions
	HookLockPath    string
	HookLockStatus  string
	Targets         []statusTarget
	DetectedTargets int
}

type statusTarget struct {
	Name            string
	Directory       string
	MatchedFiles    []string
	Managers        []string
	Warnings        []string
	Freshness       string
	FreshnessReason string
	LastSuccessAt   string
	InstallStatus   string
	InstallReason   string
	InstallCommand  string
	ManagerPath     string
	InstallGuidance []string
}

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status",
		Short:   "Explain what pupdate would do here",
		Long:    "Explain why pupdate would run, skip, or fail in the current directory without changing state or running installs.",
		Example: `pupdate status`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			snapshot, err := collectStatusSnapshot(cmd.Context(), dependenciesFor(cmd))
			if err != nil {
				return err
			}
			return writeStatusSnapshot(cmd.OutOrStdout(), snapshot)
		},
	}

	return cmd
}

func collectStatusSnapshot(ctx context.Context, deps dependencies) (statusSnapshot, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return statusSnapshot{}, fmt.Errorf("failed to resolve working directory: %w", err)
	}
	preflight, err := collectPreflight(ctx, workingDirectory, preflightOptions{LoadStateOnSkip: true}, deps)
	if err != nil {
		return statusSnapshot{}, err
	}

	snapshot := statusSnapshot{
		preflightResult: preflight,
		RunOptions:      resolveRunOptions(nil, preflight.ResolvedConfig, false, false),
	}

	hookLockPath, hookLockStatus, err := currentBackgroundHookStatus(".")
	if err != nil {
		return statusSnapshot{}, err
	}
	snapshot.HookLockPath = hookLockPath
	snapshot.HookLockStatus = hookLockStatus
	snapshot.RunGuidance = statusRunGuidance(preflight.SkipReason, hookLockStatus)

	if preflight.SkipReason != preflightSkipNone {
		snapshot.RunStatus = "skip"
		snapshot.RunReason = statusSkipReason(preflight.SkipReason)
		return snapshot, nil
	}

	installDisabled := isInstallDisabled()
	targets := make([]statusTarget, 0, len(preflight.Results))
	for _, result := range preflight.Results {
		decision, ok := preflight.DecisionByEcosystem[result.StateKey()]
		if !ok {
			continue
		}
		lastSuccessAt := preflight.CurrentState.Ecosystems[result.StateKey()].LastSuccessAt
		targets = append(targets, buildStatusTarget(result, decision, snapshot.RunOptions.AllowScripts, installDisabled, lastSuccessAt, deps.LookPath))
	}

	snapshot.Targets = targets
	snapshot.DetectedTargets = len(targets)
	targetRunGuidance := []string(nil)
	snapshot.RunStatus, snapshot.RunReason, targetRunGuidance = summarizeStatusTargets(targets)
	snapshot.RunGuidance = append(snapshot.RunGuidance, targetRunGuidance...)
	return snapshot, nil
}

func statusRunGuidance(reason preflightSkipReason, hookLockStatus string) []string {
	guidance := make([]string, 0, 2)

	switch reason {
	case preflightSkipHomeDirectory:
		guidance = append(guidance, "run pupdate from a project directory instead of $HOME")
	case preflightSkipOutsideRoots:
		guidance = append(guidance, "move this project under root_directories, or update root_directories to include it")
	case preflightSkipPupIgnore:
		guidance = append(guidance, "remove .pupignore if you want pupdate to manage this repo")
	}

	if hookLockStatus == "active" {
		guidance = append(guidance, "wait for the current background hook run to finish before expecting another async hook run")
	}

	return guidance
}

func statusSkipReason(reason preflightSkipReason) string {
	switch reason {
	case preflightSkipHomeDirectory:
		return "current directory is $HOME"
	case preflightSkipOutsideRoots:
		return "current directory is outside configured root_directories"
	case preflightSkipPupIgnore:
		return "repo marked with .pupignore"
	default:
		return ""
	}
}

func buildStatusTarget(
	result detection.DetectionResult,
	decision freshness.EcosystemDecision,
	allowScripts bool,
	installDisabled bool,
	lastSuccessAt string,
	lookPath func(string) (string, error),
) statusTarget {
	target := statusTarget{
		Name:            resultTarget(result),
		Directory:       result.Directory,
		MatchedFiles:    append([]string(nil), result.MatchedFiles...),
		Managers:        append([]string(nil), result.Managers...),
		Warnings:        detectionWarnings(result.Warnings),
		Freshness:       string(decision.Decision),
		FreshnessReason: decision.Reason,
		LastSuccessAt:   lastSuccessAt,
	}
	action := effectiveActionFor(result, decision, allowScripts, installDisabled, lookPath)
	if action.Kind == effectiveActionUnchanged {
		target.InstallStatus = "skip"
		target.InstallReason = decision.Reason
		return target
	}
	if action.Kind == effectiveActionDisabled {
		target.InstallStatus = "disabled"
		target.InstallReason = action.Reason
		return target
	}
	if action.Plan.Manager != "" {
		target.InstallCommand = formatInstallCommand(action.Plan)
	}
	if action.Kind == effectiveActionPolicyBlocked {
		target.InstallStatus = "blocked"
		target.InstallReason = action.Reason
		return target
	}
	if action.Kind == effectiveActionManagerMissing {
		target.InstallStatus = "blocked"
		target.InstallReason = action.Reason
		target.InstallGuidance = []string{fmt.Sprintf("install %s or add it to PATH, then rerun pupdate status", action.Plan.Manager)}
		return target
	}
	target.InstallStatus = "ready"
	target.ManagerPath = action.ManagerPath
	return target
}

func detectionWarnings(warnings []detection.Warning) []string {
	if len(warnings) == 0 {
		return nil
	}

	messages := make([]string, 0, len(warnings))
	for _, warning := range warnings {
		messages = append(messages, warning.Message)
	}
	return messages
}

func summarizeStatusTargets(targets []statusTarget) (string, string, []string) {
	if len(targets) == 0 {
		return "idle", "no supported ecosystems detected", nil
	}

	readyCount := 0
	blockedCount := 0
	disabledCount := 0
	policyBlockedCount := 0
	for _, target := range targets {
		switch target.InstallStatus {
		case "ready":
			readyCount++
		case "blocked":
			blockedCount++
			if strings.Contains(target.InstallReason, "--allow-scripts") {
				policyBlockedCount++
			}
		case "disabled":
			disabledCount++
		}
	}

	policyGuidance := []string(nil)
	if policyBlockedCount > 0 {
		policyGuidance = []string{"rerun with --allow-scripts, or set allow_scripts: true in config to allow Python installs"}
	}
	if disabledCount > 0 {
		return "disabled", fmt.Sprintf("%s but installs are disabled via PUPDATE_SKIP_INSTALL", formatEcosystemNeedCount(disabledCount)), []string{"unset PUPDATE_SKIP_INSTALL to allow installs"}
	}

	if readyCount > 0 && blockedCount > 0 {
		if policyBlockedCount == blockedCount {
			return "mixed", fmt.Sprintf("%s; %d are blocked by the allow-scripts policy", formatEcosystemNeedCount(readyCount+blockedCount), blockedCount), policyGuidance
		}
		return "mixed", fmt.Sprintf("%s; %d are blocked", formatEcosystemNeedCount(readyCount+blockedCount), blockedCount), policyGuidance
	}
	if readyCount > 0 {
		return "ready", formatEcosystemNeedCount(readyCount), nil
	}
	if blockedCount > 0 {
		if policyBlockedCount == blockedCount {
			return "blocked", fmt.Sprintf("%s but updates are blocked by the allow-scripts policy", formatEcosystemNeedCount(blockedCount)), policyGuidance
		}
		if policyBlockedCount > 0 {
			return "blocked", fmt.Sprintf("%s but updates are blocked", formatEcosystemNeedCount(blockedCount)), policyGuidance
		}
		return "blocked", fmt.Sprintf("%s but required managers are unavailable", formatEcosystemNeedCount(blockedCount)), nil
	}
	return "idle", "all detected ecosystems are already up to date", nil
}

func formatEcosystemNeedCount(count int) string {
	if count == 1 {
		return "1 ecosystem needs updates"
	}
	return fmt.Sprintf("%d ecosystems need updates", count)
}

func writeStatusSnapshot(w io.Writer, snapshot statusSnapshot) error {
	if _, err := fmt.Fprintf(w, "working_directory: %s\n", safeDisplay(snapshot.WorkingDirectory)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "run_status: %s\n", safeDisplay(snapshot.RunStatus)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "run_reason: %s\n\n", safeDisplay(snapshot.RunReason)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "run_guidance: %s\n\n", displayStatusValues(snapshot.RunGuidance, "(none)")); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "config_path: %s\n", safeDisplay(snapshot.ConfigPath)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "config_exists: %t\n", snapshot.ConfigExists); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "root_directories: %s\n", displayConfigValues(snapshot.RawConfig.RootDirectories)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "root_directories_resolved: %s\n", displayConfigValues(snapshot.ResolvedUserConfig.RootDirectories)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "workspace_globs: %s\n", displayConfigValues(snapshot.RawConfig.WorkspaceGlobs)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "workspace_globs_resolved: %s\n", displayConfigValues(snapshot.ResolvedUserConfig.WorkspaceGlobs)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "folder_blacklist: %s\n", displayConfigValues(snapshot.RawConfig.FolderBlacklist)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "folder_blacklist_resolved: %s\n", displayConfigValues(snapshot.ResolvedUserConfig.FolderBlacklist)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "quiet: %s\n", displayOptionalBool(snapshot.RawConfig.Quiet)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "allow_scripts: %s\n", displayOptionalBool(snapshot.RawConfig.AllowScripts)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "effective_quiet: %t\n", snapshot.RunOptions.Quiet); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "effective_allow_scripts: %t\n", snapshot.RunOptions.AllowScripts); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "project_config_path: %s\n", safeDisplay(snapshot.ProjectConfigPath)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "project_config_exists: %t\n", snapshot.ProjectConfigExists); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "project_workspace_globs: %s\n", displayConfigValues(snapshot.RawProjectConfig.WorkspaceGlobs)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "project_workspace_globs_resolved: %s\n", displayConfigValues(snapshot.ResolvedProjectConfig.WorkspaceGlobs)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "project_folder_blacklist: %s\n", displayConfigValues(snapshot.RawProjectConfig.FolderBlacklist)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "project_folder_blacklist_resolved: %s\n", displayConfigValues(snapshot.ResolvedProjectConfig.FolderBlacklist)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "effective_workspace_globs: %s\n", displayConfigValues(snapshot.ResolvedConfig.WorkspaceGlobs)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "effective_folder_blacklist: %s\n\n", displayConfigValues(snapshot.ResolvedConfig.FolderBlacklist)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "background_hook_lock_path: %s\n", safeDisplay(snapshot.HookLockPath)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "background_hook_lock_status: %s\n\n", safeDisplay(snapshot.HookLockStatus)); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "state_path: %s\n", safeDisplay(snapshot.StatePath)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "state_exists: %t\n", snapshot.StateExists); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "state_warnings: %s\n\n", displayStatusValues(snapshot.StateWarnings, "(none)")); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "detected_targets: %d\n", snapshot.DetectedTargets); err != nil {
		return err
	}
	for _, target := range snapshot.Targets {
		if _, err := fmt.Fprintf(w, "\n[%s]\n", safeDisplay(target.Name)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "directory: %s\n", displayDirectoryValue(target.Directory)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "matched_files: %s\n", displayStatusValues(target.MatchedFiles, "(none)")); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "managers: %s\n", displayStatusValues(target.Managers, "(none)")); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "warnings: %s\n", displayStatusValues(target.Warnings, "(none)")); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "freshness: %s\n", safeDisplay(target.Freshness)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "freshness_reason: %s\n", safeDisplay(target.FreshnessReason)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "last_success_at: %s\n", displayStatusValue(target.LastSuccessAt)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "install_status: %s\n", safeDisplay(target.InstallStatus)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "install_reason: %s\n", displayStatusValue(target.InstallReason)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "install_command: %s\n", displayStatusValue(target.InstallCommand)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "manager_path: %s\n", displayStatusValue(target.ManagerPath)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "install_guidance: %s\n", displayStatusValues(target.InstallGuidance, "(none)")); err != nil {
			return err
		}
	}

	return nil
}

func displayStatusValues(values []string, empty string) string {
	if len(values) == 0 {
		return empty
	}
	displayed := make([]string, len(values))
	for index, value := range values {
		displayed[index] = safeDisplay(value)
	}
	return strings.Join(displayed, ", ")
}

func displayStatusValue(value string) string {
	if value == "" {
		return "(none)"
	}
	return safeDisplay(value)
}

func displayDirectoryValue(value string) string {
	if value == "" || value == "." {
		return "."
	}
	return safeDisplay(value)
}
