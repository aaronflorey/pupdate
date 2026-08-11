package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func executeInitManage(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newTestRootCmd()
	cmd.SetArgs(append([]string{"init"}, args...))
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func TestInitInstallPreviewDoesNotChangeTarget(t *testing.T) {
	dir := t.TempDir()
	withChdir(t, dir)
	target := filepath.Join(dir, ".bashrc")
	original := []byte("export TEST=1\n")
	if err := os.WriteFile(target, original, 0o640); err != nil {
		t.Fatalf("write target: %v", err)
	}

	stdout, stderr, err := executeInitManage(t, "install", "--shell", "bash", "--target", ".bashrc")
	if err != nil || stderr != "" {
		t.Fatalf("preview failed: stderr=%q err=%v", stderr, err)
	}
	if !strings.Contains(stdout, "target: "+target) || !strings.Contains(stdout, "operation: install") || !strings.Contains(stdout, "apply: false") {
		t.Fatalf("unexpected preview: %q", stdout)
	}
	raw, err := os.ReadFile(target)
	if err != nil || !bytes.Equal(raw, original) {
		t.Fatalf("preview changed target: raw=%q err=%v", raw, err)
	}
	if _, err := os.Stat(target + ".pupdate.bak"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("preview created backup, err=%v", err)
	}
}

func TestInitInstallApplyBacksUpPreservesModeAndIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, ".zshrc")
	original := []byte("export ORIGINAL=yes\n")
	if err := os.WriteFile(target, original, 0o640); err != nil {
		t.Fatalf("write target: %v", err)
	}

	stdout, _, err := executeInitManage(t, "install", "--shell", "zsh", "--target", target, "--apply")
	if err != nil || !strings.Contains(stdout, "installed pupdate hook") {
		t.Fatalf("apply install failed: stdout=%q err=%v", stdout, err)
	}
	installed, err := os.ReadFile(target)
	if err != nil || !strings.Contains(string(installed), managedHookBegin) || !strings.Contains(string(installed), " --async") {
		t.Fatalf("unexpected installed target: %q err=%v", installed, err)
	}
	backup, err := os.ReadFile(target + ".pupdate.bak")
	if err != nil || !bytes.Equal(backup, original) {
		t.Fatalf("unexpected backup: %q err=%v", backup, err)
	}
	info, err := os.Stat(target)
	if err != nil || info.Mode().Perm() != 0o640 {
		t.Fatalf("target mode changed: mode=%v err=%v", info.Mode().Perm(), err)
	}

	stdout, _, err = executeInitManage(t, "install", "--shell", "zsh", "--target", target, "--apply")
	if err != nil || !strings.Contains(stdout, "no pupdate hook change needed") {
		t.Fatalf("idempotent reinstall failed: stdout=%q err=%v", stdout, err)
	}
	again, _ := os.ReadFile(target)
	if !bytes.Equal(again, installed) {
		t.Fatalf("idempotent reinstall changed bytes")
	}
}

func TestInitInstallUpdatesModeWithoutOverwritingBackup(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, ".bashrc")
	if err := os.WriteFile(target, []byte("original\n"), 0o600); err != nil {
		t.Fatalf("write target: %v", err)
	}
	if _, _, err := executeInitManage(t, "install", "--shell", "bash", "--target", target, "--apply"); err != nil {
		t.Fatalf("initial install: %v", err)
	}
	backupPath := target + ".pupdate.bak"
	if err := os.WriteFile(backupPath, []byte("existing backup\n"), 0o600); err != nil {
		t.Fatalf("replace backup sentinel: %v", err)
	}
	if _, _, err := executeInitManage(t, "install", "--shell", "bash", "--mode", "foreground", "--target", target, "--apply"); err != nil {
		t.Fatalf("mode update: %v", err)
	}
	raw, _ := os.ReadFile(target)
	if strings.Contains(string(raw), " --async") {
		t.Fatalf("foreground update retained async mode: %q", raw)
	}
	backup, _ := os.ReadFile(backupPath)
	if string(backup) != "existing backup\n" {
		t.Fatalf("existing backup was overwritten: %q", backup)
	}
}

