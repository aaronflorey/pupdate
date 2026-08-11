package cli

import (
	"github.com/aaronflorey/pupdate/internal/install"
	"github.com/spf13/cobra"
)

func runInstall(cmd *cobra.Command, quiet bool, workDir string, deps dependencies, name string, args ...string) error {
	runner := install.Runner{
		Command: deps.Command,
		Timeout: deps.InstallTimeout,
	}
	return runner.Run(cmd.Context(), quiet, workDir, install.Plan{Manager: name, Args: args}, cmd.ErrOrStderr())
}
