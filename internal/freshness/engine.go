package freshness

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/state"
)

var gitSubmoduleStatusFn = defaultGitSubmoduleStatus

var gitSubmoduleStatusTimeout = 2 * time.Second

var runGitSubmoduleStatusCommand = func(ctx context.Context, dir string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, "git", "submodule", "status", "--recursive")
	cmd.Dir = dir
	output, err := cmd.Output()
	var stderr []byte
	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr = exitErr.Stderr
	}
	return output, stderr, err
}

type Decision string

const (
	DecisionUpdate       Decision = "update"
	DecisionSkip         Decision = "skip"
	phpVendorChecksumKey          = "vendor/.pupdate-checksum"
)

type EcosystemDecision struct {
	Ecosystem        string
	StateKey         string
	Decision         Decision
	Reason           string
	Lockfiles        map[string]string
	LockfileMetadata map[string]state.LockfileMetadata
}

func Evaluate(dir string, detections []detection.DetectionResult, current state.FileState) ([]EcosystemDecision, error) {
	return EvaluateContext(context.Background(), dir, detections, current)
}

func EvaluateContext(ctx context.Context, dir string, detections []detection.DetectionResult, current state.FileState) ([]EcosystemDecision, error) {
	decisions := make([]EcosystemDecision, 0, len(detections))

	for _, result := range detections {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		ecosystem := string(result.Ecosystem)
		stateKey := result.StateKey()
		ecosystemState, hasState := current.Ecosystems[stateKey]
		if !hasState && result.IsRoot() {
			ecosystemState, hasState = current.Ecosystems[ecosystem]
		}

		snapshot, err := SnapshotMatchedFiles(ctx, dir, result.MatchedFiles, ecosystemState)
		if err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		lockfiles := snapshot.Lockfiles
		lockfileMetadata := snapshot.LockfileMetadata

		storedLockfiles := comparableLockfiles(result.Ecosystem, ecosystemState.Lockfiles)
		decision := EcosystemDecision{
			Ecosystem:        ecosystem,
			StateKey:         stateKey,
			Decision:         DecisionUpdate,
			Reason:           "missing prior lockfile hash",
			Lockfiles:        lockfiles,
			LockfileMetadata: lockfileMetadata,
		}
		legacyPHPVendorTracked := result.Ecosystem == detection.EcosystemPHP && len(storedLockfiles) != len(ecosystemState.Lockfiles)
		if hasState && len(storedLockfiles) > 0 {
			if lockfilesEqual(storedLockfiles, lockfiles) {
				decision.Decision = DecisionSkip
				decision.Reason = "dependency lockfiles unchanged since last successful run"
				if legacyPHPVendorTracked {
					vendorExists, err := isDirectory(phpVendorPath(dir, result))
					if err != nil {
						return nil, fmt.Errorf("stat php vendor directory: %w", err)
					}
					if !vendorExists {
						decision.Decision = DecisionUpdate
						decision.Reason = "composer vendor directory missing since last successful run"
					}
				}
			} else {
				decision.Reason = "dependency lockfiles changed since last successful run"
			}
		}

		if result.Ecosystem == detection.EcosystemGit {
			gitDirectory := dir
			if !result.IsRoot() {
				gitDirectory = filepath.Join(dir, result.Directory)
			}
			lines, err := gitSubmoduleStatusFn(ctx, gitDirectory)
			if cancellationErr := ctx.Err(); cancellationErr != nil {
				return nil, cancellationErr
			}
			if err != nil {
				decision.Decision = DecisionSkip
				decision.Reason = fmt.Sprintf("git submodule status failed: %v", err)
			} else if hasGitSubmoduleDrift(lines) {
				decision.Decision = DecisionUpdate
				decision.Reason = "git submodule state drifted from recorded revision"
			}
		}

		decisions = append(decisions, decision)
	}

	return decisions, nil
}

func phpVendorPath(dir string, result detection.DetectionResult) string {
	return filepath.Join(phpProjectPath(dir, result), "vendor")
}

func comparableLockfiles(ecosystem detection.Ecosystem, lockfiles map[string]string) map[string]string {
	if ecosystem != detection.EcosystemPHP {
		return lockfiles
	}
	if len(lockfiles) == 0 {
		return lockfiles
	}
	if _, ok := lockfiles[phpVendorChecksumKey]; !ok {
		return lockfiles
	}

	trimmed := make(map[string]string, len(lockfiles)-1)
	for key, value := range lockfiles {
		if key == phpVendorChecksumKey {
			continue
		}
		trimmed[key] = value
	}
	return trimmed
}

