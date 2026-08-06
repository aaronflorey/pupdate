package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const (
	managedHookBegin = "# >>> pupdate shell hook >>>"
	managedHookEnd   = "# <<< pupdate shell hook <<<"
)

type managedHookOperation string

const (
	managedHookInstall   managedHookOperation = "install"
	managedHookUninstall managedHookOperation = "uninstall"
)

type managedHookChange struct {
	Target    string
	Operation managedHookOperation
	Content   []byte
	Mode      fs.FileMode
	Existed   bool
	Changed   bool
}

func newInitInstallCmd() *cobra.Command {
	var shell string
	var mode string
	var target string
	var apply bool
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Preview or install a persistent shell hook",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			resolvedShell, err := resolveShell(shell)
			if err != nil {
				return err
			}
			resolvedMode, err := resolveHookMode(mode)
			if err != nil {
				return err
			}
			resolvedTarget, err := resolveManagedHookTarget(resolvedShell, target)
			if err != nil {
				return err
			}
			snippet, err := initSnippetForShell(resolvedShell, resolvedMode)
			if err != nil {
				return err
			}
			change, err := planManagedHookChange(resolvedTarget, managedHookInstall, snippet)
			if err != nil {
				return err
			}
			if !apply {
				return writeManagedHookPreview(cmd, change)
			}
			if err := applyManagedHookChange(change); err != nil {
				return err
			}
			return writeManagedHookResult(cmd, change)
		},
	}
	cmd.Flags().StringVar(&shell, "shell", "", "shell to configure (bash, zsh, or fish)")
	cmd.Flags().StringVar(&mode, "mode", hookModeAsync, "hook execution mode (async or foreground)")
	cmd.Flags().StringVar(&target, "target", "", "startup file to manage")
	cmd.Flags().BoolVar(&apply, "apply", false, "apply the proposed startup-file change")
	return cmd
}

func newInitUninstallCmd() *cobra.Command {
	var shell string
	var target string
	var apply bool
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Preview or remove a persistent shell hook",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			resolvedShell, err := resolveShell(shell)
			if err != nil {
				return err
			}
			resolvedTarget, err := resolveManagedHookTarget(resolvedShell, target)
			if err != nil {
				return err
			}
			change, err := planManagedHookChange(resolvedTarget, managedHookUninstall, "")
			if err != nil {
				return err
			}
			if !apply {
				return writeManagedHookPreview(cmd, change)
			}
			if err := applyManagedHookChange(change); err != nil {
				return err
			}
			return writeManagedHookResult(cmd, change)
		},
	}
	cmd.Flags().StringVar(&shell, "shell", "", "shell to configure (bash, zsh, or fish)")
	cmd.Flags().StringVar(&target, "target", "", "startup file to manage")
	cmd.Flags().BoolVar(&apply, "apply", false, "apply the proposed startup-file change")
	return cmd
}

func resolveManagedHookTarget(shell string, requested string) (string, error) {
	target := requested
	if target == "" {
		var err error
		target, err = defaultManagedHookTarget(shell)
		if err != nil {
			return "", err
		}
	}

	expanded, err := expandManagedHookPath(target)
	if err != nil {
		return "", fmt.Errorf("failed to resolve shell hook target %q: %w", target, err)
	}
	resolved, err := filepath.EvalSymlinks(expanded)
	if err == nil {
		return resolved, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("failed to resolve shell hook target %s: %w", expanded, err)
	}

	parent := filepath.Dir(expanded)
	resolvedParent, parentErr := filepath.EvalSymlinks(parent)
	if parentErr == nil {
		return filepath.Join(resolvedParent, filepath.Base(expanded)), nil
	}
	if !errors.Is(parentErr, os.ErrNotExist) {
		return "", fmt.Errorf("failed to resolve shell hook target parent %s: %w", parent, parentErr)
	}
	return expanded, nil
}