func TestInitUninstallRemovesOnlyManagedBlockAndNeverRestoresBackup(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, ".bashrc")
	if err := os.WriteFile(target, []byte("before\n"), 0o600); err != nil {
		t.Fatalf("write target: %v", err)
	}
	if _, _, err := executeInitManage(t, "install", "--shell", "bash", "--target", target, "--apply"); err != nil {
		t.Fatalf("install: %v", err)
	}
	installed, _ := os.ReadFile(target)
	installed = append(installed, []byte("after\n")...)
	if err := os.WriteFile(target, installed, 0o600); err != nil {
		t.Fatalf("append user content: %v", err)
	}
	backupPath := target + ".pupdate.bak"
	if err := os.WriteFile(backupPath, []byte("do not restore\n"), 0o600); err != nil {
		t.Fatalf("write backup sentinel: %v", err)
	}

	if _, _, err := executeInitManage(t, "uninstall", "--shell", "bash", "--target", target, "--apply"); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	raw, _ := os.ReadFile(target)
	if strings.Contains(string(raw), managedHookBegin) || !strings.Contains(string(raw), "before\n") || !strings.Contains(string(raw), "after\n") {
		t.Fatalf("uninstall changed content outside managed block: %q", raw)
	}
	if string(raw) == "do not restore\n" {
		t.Fatal("uninstall restored backup")
	}
	backup, _ := os.ReadFile(backupPath)
	if string(backup) != "do not restore\n" {
		t.Fatalf("uninstall changed backup: %q", backup)
	}

	stdout, _, err := executeInitManage(t, "uninstall", "--shell", "bash", "--target", filepath.Join(dir, "missing"), "--apply")
	if err != nil || !strings.Contains(stdout, "no pupdate hook change needed") {
		t.Fatalf("missing-block uninstall should be a no-op: stdout=%q err=%v", stdout, err)
	}
}

func TestInitManagementRejectsMalformedMarkers(t *testing.T) {
	tests := []string{
		managedHookBegin + "\nmissing end\n",
		managedHookEnd + "\nmissing begin\n",
		managedHookBegin + "\n" + managedHookBegin + "\n" + managedHookEnd + "\n" + managedHookEnd + "\n",
		managedHookEnd + "\n" + managedHookBegin + "\n",
		"echo '" + managedHookBegin + "'\n" + managedHookEnd + "\n",
	}
	for index, content := range tests {
		t.Run(string(rune('a'+index)), func(t *testing.T) {
			target := filepath.Join(t.TempDir(), ".bashrc")
			if err := os.WriteFile(target, []byte(content), 0o600); err != nil {
				t.Fatalf("write malformed target: %v", err)
			}
			if _, _, err := executeInitManage(t, "install", "--shell", "bash", "--target", target, "--apply"); err == nil {
				t.Fatal("expected malformed marker error")
			}
		})
	}
}

func TestInitInstallCreatesMissingTargetWithPrivatePermissions(t *testing.T) {
	target := filepath.Join(t.TempDir(), "nested", "config.fish")
	if _, _, err := executeInitManage(t, "install", "--shell", "fish", "--target", target, "--apply"); err != nil {
		t.Fatalf("install missing target: %v", err)
	}
	info, err := os.Stat(target)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("new target mode: mode=%v err=%v", info.Mode().Perm(), err)
	}
	if _, err := os.Stat(target + ".pupdate.bak"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("new target should not get an empty backup, err=%v", err)
	}
}

func TestInitInstallResolvesSymlinkedStartupFile(t *testing.T) {
	dir := t.TempDir()
	realTarget := filepath.Join(dir, "real-bashrc")
	symlinkTarget := filepath.Join(dir, ".bashrc")
	if err := os.WriteFile(realTarget, []byte("real\n"), 0o600); err != nil {
		t.Fatalf("write real target: %v", err)
	}
	if err := os.Symlink(realTarget, symlinkTarget); err != nil {
		t.Fatalf("create symlink: %v", err)
	}
	stdout, _, err := executeInitManage(t, "install", "--shell", "bash", "--target", symlinkTarget, "--apply")
	if err != nil || !strings.Contains(stdout, realTarget) {
		t.Fatalf("symlink install failed: stdout=%q err=%v", stdout, err)
	}
	if info, err := os.Lstat(symlinkTarget); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("startup symlink was replaced: mode=%v err=%v", info.Mode(), err)
	}
	raw, _ := os.ReadFile(realTarget)
	if !strings.Contains(string(raw), managedHookBegin) {
		t.Fatalf("resolved target was not updated: %q", raw)
	}
}

func TestDefaultManagedHookTargets(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("ZDOTDIR", filepath.Join(home, "zdot"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "xdg"))
	tests := map[string]string{
		"bash": filepath.Join(home, ".bashrc"),
		"zsh":  filepath.Join(home, "zdot", ".zshrc"),
		"fish": filepath.Join(home, "xdg", "fish", "config.fish"),
	}
	for shell, want := range tests {
		got, err := defaultManagedHookTarget(shell)
		if err != nil || got != want {
			t.Fatalf("default %s target: got=%q want=%q err=%v", shell, got, want, err)
		}
	}
}
