package freshness

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/state"
)

var gitSubmoduleStatusFn = defaultGitSubmoduleStatus

type Decision string

const (
	DecisionUpdate       Decision = "update"
	DecisionSkip         Decision = "skip"
	phpVendorChecksumKey          = "vendor/.pupdate-checksum"
)

type EcosystemDecision struct {
	Ecosystem string
	StateKey  string
	Decision  Decision
	Reason    string
	Lockfiles map[string]string
}

func Evaluate(dir string, detections []detection.DetectionResult, current state.FileState) ([]EcosystemDecision, error) {
	decisions := make([]EcosystemDecision, 0, len(detections))

	for _, result := range detections {
		lockfiles, err := hashMatchedFiles(dir, result.MatchedFiles)
		if err != nil {
			return nil, err
		}

		ecosystem := string(result.Ecosystem)
		stateKey := result.StateKey()
		ecosystemState, hasState := current.Ecosystems[stateKey]
		if !hasState && (result.Directory == "" || result.Directory == ".") {
			ecosystemState, hasState = current.Ecosystems[ecosystem]
		}

		if result.Ecosystem == detection.EcosystemPHP {
			shouldTrackVendor, err := shouldTrackPHPVendor(dir, result, ecosystemState, hasState)
			if err != nil {
				return nil, err
			}
			if shouldTrackVendor {
				checksum, err := hashPHPVendorChecksum(dir, result)
				if err != nil {
					return nil, err
				}
				lockfiles[phpVendorChecksumKey] = checksum
			}
		}
		decision := EcosystemDecision{
			Ecosystem: ecosystem,
			StateKey:  stateKey,
			Decision:  DecisionUpdate,
			Reason:    "missing prior lockfile hash",
			Lockfiles: lockfiles,
		}
		if hasState && len(ecosystemState.Lockfiles) > 0 {
			if lockfilesEqual(ecosystemState.Lockfiles, lockfiles) {
				decision.Decision = DecisionSkip
				decision.Reason = "dependency lockfiles unchanged since last successful run"
			} else {
				decision.Reason = "dependency lockfiles changed since last successful run"
			}
		}

		if result.Ecosystem == detection.EcosystemGit {
			lines, err := gitSubmoduleStatusFn(dir)
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

func shouldTrackPHPVendor(dir string, result detection.DetectionResult, ecosystemState state.EcosystemState, hasState bool) (bool, error) {
	vendorPath := phpVendorPath(dir, result)
	vendorExists, err := isDirectory(vendorPath)
	if err != nil {
		return false, fmt.Errorf("stat php vendor directory: %w", err)
	}
	if vendorExists {
		return true, nil
	}

	if !hasState {
		return false, nil
	}

	_, trackedVendor := ecosystemState.Lockfiles[phpVendorChecksumKey]
	return trackedVendor, nil
}

func hashPHPVendorChecksum(dir string, result detection.DetectionResult) (string, error) {
	hasher := sha256.New()
	trackedPaths := []string{
		"vendor",
		"vendor/autoload.php",
		"vendor/composer/installed.json",
		"vendor/composer/installed.php",
	}

	for _, rel := range trackedPaths {
		fullPath := filepath.Join(phpProjectPath(dir, result), filepath.FromSlash(rel))
		if err := writeFingerprintEntry(hasher, rel, fullPath); err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func writeFingerprintEntry(hasher io.Writer, relPath string, fullPath string) error {
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			_, writeErr := io.WriteString(hasher, relPath+"|missing\n")
			return writeErr
		}
		return fmt.Errorf("stat %q: %w", relPath, err)
	}

	if _, err := io.WriteString(hasher, fmt.Sprintf("%s|%d|%d|%s\n", relPath, info.Size(), info.ModTime().UTC().UnixNano(), info.Mode().String())); err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}

	contentHash, err := hashFile(fullPath)
	if err != nil {
		return fmt.Errorf("hash %q: %w", relPath, err)
	}
	_, err = io.WriteString(hasher, contentHash+"\n")
	return err
}

func phpVendorPath(dir string, result detection.DetectionResult) string {
	return filepath.Join(phpProjectPath(dir, result), "vendor")
}

func phpProjectPath(dir string, result detection.DetectionResult) string {
	if result.Directory == "" || result.Directory == "." {
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

func defaultGitSubmoduleStatus(dir string) ([]string, error) {
	cmd := exec.Command("git", "submodule", "status", "--recursive")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		var stderr []byte
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = exitErr.Stderr
		}
		trimmed := strings.TrimSpace(string(bytes.TrimSpace(stderr)))
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

func hashMatchedFiles(dir string, matchedFiles []string) (map[string]string, error) {
	lockfiles := make(map[string]string, len(matchedFiles))
	for _, matchedFile := range matchedFiles {
		fullPath := filepath.Join(dir, matchedFile)
		hash, err := hashFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("hash matched file %q: %w", matchedFile, err)
		}

		lockfiles[strings.ToLower(matchedFile)] = hash
	}

	return lockfiles, nil
}

func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func lockfilesEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}

	keys := make([]string, 0, len(a))
	for key := range a {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	for _, key := range keys {
		if a[key] != b[key] {
			return false
		}
	}

	return true
}