func phpProjectPath(dir string, result detection.DetectionResult) string {
	if result.IsRoot() {
		return dir
	}
	return filepath.Join(dir, result.Directory)
}

func isDirectory(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

func defaultGitSubmoduleStatus(parent context.Context, dir string) ([]string, error) {
	ctx, cancel := context.WithTimeout(parent, gitSubmoduleStatusTimeout)
	defer cancel()

	output, stderr, err := runGitSubmoduleStatusCommand(ctx, dir)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("timed out after %s", gitSubmoduleStatusTimeout)
		}
		trimmed := strings.TrimSpace(string(stderr))
		if trimmed != "" {
			return nil, fmt.Errorf("%w: %s", err, trimmed)
		}
		return nil, err
	}

	lines := []string{}
	for _, line := range strings.Split(string(output), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines, nil
}

func hasGitSubmoduleDrift(lines []string) bool {
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		switch trimmed[0] {
		case '-', '+', 'U':
			return true
		}
	}
	return false
}

var hashFileFn = hashFileContext

type MatchedFilesSnapshot struct {
	Lockfiles        map[string]string
	LockfileMetadata map[string]state.LockfileMetadata
}

func SnapshotMatchedFiles(ctx context.Context, dir string, matchedFiles []string, previous state.EcosystemState) (MatchedFilesSnapshot, error) {
	lockfiles, metadata, err := snapshotMatchedFiles(ctx, dir, matchedFiles, previous)
	if err != nil {
		return MatchedFilesSnapshot{}, err
	}
	return MatchedFilesSnapshot{Lockfiles: lockfiles, LockfileMetadata: metadata}, nil
}

func hashMatchedFiles(dir string, matchedFiles []string, previous state.EcosystemState) (map[string]string, map[string]state.LockfileMetadata, error) {
	return snapshotMatchedFiles(context.Background(), dir, matchedFiles, previous)
}

func snapshotMatchedFiles(ctx context.Context, dir string, matchedFiles []string, previous state.EcosystemState) (map[string]string, map[string]state.LockfileMetadata, error) {
	lockfiles := make(map[string]string, len(matchedFiles))
	metadata := make(map[string]state.LockfileMetadata, len(matchedFiles))
	for _, matchedFile := range matchedFiles {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		fullPath := filepath.Join(dir, matchedFile)
		info, err := os.Stat(fullPath)
		if err != nil {
			return nil, nil, fmt.Errorf("stat matched file %q: %w", matchedFile, err)
		}

		key := strings.ToLower(matchedFile)
		currentMetadata := state.LockfileMetadata{
			Size:            info.Size(),
			ModTimeUnixNano: info.ModTime().UTC().UnixNano(),
			Mode:            info.Mode().String(),
		}
		enrichLockfileMetadata(info, &currentMetadata)
		metadata[key] = currentMetadata

		if canReuseStoredLockfileHash(currentMetadata, previous.LockfileMetadata[key], previous.Lockfiles[key]) {
			lockfiles[key] = previous.Lockfiles[key]
			continue
		}

		hash, err := hashFileFn(ctx, fullPath)
		if err != nil {
			return nil, nil, fmt.Errorf("hash matched file %q: %w", matchedFile, err)
		}

		lockfiles[key] = hash
	}

	return lockfiles, metadata, nil
}

func canReuseStoredLockfileHash(current state.LockfileMetadata, previous state.LockfileMetadata, previousHash string) bool {
	if previousHash == "" {
		return false
	}
	if current.Size != previous.Size || current.ModTimeUnixNano != previous.ModTimeUnixNano || current.Mode != previous.Mode {
		return false
	}
	if current.FileID == "" || current.ChangeTimeUnixNano == 0 {
		return false
	}
	return current.FileID == previous.FileID && current.ChangeTimeUnixNano == previous.ChangeTimeUnixNano
}

func hashFile(path string) (string, error) {
	return hashFileContext(context.Background(), path)
}

func hashFileContext(ctx context.Context, path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	return hashReader(ctx, file)
}

func hashReader(ctx context.Context, reader io.Reader) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, contextReader{Context: ctx, Reader: reader}); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

type contextReader struct {
	context.Context
	io.Reader
}

func (r contextReader) Read(buffer []byte) (int, error) {
	if err := r.Context.Err(); err != nil {
		return 0, err
	}
	return r.Reader.Read(buffer)
}

func lockfilesEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}

	for key, value := range a {
		other, ok := b[key]
		if !ok || other != value {
			return false
		}
	}

	return true
}
