package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

type BuildInfo struct {
	Version string
	Commit  string
	Date    string
	BuiltBy string
}

type IOStreams struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

func newRootCmd() *cobra.Command {
	return newRootCmdWithDependencies(BuildInfo{Version: "dev", Commit: "none", Date: "unknown", BuiltBy: "unknown"}, defaultDependencies())
}

func newRootCmdWithBuildInfo(build BuildInfo) *cobra.Command {
	return newRootCmdWithDependencies(build, defaultDependencies())
}

func newRootCmdWithDependencies(build BuildInfo, deps dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pupdate",
		Short: "Install dependencies on directory entry",
		Long: "pupdate detects supported dependency ecosystems in a directory and runs safe, " +
			"non-interactive installs when manifests change. It ships a shell hook that runs " +
			"automatically on directory entry, and caches successful installs in local state.",
		Version: build.Version,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Context().Err()
		},
	}

	cmd.AddCommand(newRunCmd())
	cmd.AddCommand(newHookCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newResetCmd())
	cmd.SetContext(contextWithDependencies(context.Background(), deps))

	return cmd
}

func ExecuteContext(ctx context.Context, build BuildInfo, args []string, streams IOStreams) int {
	deps := defaultDependencies()
	cmd := newRootCmdWithDependencies(build, deps)
	cmd.SetArgs(args)
	cmd.SetIn(streams.In)
	cmd.SetOut(streams.Out)
	cmd.SetErr(streams.Err)
	cmd.SetContext(contextWithDependencies(ctx, deps))
	return executeCmd(cmd)
}

func executeCmd(cmd *cobra.Command) int {
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	if err := cmd.Execute(); err != nil {
		if errors.Is(cmd.Context().Err(), context.Canceled) {
			err = errors.New("operation canceled")
		}
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "pupdate: error: %s\n", safeDisplay(err.Error()))
		return 1
	}

	return 0
}
