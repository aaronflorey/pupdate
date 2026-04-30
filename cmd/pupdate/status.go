package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/freshness"
	"github.com/aaronflorey/pupdate/internal/state"
	"github.com/spf13/cobra"
)

type statusSnapshot struct {
	WorkingDirectory string
	RunStatus        string
	RunReason        string
	RunOptions       runOptions
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
	workingDirectory, err := os.Getwd()
	if err != nil {
		return statusSnapshot{}, fmt.Errorf("failed to resolve working directory: %w", err)
	}

	configPath, err := resolveUserConfigPath()
	if err != nil {
		return statusSnapshot{}, err
	}

	rawConfig, resolvedConfig, configExists, err := loadStatusConfig(configPath)
	if err != nil {
		return statusSnapshot{}, err
	}

	statePath := filepath.Join(workingDirectory, state.FileName)
	stateExists, err := pathExists(statePath)
	if err != nil {
		return statusSnapshot{}, fmt.Errorf("failed to stat %s: %w", statePath, err)
	}

	store := state.NewStore(".")
	currentState, warnings, err := store.Load()
	if err != nil {
		return statusSnapshot{}, fmt.Errorf("failed to load state: %w", err)
	}

	snapshot := statusSnapshot{
		WorkingDirectory: workingDirectory,
		RunOptions:       resolveRunOptions(nil, resolvedConfig, false, false),
		ConfigPath:       configPath,
		ConfigExists:     configExists,
		RawConfig:        rawConfig,
		ResolvedConfig:   resolvedConfig,
		StatePath:        statePath,
		StateExists:      stateExists,
		StateWarnings:    warnings,
	}

	repoStatus, repoReason, err := statusPrecheck(resolvedConfig)
	if err != nil {
		return statusSnapshot{}, err
	}
	if repoStatus != "" {
		snapshot.RunStatus = repoStatus
		snapshot.RunReason = repoReason
		return snapshot, nil
	}

	results, err := detectFn(".")
	if err != nil {
		return statusSnapshot{}, fmt.Errorf("detection failed: %w", err)
	}

	decisions, err := evaluateFreshnessFn(".", results, currentState)
	if err != nil {
		return statusSnapshot{}, fmt.Errorf("failed to evaluate dependency freshness: %w", err)
	}

	indexedDecisions := indexDecisionsByEcosystem(decisions)
	targets := make([]statusTarget, 0, len(results))
	for _, result := range results {
		decision, ok := indexedDecisions[result.StateKey()]
		if !ok {
			continue
		}
		targets = append(targets, buildStatusTarget(result, decision, snapshot.RunOptions.AllowScripts))
	}

	snapshot.Targets = targets
	snapshot.DetectedTargets = len(targets)
	snapshot.RunStatus, snapshot.RunReason = summarizeStatusTargets(targets)
	return snapshot, nil
}

func loadStatusConfig(configPath string) (userConfig, userConfig, bool, error) {
	configExists, err := pathExists(configPath)
	if err != nil {
		return userConfig{}, userConfig{}, false, fmt.Errorf("failed to stat %s: %w", configPath, err)
	}

	rawConfig, err := readUserConfig(configPath)
	if err != nil {
		return userConfig{}, userConfig{}, false, err
	}

	resolvedConfig, err := resolveUserConfig(rawConfig)
	if err != nil {
		return userConfig{}, userConfig{}, false, err
	}

	return rawConfig, resolvedConfig, configExists, nil
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func statusPrecheck(config userConfig) (string, string, error) {
	inHomeDir, err := isHomeDirectory()
	if err != nil {
		return "", "", err
	}
	if inHomeDir {
		return "skip", "current directory is $HOME", nil
	}

	outsideConfiguredRoots, err := isOutsideConfiguredRootDirectories(config)
	if err != nil {
		return "", "", err
	}
	if outsideConfiguredRoots {
		return "skip", "current directory is outside configured root_directories", nil
	}

	ignored, err := hasPupIgnore(".")
	if err != nil {
		return "", "", fmt.Errorf("failed to check .pupignore: %w", err)
	}
	if ignored {
		return "skip", "repo marked with .pupignore", nil
	}

	return "", "", nil
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

func summarizeStatusTargets(targets []statusTarget) (string, string) {
	if len(targets) == 0 {
		return "idle", "no supported ecosystems detected"
	}

	readyCount := 0
	blockedCount := 0
	for _, target := range targets {
		switch target.InstallStatus {
		case "ready":
			readyCount++
		case "blocked":
			blockedCount++
		}
	}

	if readyCount > 0 && blockedCount > 0 {
		return "mixed", fmt.Sprintf("%s; %d are blocked", formatEcosystemNeedCount(readyCount+blockedCount), blockedCount)
	}
	if readyCount > 0 {
		return "ready", formatEcosystemNeedCount(readyCount)
	}
	if blockedCount > 0 {
		return "blocked", fmt.Sprintf("%s but required managers are unavailable", formatEcosystemNeedCount(blockedCount))
	}
	return "idle", "all detected ecosystems are already up to date"
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
