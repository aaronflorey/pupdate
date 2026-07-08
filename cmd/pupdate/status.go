package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/freshness"
	"github.com/spf13/cobra"
)

type statusSnapshot struct {
	WorkingDirectory string
	RunStatus        string
	RunReason        string
	RunGuidance      []string
	RunOptions       runOptions
	HookLockPath     string
	HookLockStatus   string
	ConfigPath       string
	ConfigExists     bool
	RawConfig        userConfig
	ResolvedConfig   userConfig
	StatePath        string
	StateExists      bool
	StateWarnings    []string
	Targets          []statusTarget
	DetectedTargets  int
}

type statusTarget struct {
	Name            string
	Directory       string
	MatchedFiles    []string
	Managers        []string
	Warnings        []string
	Freshness       string
	FreshnessReason string
	InstallStatus   string
	InstallReason   string
	InstallCommand  string
	ManagerPath     string
	InstallGuidance []string
}

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Explain what pupdate would do here",
		Long:  "Explain why pupdate would run, skip, or fail in the current directory without changing state or running installs.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			snapshot, err := collectStatusSnapshot()
			if err != nil {
				return err
			}
			return writeStatusSnapshot(cmd.OutOrStdout(), snapshot)
		},
	}

	return cmd
}

func collectStatusSnapshot() (statusSnapshot, error) {
	preflight, err := collectPreflight(preflightOptions{LoadStateOnSkip: true})
	if err != nil {
		return statusSnapshot{}, err
	}

	snapshot := statusSnapshot{
		WorkingDirectory: preflight.WorkingDirectory,
		RunOptions:       resolveRunOptions(nil, preflight.ResolvedConfig, false, false),
		ConfigPath:       preflight.ConfigPath,
		ConfigExists:     preflight.ConfigExists,
		RawConfig:        preflight.RawConfig,
		ResolvedConfig:   preflight.ResolvedConfig,
		StatePath:        preflight.StatePath,
		StateExists:      preflight.StateExists,
		StateWarnings:    preflight.StateWarnings,
	}

	hookLockPath, hookLockStatus, err := currentBackgroundHookStatus(".", backgroundHookNow())
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

	targets := make([]statusTarget, 0, len(preflight.Results))
	for _, result := range preflight.Results {
		decision, ok := preflight.DecisionByEcosystem[result.StateKey()]
		if !ok {
			continue
		}
		targets = append(targets, buildStatusTarget(result, decision, snapshot.RunOptions.AllowScripts))
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

func buildStatusTarget(result detection.DetectionResult, decision freshness.EcosystemDecision, allowScripts bool) statusTarget {
	target := statusTarget{
		Name:            resultTarget(result),
		Directory:       result.Directory,
		MatchedFiles:    append([]string(nil), result.MatchedFiles...),
		Managers:        append([]string(nil), result.Managers...),
		Warnings:        detectionWarnings(result.Warnings),
		Freshness:       string(decision.Decision),
		FreshnessReason: decision.Reason,
	}

	if decision.Decision != freshness.DecisionUpdate {
		target.InstallStatus = "skip"
		target.InstallReason = decision.Reason
		return target
	}

	plan, ok, reason := selectManagerPlan(result, allowScripts)
	if !ok {
		target.InstallStatus = "blocked"
		target.InstallReason = reason
		return target
	}

	target.InstallCommand = strings.TrimPrefix(formatRunLine(result, plan), "pupdate: run ")
	managerPath, err := lookPath(plan.Manager)
	if err != nil {
		target.InstallStatus = "blocked"
		target.InstallReason = fmt.Sprintf("%s not found on PATH", plan.Manager)
		target.InstallGuidance = []string{fmt.Sprintf("install %s or add it to PATH, then rerun pupdate status", plan.Manager)}
		return target
	}

	target.InstallStatus = "ready"
	target.ManagerPath = managerPath
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
		}
	}

	policyGuidance := []string(nil)
	if policyBlockedCount > 0 {
		policyGuidance = []string{"rerun with --allow-scripts, or set allow_scripts: true in config to allow Python installs"}
	}

	if readyCount > 0 && blockedCount > 0 {
		if policyBlockedCount == blockedCount {
			return "mixed", fmt.Sprintf("%s; %d are blocked by the allow-scripts policy", formatEcosystemNeedCount(readyCount+blockedCount), blockedCount), policyGuidance
		}
		return "mixed", fmt.Sprintf("%s; %d are blocked", formatEcosystemNeedCount(readyCount+blockedCount), blockedCount), nil
	}
	if readyCount > 0 {
		return "ready", formatEcosystemNeedCount(readyCount), nil
	}
	if blockedCount > 0 {
		if policyBlockedCount == blockedCount {
			return "blocked", fmt.Sprintf("%s but updates are blocked by the allow-scripts policy", formatEcosystemNeedCount(blockedCount)), policyGuidance
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
	if _, err := fmt.Fprintf(w, "working_directory: %s\n", snapshot.WorkingDirectory); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "run_status: %s\n", snapshot.RunStatus); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "run_reason: %s\n\n", snapshot.RunReason); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "run_guidance: %s\n\n", displayStatusValues(snapshot.RunGuidance, "(none)")); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "config_path: %s\n", snapshot.ConfigPath); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "config_exists: %t\n", snapshot.ConfigExists); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "root_directories: %s\n", displayConfigValues(snapshot.RawConfig.RootDirectories)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "root_directories_resolved: %s\n", displayConfigValues(snapshot.ResolvedConfig.RootDirectories)); err != nil {
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
	if _, err := fmt.Fprintf(w, "effective_allow_scripts: %t\n\n", snapshot.RunOptions.AllowScripts); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "background_hook_lock_path: %s\n", snapshot.HookLockPath); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "background_hook_lock_status: %s\n\n", snapshot.HookLockStatus); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "state_path: %s\n", snapshot.StatePath); err != nil {
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
		if _, err := fmt.Fprintf(w, "\n[%s]\n", target.Name); err != nil {
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
		if _, err := fmt.Fprintf(w, "freshness: %s\n", target.Freshness); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "freshness_reason: %s\n", target.FreshnessReason); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "install_status: %s\n", target.InstallStatus); err != nil {
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
	return strings.Join(values, ", ")
}

func displayStatusValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "(none)"
	}
	return value
}

func displayDirectoryValue(value string) string {
	if value == "" || value == "." {
		return "."
	}
	return value
}
