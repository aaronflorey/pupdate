package install

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/git-pkgs/managers"
	"github.com/git-pkgs/managers/definitions"
)

type Plan struct {
	Manager string
	Args    []string
}

var (
	installTranslatorOnce sync.Once
	installTranslator     *managers.Translator
	installTranslatorErr  error
)

func SelectPlan(result detection.DetectionResult, allowScripts bool) (Plan, bool, string) {
	switch result.Ecosystem {
	case detection.EcosystemPHP:
		extra := []string{"--no-interaction", "--prefer-dist"}
		if !allowScripts {
			extra = append(extra, "--no-plugins", "--no-scripts")
		}
		return buildInstallPlan("composer", managers.CommandInput{Extra: extra})
	case detection.EcosystemNode:
		if len(result.MatchedFiles) != 1 || len(result.Managers) != 1 {
			return Plan{}, false, "multiple Node lockfiles detected; skipping install"
		}
		for _, warning := range result.Warnings {
			if warning.Code == detection.WarningYarnModernUnsupported {
				return Plan{}, false, "Yarn Modern is not supported; only Yarn Classic 1.x can be installed"
			}
		}
		return selectNodePlan(result.Managers[0], allowScripts)
	case detection.EcosystemPython:
		if len(result.Managers) != 1 {
			return Plan{}, false, "multiple Python lockfiles detected; skipping install"
		}

		plan, ok, reason := selectPythonManagerPlan(result.Managers[0])
		if !ok {
			return Plan{}, false, reason
		}
		if !allowScripts {
			return Plan{}, false, fmt.Sprintf("python manager %s can execute install/build code; rerun with --allow-scripts to allow", result.Managers[0])
		}
		return plan, true, ""
	case detection.EcosystemKasetto:
		if len(result.Managers) != 1 {
			return Plan{}, false, "multiple Kasetto signals detected; skipping install"
		}
		if result.Managers[0] != "kst" {
			return Plan{}, false, fmt.Sprintf("unsupported Kasetto manager %q; skipping install", result.Managers[0])
		}
		configPath := kasettoConfigPath(result.MatchedFiles)
		if configPath == "" {
			return Plan{}, false, "Kasetto project has no local kasetto.yaml or kasetto.yml; skipping install"
		}
		return Plan{Manager: "kst", Args: []string{"sync", "--project", "--config", configPath}}, true, ""
	case detection.EcosystemGo:
		return buildInstallPlan("gomod", managers.CommandInput{})
	case detection.EcosystemRust:
		return buildInstallPlan("cargo", managers.CommandInput{Flags: map[string]any{"locked": true}})
	case detection.EcosystemGit:
		return Plan{Manager: "git", Args: []string{"submodule", "update", "--init", "--recursive"}}, true, ""
	default:
		return Plan{}, false, fmt.Sprintf("unsupported ecosystem %q; skipping install", result.Ecosystem)
	}
}

func selectNodePlan(managerName string, allowScripts bool) (Plan, bool, string) {
	switch managerName {
	case "bun", "npm", "pnpm", "yarn":
	default:
		return Plan{}, false, fmt.Sprintf("unsupported Node manager %q; skipping install", managerName)
	}

	input := managers.CommandInput{Flags: map[string]any{"frozen": true}}
	if !allowScripts {
		input.Extra = []string{"--ignore-scripts"}
	}
	return buildInstallPlan(managerName, input)
}

func kasettoConfigPath(matchedFiles []string) string {
	for _, matchedFile := range matchedFiles {
		base := filepath.Base(matchedFile)
		if strings.EqualFold(base, "kasetto.yaml") || strings.EqualFold(base, "kasetto.yml") {
			return base
		}
	}

	return ""
}

func selectPythonManagerPlan(managerName string) (Plan, bool, string) {
	switch managerName {
	case "uv":
		return buildInstallPlan("uv", managers.CommandInput{Flags: map[string]any{"frozen": true}})
	case "poetry":
		return buildInstallPlan("poetry", managers.CommandInput{
			Extra: []string{"--no-interaction", "--sync"},
		})
	case "pip":
		return buildInstallPlan("pip", managers.CommandInput{Extra: []string{"--disable-pip-version-check", "--no-input"}})
	default:
		return Plan{}, false, fmt.Sprintf("unsupported Python manager %q; skipping install", managerName)
	}
}

func buildInstallPlan(managerName string, input managers.CommandInput) (Plan, bool, string) {
	translator, err := loadInstallTranslator()
	if err != nil {
		return Plan{}, false, fmt.Sprintf("failed to load manager definitions: %v", err)
	}

	command, err := translator.BuildCommand(managerName, "install", input)
	if err != nil {
		return Plan{}, false, fmt.Sprintf("failed to build %s install command: %v", managerName, err)
	}
	if len(command) == 0 {
		return Plan{}, false, fmt.Sprintf("failed to build %s install command: empty command", managerName)
	}

	return Plan{Manager: command[0], Args: command[1:]}, true, ""
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
