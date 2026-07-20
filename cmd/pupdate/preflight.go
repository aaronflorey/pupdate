package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/freshness"
	"github.com/aaronflorey/pupdate/internal/state"
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

func collectPreflight(options preflightOptions) (preflightResult, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return preflightResult{}, fmt.Errorf("failed to resolve working directory: %w", err)
	}

	configPath, err := resolveUserConfigPath()
	if err != nil {
		return preflightResult{}, err
	}

	projectConfigPath := filepath.Join(workingDirectory, projectConfigFileName)
	loadedConfig, err := loadPreflightConfig(configPath, projectConfigPath)
	if err != nil {
		return preflightResult{}, err
	}

	statePath := filepath.Join(workingDirectory, state.FileName)
	stateExists, err := pathExists(statePath)
	if err != nil {
		return preflightResult{}, fmt.Errorf("failed to stat %s: %w", statePath, err)
	}

	result := preflightResult{
		WorkingDirectory:      workingDirectory,
		ConfigPath:            configPath,
		ConfigExists:          loadedConfig.UserExists,
		RawConfig:             loadedConfig.RawUser,
		ResolvedUserConfig:    loadedConfig.ResolvedUser,
		ProjectConfigPath:     projectConfigPath,
		ProjectConfigExists:   loadedConfig.ProjectExists,
		RawProjectConfig:      loadedConfig.RawProject,
		ResolvedProjectConfig: loadedConfig.ResolvedProject,
		ResolvedConfig:        loadedConfig.Effective,
		StatePath:             statePath,
		StateExists:           stateExists,
		Store:                 state.NewStore("."),
	}

	skipReason, err := collectPreflightSkipReason(loadedConfig.Effective)
	if err != nil {
		return preflightResult{}, err
	}
	result.SkipReason = skipReason

	if skipReason != preflightSkipNone && !options.LoadStateOnSkip {
		return result, nil
	}

	currentState, warnings, err := result.Store.Load()
	if err != nil {
		return preflightResult{}, fmt.Errorf("failed to load state: %w", err)
	}
	result.CurrentState = currentState
	result.StateWarnings = warnings

	if skipReason != preflightSkipNone {
		return result, nil
	}

	results, err := detectFn(".", detection.Options{
		WorkspaceGlobs:  loadedConfig.Effective.WorkspaceGlobs,
		FolderBlacklist: loadedConfig.Effective.FolderBlacklist,
	})
	if err != nil {
		return preflightResult{}, fmt.Errorf("detection failed: %w", err)
	}
	result.Results = results

	decisions, err := evaluateFreshnessFn(".", results, currentState)
	if err != nil {
		return preflightResult{}, fmt.Errorf("failed to evaluate dependency freshness: %w", err)
	}
	result.DecisionByEcosystem = indexDecisionsByEcosystem(decisions)

	return result, nil
}

type loadedPreflightConfig struct {
	RawUser         userConfig
	ResolvedUser    userConfig
	UserExists      bool
	RawProject      projectConfig
	ResolvedProject projectConfig
	ProjectExists   bool
	Effective       userConfig
}

func loadPreflightConfig(configPath string, projectPath string) (loadedPreflightConfig, error) {
	configExists, err := pathExists(configPath)
	if err != nil {
		return loadedPreflightConfig{}, fmt.Errorf("failed to stat %s: %w", configPath, err)
	}
	rawConfig, err := readUserConfig(configPath)
	if err != nil {
		return loadedPreflightConfig{}, err
	}
	resolvedConfig, err := resolveUserConfig(rawConfig)
	if err != nil {
		return loadedPreflightConfig{}, err
	}

	projectExists, err := pathExists(projectPath)
	if err != nil {
		return loadedPreflightConfig{}, fmt.Errorf("failed to stat %s: %w", projectPath, err)
	}
	rawProject, err := readProjectConfig(projectPath)
	if err != nil {
		return loadedPreflightConfig{}, err
	}
	resolvedProject, err := resolveProjectConfig(rawProject)
	if err != nil {
		return loadedPreflightConfig{}, err
	}

	return loadedPreflightConfig{
		RawUser:         rawConfig,
		ResolvedUser:    resolvedConfig,
		UserExists:      configExists,
		RawProject:      rawProject,
		ResolvedProject: resolvedProject,
		ProjectExists:   projectExists,
		Effective:       mergeUserAndProjectConfig(resolvedConfig, resolvedProject),
	}, nil
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

func collectPreflightSkipReason(config userConfig) (preflightSkipReason, error) {
	inHomeDir, err := isHomeDirectory()
	if err != nil {
		return preflightSkipNone, err
	}
	if inHomeDir {
		return preflightSkipHomeDirectory, nil
	}

	outsideConfiguredRoots, err := isOutsideConfiguredRootDirectories(config)
	if err != nil {
		return preflightSkipNone, err
	}
	if outsideConfiguredRoots {
		return preflightSkipOutsideRoots, nil
	}

	ignored, err := hasPupIgnore(".")
	if err != nil {
		return preflightSkipNone, fmt.Errorf("failed to check .pupignore: %w", err)
	}
	if ignored {
		return preflightSkipPupIgnore, nil
	}

	return preflightSkipNone, nil
}
