package main

import (
	"os"

	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pupdate",
		Short: "Install dependencies on directory entry",
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
