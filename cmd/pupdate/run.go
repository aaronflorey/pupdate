package main

import (
	"os/exec"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/freshness"
	"github.com/spf13/cobra"
)

type runOptions struct {
	Quiet        bool
	AllowScripts bool
}

var detectFn = detection.Detect
var execCommand = exec.CommandContext
var lookPath = exec.LookPath
var evaluateFreshnessFn = freshness.Evaluate

func newRunCmd() *cobra.Command {
	var quiet bool
	var allowScripts bool

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Detect ecosystems and run dependency updates when needed",
		Long: "Detect supported dependency ecosystems in the current directory and depth-1 subdirectories. " +
			"The command emits concise human-readable status lines and only runs installs when dependency inputs changed.",
		Example: `pupdate run
pupdate run --quiet
PUPDATE_SKIP_INSTALL=1 pupdate run`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeRun(cmd, runOptions{Quiet: quiet, AllowScripts: allowScripts})
		},
		SilenceErrors: true,
	}
	cmd.Flags().BoolVar(&quiet, "quiet", false, "suppress no-op output and child command output")
	cmd.Flags().BoolVar(&allowScripts, "allow-scripts", false, "allow dependency manager lifecycle scripts")
	return cmd
}
