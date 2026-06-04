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
	WorkingDirectory    string
	ConfigPath          string
	ConfigExists        bool
	RawConfig           userConfig
	ResolvedConfig      userConfig
	StatePath           string
	StateExists         bool
	Store               state.Store
	CurrentState        state.FileState
	StateWarnings       []string
	Results             []detection.DetectionResult
	DecisionByEcosystem map[string]freshness.EcosystemDecision
	SkipReason          preflightSkipReason
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

	rawConfig, resolvedConfig, configExists, err := loadPreflightConfig(configPath)
	if err != nil {
		return preflightResult{}, err
	}

	statePath := filepath.Join(workingDirectory, state.FileName)
	stateExists, err := pathExists(statePath)
	if err != nil {
		return preflightResult{}, fmt.Errorf("failed to stat %s: %w", statePath, err)
	}

	result := preflightResult{
		WorkingDirectory: workingDirectory,
		ConfigPath:       configPath,
		ConfigExists:     configExists,
		RawConfig:        rawConfig,
		ResolvedConfig:   resolvedConfig,
		StatePath:        statePath,
		StateExists:      stateExists,
		Store:            state.NewStore("."),
	}

	skipReason, err := collectPreflightSkipReason(resolvedConfig)
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

	results, err := detectFn(".", resolvedConfig.WorkspaceGlobs)
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

func loadPreflightConfig(configPath string) (userConfig, userConfig, bool, error) {
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
