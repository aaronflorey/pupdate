package cli

import "github.com/aaronflorey/pupdate/internal/shellhook"

const (
	managedHookBegin = shellhook.ManagedBegin
	managedHookEnd   = shellhook.ManagedEnd
)

func resolvedExecutablePath() string {
	return shellhook.ResolvedExecutablePath()
}

func shellQuote(value string) string {
	return shellhook.ShellQuote(value)
}

func defaultManagedHookTarget(shell string) (string, error) {
	return shellhook.DefaultTarget(shell)
}
