package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/freshness"
	"github.com/aaronflorey/pupdate/internal/state"
	"github.com/spf13/cobra"
)

type runPayload struct {
	Directory  string                `json:"directory"`
	Ecosystems []runEcosystemPayload `json:"ecosystems"`
	Warnings   []runWarningPayload   `json:"warnings"`
}

type runEcosystemPayload struct {
	Ecosystem    detection.Ecosystem `json:"ecosystem"`
	Managers     []string            `json:"managers"`
	MatchedFiles []string            `json:"matched_files"`
	Warnings     []runWarningPayload `json:"warnings"`
}

type runWarningPayload struct {
	Code    detection.WarningCode `json:"code"`
	Message string                `json:"message"`
}

type ecosystemOutcome struct {
	Ecosystem string
	Succeeded bool
	Lockfiles map[string]string
}

var detectFn = detection.Detect
var execCommand = exec.CommandContext
var lookPath = exec.LookPath
var evaluateFreshnessFn = freshness.Evaluate

func applySuccessfulOutcomes(now time.Time, current state.FileState, outcomes []ecosystemOutcome) state.FileState {
	next := state.FileState{
		Version:    current.Version,
		Ecosystems: make(map[string]state.EcosystemState, len(current.Ecosystems)),
	}
	for key, value := range current.Ecosystems {
		next.Ecosystems[key] = value
	}
	if next.Version == 0 {
		next.Version = state.SchemaVersion
	}

	for _, outcome := range outcomes {
		if !outcome.Succeeded {
			continue
		}
		existing := next.Ecosystems[outcome.Ecosystem]
		lockfiles := existing.Lockfiles
		if len(outcome.Lockfiles) > 0 {
			lockfiles = cloneLockfiles(outcome.Lockfiles)
		}
		next.Ecosystems[outcome.Ecosystem] = state.EcosystemState{
			LastSuccessAt: state.FormatRFC3339UTC(now),
			Lockfiles:     lockfiles,
		}
	}

	return next
}

func newRunCmd() *cobra.Command {
	var quiet bool

	cmd := &cobra.Command{
		Use: "run",
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := detectFn(".")
			if err != nil {
				return fmt.Errorf("detection failed: %w", err)
			}

			store := state.NewStore(".")
			currentState, warnings, err := store.Load()
			if err != nil {
				return fmt.Errorf("failed to load state: %w", err)
			}
			for _, warning := range warnings {
				fmt.Fprintln(cmd.ErrOrStderr(), "pupdate:", warning)
			}

			decisions, err := evaluateFreshnessFn(".", results, currentState)
			if err != nil {
				return fmt.Errorf("failed to evaluate dependency freshness: %w", err)
			}
			decisionByEcosystem := make(map[string]freshness.EcosystemDecision, len(decisions))
			for _, decision := range decisions {
				decisionByEcosystem[decision.Ecosystem] = decision
			}

			outcomes := []ecosystemOutcome{}

			payload := runPayload{
				Directory:  ".",
				Ecosystems: make([]runEcosystemPayload, 0, len(results)),
				Warnings:   make([]runWarningPayload, 0),
			}

			for _, result := range results {
				ecosystemWarnings := make([]runWarningPayload, 0, len(result.Warnings))
				for _, warning := range result.Warnings {
					wp := runWarningPayload{Code: warning.Code, Message: warning.Message}
					ecosystemWarnings = append(ecosystemWarnings, wp)
					payload.Warnings = append(payload.Warnings, wp)
				}

				payload.Ecosystems = append(payload.Ecosystems, runEcosystemPayload{
					Ecosystem:    result.Ecosystem,
					Managers:     result.Managers,
					MatchedFiles: result.MatchedFiles,
					Warnings:     ecosystemWarnings,
				})
			}

			if !quiet {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				if err := encoder.Encode(payload); err != nil {
					return fmt.Errorf("failed to encode output: %w", err)
				}
			}

			ignored, err := hasPupIgnore(".")
			if err != nil {
				return fmt.Errorf("failed to check .pupignore: %w", err)
			}
			if ignored {
				fmt.Fprintln(cmd.ErrOrStderr(), "pupdate: skip repo (.pupignore)")
				return nil
			}

			installDisabled := isInstallDisabled()
			if installDisabled {
				fmt.Fprintln(cmd.ErrOrStderr(), "pupdate: installs disabled via PUPDATE_SKIP_INSTALL")
			}

			for _, result := range results {
				decision, ok := decisionByEcosystem[string(result.Ecosystem)]
				if !ok {
					continue
				}
				if decision.Decision != freshness.DecisionUpdate {
					if result.Ecosystem == detection.EcosystemGit && strings.HasPrefix(decision.Reason, "git submodule status failed:") {
						fmt.Fprintf(cmd.ErrOrStderr(), "pupdate: error %s\n", decision.Reason)
						continue
					}
					reason := decision.Reason
					if reason == "" {
						reason = "up-to-date"
					}
					fmt.Fprintf(cmd.ErrOrStderr(), "pupdate: skip %s (%s)\n", result.Ecosystem, reason)
					continue
				}
				if installDisabled {
					continue
				}

				plan, ok, reason := selectManagerPlan(result)
				if !ok {
					if reason != "" {
						fmt.Fprintln(cmd.ErrOrStderr(), "pupdate:", reason)
					}
					continue
				}

				if _, err := lookPath(plan.Manager); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "pupdate: skip %s (%s not found on PATH)\n", result.Ecosystem, plan.Manager)
					continue
				}

				fmt.Fprintf(cmd.ErrOrStderr(), "pupdate: run %s %s\n", plan.Manager, strings.Join(plan.Args, " "))
				err = runInstall(cmd, quiet, plan.Manager, plan.Args...)
				outcomes = append(outcomes, ecosystemOutcome{
					Ecosystem: string(result.Ecosystem),
					Succeeded: err == nil,
					Lockfiles: decision.Lockfiles,
				})
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "pupdate: error %s install failed: %v\n", plan.Manager, err)
				}
			}

			hasSuccess := false
			for _, outcome := range outcomes {
				if outcome.Succeeded {
					hasSuccess = true
					break
				}
			}
			if hasSuccess {
				updated := applySuccessfulOutcomes(time.Now().UTC(), currentState, outcomes)
				if err := store.Save(updated); err != nil {
					return fmt.Errorf("failed to save state: %w", err)
				}
			}

			return nil
		},
		SilenceErrors: true,
	}
	cmd.Flags().BoolVar(&quiet, "quiet", false, "suppress output")
	return cmd
}

