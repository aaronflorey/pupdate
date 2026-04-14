package main

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/spf13/cobra"
)

type managerPlan struct {
	Manager string
	Args    []string
}

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
		args := []string{"install", "--no-interaction", "--prefer-dist"}
		if !allowScripts {
			args = append(args, "--no-scripts")
		}
		return managerPlan{Manager: "composer", Args: args}, true, ""
	case detection.EcosystemNode:
		if len(result.Managers) != 1 {
			return managerPlan{}, false, "multiple Node lockfiles detected; skipping install"
		}
		switch result.Managers[0] {
		case "bun":
			return managerPlan{Manager: "bun", Args: nodeInstallArgs("install", "--frozen-lockfile", allowScripts)}, true, ""
		case "npm":
			return managerPlan{Manager: "npm", Args: nodeInstallArgs("ci", "", allowScripts)}, true, ""
		case "pnpm":
			return managerPlan{Manager: "pnpm", Args: nodeInstallArgs("install", "--frozen-lockfile", allowScripts)}, true, ""
		case "yarn":
			return managerPlan{Manager: "yarn", Args: nodeInstallArgs("install", "--frozen-lockfile", allowScripts)}, true, ""
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

func nodeInstallArgs(command string, frozenFlag string, allowScripts bool) []string {
	args := []string{command}
	if frozenFlag != "" {
		args = append(args, frozenFlag)
	}
	if !allowScripts {
		args = append(args, "--ignore-scripts")
	}
	return args
}
