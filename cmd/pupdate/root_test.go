package main

import (
	"bytes"
	"errors"
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

func TestExecuteCmdSuppressesUsageForRuntimeErrors(t *testing.T) {
	cmd := &cobra.Command{Use: "pupdate"}
	cmd.AddCommand(&cobra.Command{
		Use:  "status",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("boom")
		},
	})
	cmd.SetArgs([]string{"status"})

	output := assertExecuteCmdShowsErrorWithoutUsage(t, cmd, "pupdate: error: boom")
	if strings.Count(output, "pupdate: error:") != 1 {
		t.Fatalf("expected exactly one formatted error message, got %q", output)
	}
}

func TestSafeDisplayPreservesOrdinaryTextAndQuotesControls(t *testing.T) {
	if got := safeDisplay("packages/web – ready"); got != "packages/web – ready" {
		t.Fatalf("ordinary output changed: %q", got)
	}
	for _, value := range []string{"line\nbreak", "escape\x1b[2J", "delete\x7fbyte"} {
		got := safeDisplay(value)
		if !strings.HasPrefix(got, `"`) || !strings.HasSuffix(got, `"`) {
			t.Fatalf("expected %q to be quoted, got %q", value, got)
		}
		if strings.Contains(got, "\n") || strings.ContainsRune(got, '\x1b') || strings.ContainsRune(got, '\x7f') {
			t.Fatalf("quoted display still contains raw controls: %q", got)
		}
	}
}

func TestExecuteCmdQuotesControlCharactersInFinalError(t *testing.T) {
	cmd := &cobra.Command{
		Use: "test",
		RunE: func(*cobra.Command, []string) error {
			return errors.New("invalid repo\nname")
		},
	}
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	if code := executeCmd(cmd); code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if got := stderr.String(); got != "pupdate: error: \"invalid repo\\nname\"\n" {
		t.Fatalf("unexpected safe final error: %q", got)
	}
}
