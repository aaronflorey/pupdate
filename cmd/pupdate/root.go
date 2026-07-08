package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

var rootCommandFactory = newRootCmd

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pupdate",
		Short:   "Install dependencies on directory entry",
		Version: version,
	}

	cmd.AddCommand(newRunCmd())
	cmd.AddCommand(newHookCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newResetCmd())

	return cmd
}

func main() {
	os.Exit(execute())
}

func execute() int {
	return executeCmd(rootCommandFactory())
}

func executeCmd(cmd *cobra.Command) int {
	cmd.SilenceErrors = true
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "pupdate: error: %v\n", err)
		return 1
	}

	return 0
}