func runInstall(cmd *cobra.Command, quiet bool, name string, args ...string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Minute)
	defer cancel()

	command := execCommand(ctx, name, args...)
	if quiet {
		command.Stdout = io.Discard
		command.Stderr = io.Discard
	} else {
		command.Stdout = cmd.ErrOrStderr()
		command.Stderr = cmd.ErrOrStderr()
	}

	return command.Run()
}

type managerPlan struct {
	Manager string
	Args    []string
}

func selectManagerPlan(result detection.DetectionResult) (managerPlan, bool, string) {
	switch result.Ecosystem {
	case detection.EcosystemPHP:
		return managerPlan{Manager: "composer", Args: []string{"install", "--no-interaction", "--prefer-dist", "--no-scripts"}}, true, ""
	case detection.EcosystemNode:
		if len(result.Managers) != 1 {
			return managerPlan{}, false, "multiple Node lockfiles detected; skipping install"
		}
		switch result.Managers[0] {
		case "bun":
			return managerPlan{Manager: "bun", Args: []string{"install", "--frozen-lockfile", "--ignore-scripts"}}, true, ""
		case "npm":
			return managerPlan{Manager: "npm", Args: []string{"ci", "--ignore-scripts"}}, true, ""
		case "pnpm":
			return managerPlan{Manager: "pnpm", Args: []string{"install", "--frozen-lockfile", "--ignore-scripts"}}, true, ""
		case "yarn":
			return managerPlan{Manager: "yarn", Args: []string{"install", "--frozen-lockfile", "--ignore-scripts"}}, true, ""
		default:
			return managerPlan{}, false, fmt.Sprintf("unsupported Node manager %q; skipping install", result.Managers[0])
		}
	case detection.EcosystemPython:
		if len(result.Managers) != 1 {
			return managerPlan{}, false, "multiple Python lockfiles detected; skipping install"
		}
		switch result.Managers[0] {
		case "uv":
			return managerPlan{Manager: "uv", Args: []string{"sync", "--frozen"}}, true, ""
		case "poetry":
			return managerPlan{Manager: "poetry", Args: []string{"install", "--no-interaction", "--sync"}}, true, ""
		case "pip":
			return managerPlan{Manager: "pip", Args: []string{"install", "-r", "requirements.txt", "--disable-pip-version-check", "--no-input"}}, true, ""
		default:
			return managerPlan{}, false, fmt.Sprintf("unsupported Python manager %q; skipping install", result.Managers[0])
		}
	case detection.EcosystemGo:
		return managerPlan{Manager: "go", Args: []string{"mod", "download"}}, true, ""
	case detection.EcosystemRust:
		return managerPlan{Manager: "cargo", Args: []string{"fetch", "--locked"}}, true, ""
	case detection.EcosystemGit:
		return managerPlan{Manager: "git", Args: []string{"submodule", "update", "--init", "--recursive"}}, true, ""
	default:
		return managerPlan{}, false, fmt.Sprintf("unsupported ecosystem %q; skipping install", result.Ecosystem)
	}
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

func cloneLockfiles(lockfiles map[string]string) map[string]string {
	if len(lockfiles) == 0 {
		return map[string]string{}
	}
	cloned := make(map[string]string, len(lockfiles))
	for key, value := range lockfiles {
		cloned[key] = value
	}
	return cloned
}

func isInstallDisabled() bool {
	value := strings.TrimSpace(os.Getenv("PUPDATE_SKIP_INSTALL"))
	if value == "" {
		return false
	}
	value = strings.ToLower(value)
	return value == "1" || value == "true" || value == "yes"
}
