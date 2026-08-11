package install

import (
	"context"
	"io"
	"os/exec"
	"time"
)

const DefaultTimeout = 30 * time.Minute

type CommandFactory func(context.Context, string, ...string) *exec.Cmd

type Runner struct {
	Command CommandFactory
	Timeout time.Duration
}

func (r Runner) Run(ctx context.Context, quiet bool, workDir string, plan Plan, stderr io.Writer) error {
	timeout := r.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	commandFactory := r.Command
	if commandFactory == nil {
		commandFactory = exec.CommandContext
	}
	command := commandFactory(ctx, plan.Manager, plan.Args...)
	configureInstallCommand(command)
	command.Dir = workDir
	if quiet {
		command.Stdout = io.Discard
		command.Stderr = io.Discard
	} else {
		command.Stdout = stderr
		command.Stderr = stderr
	}

	return command.Run()
}
