package main

import (
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
