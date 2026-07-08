package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func assertExecuteCmdShowsErrorWithoutUsage(t *testing.T, cmd *cobra.Command, want string) string {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	code := executeCmd(cmd)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}

	output := stderr.String()
	if !strings.Contains(output, want) {
		t.Fatalf("expected actionable error on configured stderr, got %q", output)
	}
	if strings.Contains(output, "Usage:") {
		t.Fatalf("expected runtime error to avoid usage output, got %q", output)
	}

	return output
}

func TestExecuteCmdReturnsZeroWithoutPrintingOnSuccess(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--version"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	previousFactory := rootCommandFactory
	rootCommandFactory = func() *cobra.Command { return cmd }
	t.Cleanup(func() { rootCommandFactory = previousFactory })

	code := execute()
	if code != 0 {
		t.Fatalf("expected zero exit code, got %d", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected success to avoid stderr output, got %q", stderr.String())
	}
}

func TestExecuteCmdPrintsSingleErrorToConfiguredStderr(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"bogus"})
	previousFactory := rootCommandFactory
	rootCommandFactory = func() *cobra.Command { return cmd }
	t.Cleanup(func() { rootCommandFactory = previousFactory })

	output := assertExecuteCmdShowsErrorWithoutUsage(t, cmd, "pupdate: error: unknown command \"bogus\" for \"pupdate\"")
	if !strings.Contains(output, "pupdate: error: unknown command \"bogus\" for \"pupdate\"") {
		t.Fatalf("expected actionable error on configured stderr, got %q", output)
	}
	if strings.Count(output, "pupdate: error:") != 1 {
		t.Fatalf("expected exactly one formatted error message, got %q", output)
	}
	if strings.Contains(output, "Error:") {
		t.Fatalf("expected cobra default error emission to be suppressed, got %q", output)
	}
}
