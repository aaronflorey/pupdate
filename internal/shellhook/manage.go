package shellhook

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/aaronflorey/pupdate/internal/userenv"
)

const (
	ManagedBegin = "# >>> pupdate shell hook >>>"
	ManagedEnd   = "# <<< pupdate shell hook <<<"
)

type Operation string

const (
	Install   Operation = "install"
	Uninstall Operation = "uninstall"
)

type Change struct {
	Target    string
	Operation Operation
	Content   []byte
	Mode      fs.FileMode
	Existed   bool
	Changed   bool
}

func ResolveTarget(shell string, requested string) (string, error) {
	target := requested
	if target == "" {
		var err error
		target, err = DefaultTarget(shell)
		if err != nil {
			return "", err
		}
	}

	expanded, err := ExpandPath(target)
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

func DefaultTarget(shell string) (string, error) {
	homeCandidates := userenv.System().HomeDirectories()
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

func ExpandPath(path string) (string, error) {
	expanded, err := userenv.System().ExpandHome(path)
	if err != nil {
		return "", err
	}
	if filepath.IsAbs(expanded) {
		return filepath.Clean(expanded), nil
	}
	return filepath.Abs(expanded)
}

func PlanChange(target string, operation Operation, snippet string) (Change, error) {
	content, info, err := readManagedHookTarget(target)
	if err != nil {
		return Change{}, err
	}
	start, end, found, err := managedHookRange(string(content))
	if err != nil {
		return Change{}, fmt.Errorf("invalid pupdate hook markers in %s: %w", target, err)
	}

	next := append([]byte(nil), content...)
	switch operation {
	case Install:
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
	case Uninstall:
		if found {
			next = append(append([]byte(nil), content[:start]...), content[end:]...)
		}
	default:
		return Change{}, fmt.Errorf("unsupported managed hook operation %q", operation)
	}

	mode := fs.FileMode(0o600)
	existed := info != nil
	if existed {
		mode = info.Mode().Perm()
	}
	return Change{
		Target:    target,
		Operation: operation,
		Content:   next,
		Mode:      mode,
		Existed:   existed,
		Changed:   !bytes.Equal(content, next),
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
	return ManagedBegin + "\n" + strings.TrimSuffix(snippet, "\n") + "\n" + ManagedEnd + "\n"
}

func managedHookRange(content string) (int, int, bool, error) {
	beginCount := strings.Count(content, ManagedBegin)
	endCount := strings.Count(content, ManagedEnd)
	if beginCount == 0 && endCount == 0 {
		return 0, 0, false, nil
	}
	if beginCount != 1 || endCount != 1 {
		return 0, 0, false, fmt.Errorf("expected exactly one matched marker pair")
	}
	start := strings.Index(content, ManagedBegin)
	endMarker := strings.Index(content, ManagedEnd)
	if endMarker < start {
		return 0, 0, false, fmt.Errorf("closing marker appears before opening marker")
	}
	if !markerOccupiesLine(content, start, ManagedBegin) || !markerOccupiesLine(content, endMarker, ManagedEnd) {
		return 0, 0, false, fmt.Errorf("markers must each occupy a complete line")
	}
	end := endMarker + len(ManagedEnd)
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

func ApplyChange(change Change) error {
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

func ChangeLabel(change Change) string {
	if !change.Changed {
		return "none"
	}
	if change.Operation == Uninstall {
		return "remove managed block"
	}
	if change.Existed {
		return "update managed block"
	}
	return "create startup file"
}
