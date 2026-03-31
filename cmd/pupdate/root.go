package main

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pupdate",
		Short:   "Install dependencies on directory entry",
		Version: version,
	}

	cmd.AddCommand(newRunCmd())
	cmd.AddCommand(newInitCmd())

	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
