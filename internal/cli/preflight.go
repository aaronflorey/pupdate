package cli

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/freshness"
	"github.com/aaronflorey/pupdate/internal/state"
	"github.com/aaronflorey/pupdate/internal/userenv"
)

type preflightSkipReason string

const (
	preflightSkipNone          preflightSkipReason = ""
	preflightSkipHomeDirectory preflightSkipReason = "home_directory"
	preflightSkipOutsideRoots  preflightSkipReason = "outside_configured_roots"
	preflightSkipPupIgnore     preflightSkipReason = "pupignore"
)

type preflightOptions struct {
	LoadStateOnSkip bool
}

type preflightResult struct {
	WorkingDirectory      string
	ConfigPath            string
	ConfigExists          bool
	RawConfig             userConfig
	ResolvedUserConfig    userConfig
	ProjectConfigPath     string
	ProjectConfigExists   bool
	RawProjectConfig      projectConfig
	ResolvedProjectConfig projectConfig
	ResolvedConfig        userConfig
	StatePath             string
	StateExists           bool
	Store                 state.Store
	CurrentState          state.FileState
	StateWarnings         []string
	Results               []detection.DetectionResult
	DecisionByEcosystem   map[string]freshness.EcosystemDecision
	SkipReason            preflightSkipReason
}

func collectPreflight(ctx context.Context, workingDirectory string, options preflightOptions, deps dependencies) (preflightResult, error) {
	configSnapshot, err := deps.ConfigLoader.Load(workingDirectory)
	if err != nil {
		return preflightResult{}, err
	}
	configPath := configSnapshot.User.Path
	projectConfigPath := configSnapshot.Project.Path

	statePath := filepath.Join(workingDirectory, state.FileName)

	result := preflightResult{
		WorkingDirectory:      workingDirectory,
		ConfigPath:            configPath,
		ConfigExists:          configSnapshot.User.Exists,
		RawConfig:             configSnapshot.User.Raw,
		ResolvedUserConfig:    configSnapshot.User.Resolved,
		ProjectConfigPath:     projectConfigPath,
		ProjectConfigExists:   configSnapshot.Project.Exists,
		RawProjectConfig:      configSnapshot.Project.Raw,
		ResolvedProjectConfig: configSnapshot.Project.Resolved,
		ResolvedConfig:        configSnapshot.Effective,
		StatePath:             statePath,
		Store:                 state.NewStore(workingDirectory),
	}

	skipReason, err := collectPreflightSkipReason(workingDirectory, configSnapshot.Effective, deps.Environment)
	if err != nil {
		return preflightResult{}, err
	}
	result.SkipReason = skipReason

	if skipReason != preflightSkipNone && !options.LoadStateOnSkip {
		return result, nil
	}

	loadedState, err := result.Store.Load()
	if err != nil {
		return preflightResult{}, fmt.Errorf("failed to load state: %w", err)
	}
	result.CurrentState = loadedState.State
	result.StateWarnings = loadedState.Warnings
	result.StateExists = loadedState.Exists

	if skipReason != preflightSkipNone {
		return result, nil
	}

	results, err := deps.Detect(workingDirectory, detection.Options{
		WorkspaceGlobs:  configSnapshot.Effective.WorkspaceGlobs,
		FolderBlacklist: configSnapshot.Effective.FolderBlacklist,
	})
	if err != nil {
		return preflightResult{}, fmt.Errorf("detection failed: %w", err)
	}
	result.Results = results

	decisions, err := deps.EvaluateFreshness(ctx, workingDirectory, results, loadedState.State)
	if err != nil {
		return preflightResult{}, fmt.Errorf("failed to evaluate dependency freshness: %w", err)
	}
	result.DecisionByEcosystem = indexDecisionsByEcosystem(decisions)

	return result, nil
}

func collectPreflightSkipReason(workingDirectory string, config userConfig, environment userenv.Environment) (preflightSkipReason, error) {
	inHomeDir := isHomeDirectory(workingDirectory, environment)
	if inHomeDir {
		return preflightSkipHomeDirectory, nil
	}

	outsideConfiguredRoots := isOutsideConfiguredRootDirectories(workingDirectory, config)
	if outsideConfiguredRoots {
		return preflightSkipOutsideRoots, nil
	}

	ignored, err := hasPupIgnore(workingDirectory)
	if err != nil {
		return preflightSkipNone, fmt.Errorf("failed to check .pupignore: %w", err)
	}
	if ignored {
		return preflightSkipPupIgnore, nil
	}

	return preflightSkipNone, nil
}