func defaultManagedHookTarget(shell string) (string, error) {
	homeCandidates := homeDirectoryCandidates()
	home := ""
	if len(homeCandidates) > 0 {
		home = homeCandidates[0]
	}
	switch shell {
	case "bash":
		if home == "" {
			return "", fmt.Errorf("failed to resolve user home directory")
		}
		return filepath.Join(home, ".bashrc"), nil
	case "zsh":
		if zdotdir := os.Getenv("ZDOTDIR"); zdotdir != "" {
			return filepath.Join(zdotdir, ".zshrc"), nil
		}
		if home == "" {
			return "", fmt.Errorf("failed to resolve user home directory")
		}
		return filepath.Join(home, ".zshrc"), nil
	case "fish":
		if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
			return filepath.Join(xdgConfigHome, "fish", "config.fish"), nil
		}
		if home == "" {
			return "", fmt.Errorf("failed to resolve user home directory")
		}
		return filepath.Join(home, ".config", "fish", "config.fish"), nil
	default:
		return "", fmt.Errorf("unsupported shell %q", shell)
	}
}

func expandManagedHookPath(path string) (string, error) {
	expanded, err := expandHomeTilde(path)
	if err != nil {
		return "", err
	}
	if filepath.IsAbs(expanded) {
		return filepath.Clean(expanded), nil
	}
	return filepath.Abs(expanded)
}

func planManagedHookChange(target string, operation managedHookOperation, snippet string) (managedHookChange, error) {
	content, info, err := readManagedHookTarget(target)
	if err != nil {
		return managedHookChange{}, err
	}
	start, end, found, err := managedHookRange(string(content))
	if err != nil {
		return managedHookChange{}, fmt.Errorf("invalid pupdate hook markers in %s: %w", target, err)
	}

	next := append([]byte(nil), content...)
	switch operation {
	case managedHookInstall:
		block := managedHookBlock(snippet)
		if found {
			next = []byte(string(content[:start]) + block + string(content[end:]))
		} else {
			prefix := string(content)
			if prefix != "" && !strings.HasSuffix(prefix, "\n") {
				prefix += "\n"
			}
			if prefix != "" && !strings.HasSuffix(prefix, "\n\n") {
				prefix += "\n"
			}
			next = []byte(prefix + block)
		}
	case managedHookUninstall:
		if found {
			next = append(append([]byte(nil), content[:start]...), content[end:]...)
		}
	default:
		return managedHookChange{}, fmt.Errorf("unsupported managed hook operation %q", operation)
	}

	mode := fs.FileMode(0o600)
	existed := info != nil
	if existed {
		mode = info.Mode().Perm()
	}
	return managedHookChange{
		Target:    target,
		Operation: operation,
		Content:   next,
		Mode:      mode,
		Existed:   existed,
		Changed:   !bytesEqual(content, next),
	}, nil
}

func readManagedHookTarget(target string) ([]byte, fs.FileInfo, error) {
	info, err := os.Stat(target)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to stat shell hook target %s: %w", target, err)
	}
	if !info.Mode().IsRegular() {
		return nil, nil, fmt.Errorf("shell hook target %s is not a regular file", target)
	}
	content, err := os.ReadFile(target)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read shell hook target %s: %w", target, err)
	}
	return content, info, nil
}

func managedHookBlock(snippet string) string {
	return managedHookBegin + "\n" + strings.TrimSuffix(snippet, "\n") + "\n" + managedHookEnd + "\n"
}

func managedHookRange(content string) (int, int, bool, error) {
	beginCount := strings.Count(content, managedHookBegin)
	endCount := strings.Count(content, managedHookEnd)
	if beginCount == 0 && endCount == 0 {
		return 0, 0, false, nil
	}
	if beginCount != 1 || endCount != 1 {
		return 0, 0, false, fmt.Errorf("expected exactly one matched marker pair")
	}
	start := strings.Index(content, managedHookBegin)
	endMarker := strings.Index(content, managedHookEnd)
	if endMarker < start {
		return 0, 0, false, fmt.Errorf("closing marker appears before opening marker")
	}
	if !markerOccupiesLine(content, start, managedHookBegin) || !markerOccupiesLine(content, endMarker, managedHookEnd) {
		return 0, 0, false, fmt.Errorf("markers must each occupy a complete line")
	}
	end := endMarker + len(managedHookEnd)
	if end+1 < len(content) && content[end] == '\r' && content[end+1] == '\n' {
		end += 2
	} else if end < len(content) && content[end] == '\n' {
		end++
	}
	return start, end, true, nil
}

