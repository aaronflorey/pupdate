package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/git-pkgs/managers"
	"github.com/git-pkgs/managers/definitions"
	"github.com/spf13/cobra"
)

type managerPlan struct {
	Manager string
	Args    []string
}

var (
	installTranslatorOnce sync.Once
	installTranslator     *managers.Translator
	installTranslatorErr  error
)

func runInstall(cmd *cobra.Command, quiet bool, workDir string, name string, args ...string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Minute)
	defer cancel()

	command := execCommand(ctx, name, args...)
	command.Dir = workDir
	if quiet {
		command.Stdout = io.Discard
		command.Stderr = io.Discard
	} else {
		command.Stdout = cmd.ErrOrStderr()
		command.Stderr = cmd.ErrOrStderr()
	}

	return command.Run()
}

func selectManagerPlan(result detection.DetectionResult, allowScripts bool) (managerPlan, bool, string) {
	switch result.Ecosystem {
	case detection.EcosystemPHP:
		extra := []string{"--no-interaction", "--prefer-dist"}
		if !allowScripts {
			extra = append(extra, "--no-scripts")
		}
		return buildInstallPlan("composer", managers.CommandInput{Extra: extra})
	case detection.EcosystemNode:
		if len(result.Managers) != 1 {
			return managerPlan{}, false, "multiple Node lockfiles detected; skipping install"
		}
		switch result.Managers[0] {
		case "bun":
			input := managers.CommandInput{Flags: map[string]any{"frozen": true}}
			if !allowScripts {
				input.Extra = append(input.Extra, "--ignore-scripts")
			}
			return buildInstallPlan("bun", input)
		case "npm":
			input := managers.CommandInput{Flags: map[string]any{"frozen": true}}
			if !allowScripts {
				input.Extra = append(input.Extra, "--ignore-scripts")
			}
			return buildInstallPlan("npm", input)
		case "pnpm":
			input := managers.CommandInput{Flags: map[string]any{"frozen": true}}
			if !allowScripts {
				input.Extra = append(input.Extra, "--ignore-scripts")
			}
			return buildInstallPlan("pnpm", input)
		case "yarn":
			input := managers.CommandInput{Flags: map[string]any{"frozen": true}}
			if !allowScripts {
				input.Extra = append(input.Extra, "--ignore-scripts")
			}
			return buildInstallPlan("yarn", input)
		default:
			return managerPlan{}, false, fmt.Sprintf("unsupported Node manager %q; skipping install", result.Managers[0])
		}
	case detection.EcosystemPython:
		if len(result.Managers) != 1 {
			return managerPlan{}, false, "multiple Python lockfiles detected; skipping install"
		}
		switch result.Managers[0] {
		case "uv":
			return buildInstallPlan("uv", managers.CommandInput{Flags: map[string]any{"frozen": true}})
		case "poetry":
			return buildInstallPlan("poetry", managers.CommandInput{
				Extra: []string{"--no-interaction", "--sync"},
			})
		case "pip":
			return buildInstallPlan("pip", managers.CommandInput{Extra: []string{"--disable-pip-version-check", "--no-input"}})
		default:
			return managerPlan{}, false, fmt.Sprintf("unsupported Python manager %q; skipping install", result.Managers[0])
		}
	case detection.EcosystemGo:
		return buildInstallPlan("gomod", managers.CommandInput{})
	case detection.EcosystemRust:
		return buildInstallPlan("cargo", managers.CommandInput{Flags: map[string]any{"locked": true}})
	case detection.EcosystemGit:
		return managerPlan{Manager: "git", Args: []string{"submodule", "update", "--init", "--recursive"}}, true, ""
	default:
		return managerPlan{}, false, fmt.Sprintf("unsupported ecosystem %q; skipping install", result.Ecosystem)
	}
}

func buildInstallPlan(managerName string, input managers.CommandInput) (managerPlan, bool, string) {
	translator, err := loadInstallTranslator()
	if err != nil {
		return managerPlan{}, false, fmt.Sprintf("failed to load manager definitions: %v", err)
	}

	command, err := translator.BuildCommand(managerName, "install", input)
	if err != nil {
		return managerPlan{}, false, fmt.Sprintf("failed to build %s install command: %v", managerName, err)
	}
	if len(command) == 0 {
		return managerPlan{}, false, fmt.Sprintf("failed to build %s install command: empty command", managerName)
	}

	return managerPlan{Manager: command[0], Args: command[1:]}, true, ""
}

func loadInstallTranslator() (*managers.Translator, error) {
	installTranslatorOnce.Do(func() {
		defs, err := definitions.LoadEmbedded()
		if err != nil {
			installTranslatorErr = err
			return
		}

		translator := managers.NewTranslator()
		for _, def := range defs {
			translator.Register(def)
		}

		if _, ok := translator.Definition("gomod"); !ok {
			installTranslatorErr = errors.New("gomod definition not loaded")
			return
		}

		installTranslator = translator
	})

	return installTranslator, installTranslatorErr
}
