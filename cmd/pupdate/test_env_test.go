package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(runTestsWithIsolatedConfigHome(m))
}

func runTestsWithIsolatedConfigHome(m *testing.M) int {
	configHome, err := os.MkdirTemp("", "pupdate-cmd-test-config-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create isolated config home: %v\n", err)
		return 1
	}
	defer os.RemoveAll(configHome)

	previous, hadPrevious := os.LookupEnv("XDG_CONFIG_HOME")
	if err := os.Setenv("XDG_CONFIG_HOME", configHome); err != nil {
		fmt.Fprintf(os.Stderr, "set XDG_CONFIG_HOME: %v\n", err)
		return 1
	}
	defer func() {
		if hadPrevious {
			_ = os.Setenv("XDG_CONFIG_HOME", previous)
			return
		}
		_ = os.Unsetenv("XDG_CONFIG_HOME")
	}()

	return m.Run()
}

func executeRootCommand(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := newRootCmd()
	cmd.SetArgs(args)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}
