package cli

import (
	"context"
	"os"
	"os/exec"
	"time"

	configpkg "github.com/aaronflorey/pupdate/internal/config"
	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/freshness"
	"github.com/aaronflorey/pupdate/internal/install"
	"github.com/aaronflorey/pupdate/internal/shellhook"
	"github.com/aaronflorey/pupdate/internal/state"
	"github.com/aaronflorey/pupdate/internal/userenv"
	"github.com/spf13/cobra"
)

type dependencies struct {
	ConfigLoader      configpkg.Loader
	Environment       userenv.Environment
	Detect            func(string, detection.Options) ([]detection.DetectionResult, error)
	EvaluateFreshness func(context.Context, string, []detection.DetectionResult, state.FileState) ([]freshness.EcosystemDecision, error)
	SnapshotFreshness func(context.Context, string, []string, state.EcosystemState) (freshness.MatchedFilesSnapshot, error)
	LookPath          func(string) (string, error)
	Command           install.CommandFactory
	InstallTimeout    time.Duration
	ResolveExecutable func() (string, error)
	StartBackground   func(string, []string) (int, error)
	Run               func(*cobra.Command, bool, bool, bool) error
}

func defaultDependencies() dependencies {
	environment := userenv.System()
	return dependencies{
		ConfigLoader:      configpkg.Loader{Environment: environment},
		Environment:       environment,
		Detect:            detection.DetectWithOptions,
		EvaluateFreshness: freshness.EvaluateContext,
		SnapshotFreshness: freshness.SnapshotMatchedFiles,
		LookPath:          exec.LookPath,
		Command:           exec.CommandContext,
		InstallTimeout:    install.DefaultTimeout,
		ResolveExecutable: os.Executable,
		StartBackground:   shellhook.StartBackground,
		Run:               executeRun,
	}
}

func (d dependencies) withEnvironment(environment userenv.Environment) dependencies {
	d.Environment = environment
	d.ConfigLoader = configpkg.Loader{Environment: environment}
	return d
}

type dependenciesContextKey struct{}

func contextWithDependencies(ctx context.Context, deps dependencies) context.Context {
	return context.WithValue(ctx, dependenciesContextKey{}, deps)
}

func dependenciesFor(cmd *cobra.Command) dependencies {
	if cmd != nil {
		if deps, ok := cmd.Context().Value(dependenciesContextKey{}).(dependencies); ok {
			return deps
		}
	}
	return defaultDependencies()
}
