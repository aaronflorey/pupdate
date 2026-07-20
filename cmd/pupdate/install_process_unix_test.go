//go:build darwin || linux

package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestInstallTimeoutKillsManagerDescendants(t *testing.T) {
	if os.Getenv("PUPDATE_PROCESS_HELPER") != "" {
		return
	}

	pidFile := t.TempDir() + "/child.pid"
	previousExec := execCommand
	previousTimeout := installTimeout
	t.Cleanup(func() {
		execCommand = previousExec
		installTimeout = previousTimeout
	})
	installTimeout = 500 * time.Millisecond
	execCommand = func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
		command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestInstallProcessGroupHelper$")
		command.Env = append(os.Environ(), "PUPDATE_PROCESS_HELPER=parent", "PUPDATE_PROCESS_PID_FILE="+pidFile)
		return command
	}

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	if err := runInstall(cmd, true, ".", "helper"); err == nil {
		t.Fatal("expected helper command to time out")
	}

	raw, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("read descendant pid: %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatalf("parse descendant pid: %v", err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		err = syscall.Kill(pid, 0)
		if errors.Is(err, syscall.ESRCH) {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("manager descendant %d remained alive after timeout", pid)
}

func TestInstallProcessGroupHelper(t *testing.T) {
	mode := os.Getenv("PUPDATE_PROCESS_HELPER")
	if mode == "" {
		return
	}
	if mode == "child" {
		time.Sleep(30 * time.Second)
		return
	}

	command := exec.Command(os.Args[0], "-test.run=^TestInstallProcessGroupHelper$")
	command.Env = append(os.Environ(), "PUPDATE_PROCESS_HELPER=child")
	if err := command.Start(); err != nil {
		os.Exit(2)
	}
	pidFile := os.Getenv("PUPDATE_PROCESS_PID_FILE")
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(command.Process.Pid)), 0o600); err != nil {
		os.Exit(3)
	}
	time.Sleep(30 * time.Second)
}