func markerOccupiesLine(content string, start int, marker string) bool {
	if start > 0 && content[start-1] != '\n' {
		return false
	}
	end := start + len(marker)
	if end == len(content) || content[end] == '\n' {
		return true
	}
	return end+1 < len(content) && content[end] == '\r' && content[end+1] == '\n'
}

func applyManagedHookChange(change managedHookChange) error {
	if !change.Changed {
		return nil
	}
	parent := filepath.Dir(change.Target)
	if err := os.MkdirAll(parent, 0o700); err != nil {
		return fmt.Errorf("failed to create shell hook target directory %s: %w", parent, err)
	}
	if change.Existed {
		if err := createManagedHookBackup(change.Target, change.Mode); err != nil {
			return err
		}
	}
	if err := atomicWriteManagedHook(change.Target, change.Content, change.Mode); err != nil {
		return err
	}
	return nil
}

func createManagedHookBackup(target string, mode fs.FileMode) error {
	backupPath := target + ".pupdate.bak"
	content, err := os.ReadFile(target)
	if err != nil {
		return fmt.Errorf("failed to read shell hook target for backup %s: %w", target, err)
	}
	backup, err := os.OpenFile(backupPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode.Perm())
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to create shell hook backup %s: %w", backupPath, err)
	}
	_, writeErr := backup.Write(content)
	if writeErr == nil {
		writeErr = backup.Sync()
	}
	closeErr := backup.Close()
	if writeErr != nil {
		return fmt.Errorf("failed to write shell hook backup %s: %w", backupPath, writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("failed to close shell hook backup %s: %w", backupPath, closeErr)
	}
	return syncManagedHookParent(backupPath)
}

func atomicWriteManagedHook(target string, content []byte, mode fs.FileMode) error {
	temporary, err := os.CreateTemp(filepath.Dir(target), ".pupdate-hook-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary shell hook file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(mode.Perm()); err != nil {
		temporary.Close()
		return fmt.Errorf("failed to set temporary shell hook permissions: %w", err)
	}
	if _, err := temporary.Write(content); err != nil {
		temporary.Close()
		return fmt.Errorf("failed to write temporary shell hook file: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return fmt.Errorf("failed to sync temporary shell hook file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("failed to close temporary shell hook file: %w", err)
	}
	if err := os.Rename(temporaryPath, target); err != nil {
		return fmt.Errorf("failed to replace shell hook target %s: %w", target, err)
	}
	return syncManagedHookParent(target)
}

func syncManagedHookParent(path string) error {
	directory, err := os.Open(filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("failed to open shell hook target directory: %w", err)
	}
	defer directory.Close()
	if err := directory.Sync(); err != nil {
		return fmt.Errorf("failed to sync shell hook target directory: %w", err)
	}
	return nil
}

func writeManagedHookPreview(cmd *cobra.Command, change managedHookChange) error {
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "target: %s\noperation: %s\nchange: %s\napply: false\n",
		safeDisplay(change.Target), change.Operation, managedHookChangeLabel(change))
	return err
}

func writeManagedHookResult(cmd *cobra.Command, change managedHookChange) error {
	if !change.Changed {
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "no pupdate hook change needed in %s\n", safeDisplay(change.Target))
		return err
	}
	verb := "installed"
	if change.Operation == managedHookUninstall {
		verb = "removed"
	}
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s pupdate hook in %s\n", verb, safeDisplay(change.Target))
	return err
}

func managedHookChangeLabel(change managedHookChange) string {
	if !change.Changed {
		return "none"
	}
	if change.Operation == managedHookUninstall {
		return "remove managed block"
	}
	if change.Existed {
		return "update managed block"
	}
	return "create startup file"
}

func bytesEqual(left []byte, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
