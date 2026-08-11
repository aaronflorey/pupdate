package cli

import (
	"context"
	"os"
	"os/exec"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/freshness"
	"github.com/aaronflorey/pupdate/internal/install"
	"github.com/aaronflorey/pupdate/internal/shellhook"
	"github.com/spf13/cobra"
)

var detectFn = detection.DetectWithOptions
var execCommand = exec.CommandContext
var lookPath = exec.LookPath
var evaluateFreshnessFn = freshness.EvaluateContext
var snapshotFreshnessFn = freshness.SnapshotMatchedFiles
var installTimeout = install.DefaultTimeout
var resolveExecutablePath = os.Executable
var startBackgroundProcess = shellhook.StartBackground
var executeRunFn = executeRun

func testDependencies() dependencies {
	deps := defaultDependencies().withEnvironment(currentEnvironment())
	deps.Detect = detectFn
	deps.Command = execCommand
	deps.LookPath = lookPath
	deps.EvaluateFreshness = evaluateFreshnessFn
	deps.SnapshotFreshness = snapshotFreshnessFn
	deps.InstallTimeout = installTimeout
	deps.ResolveExecutable = resolveExecutablePath
	deps.StartBackground = startBackgroundProcess
	deps.Run = executeRunFn
	return deps
}

func newTestRootCmd() *cobra.Command {
	return newRootCmdWithDependencies(BuildInfo{
		Version: "dev", Commit: "none", Date: "unknown", BuiltBy: "unknown",
	}, testDependencies())
}

func newTestRunCmd() *cobra.Command {
	cmd := newRunCmd()
	cmd.SetContext(contextWithDependencies(context.Background(), testDependencies()))
	return cmd
}

func newTestHookCmd() *cobra.Command {
	cmd := newHookCmd()
	cmd.SetContext(contextWithDependencies(context.Background(), testDependencies()))
	return cmd
}
