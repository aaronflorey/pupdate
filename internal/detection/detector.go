package detection

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

type ignoreMatcher interface {
	Match(path []string, isDir bool) bool
}

func Detect(dir string) ([]DetectionResult, error) {
	directories, err := scanDirectories(dir)
	if err != nil {
		return nil, err
	}

	results := []DetectionResult{}
	for _, directory := range directories {
		dirPath := dir
		if directory != "." {
			dirPath = filepath.Join(dir, directory)
		}

		directoryResults, err := detectDirectory(dirPath, directory)
		if err != nil {
			return nil, err
		}
		results = append(results, directoryResults...)
	}

	return results, nil
}

func scanDirectories(dir string) ([]string, error) {
	matcher, err := loadIgnoreMatcher(dir)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	directories := []string{"."}
	for _, entry := range entries {
		if !entry.IsDir() || (entry.Type()&os.ModeSymlink) != 0 {
			continue
		}

		relPath := entry.Name()
		if shouldSkipDirectory(matcher, relPath) {
			continue
		}

		directories = append(directories, relPath)

		if relPath != "packages" {
			continue
		}

		packageEntries, err := os.ReadDir(filepath.Join(dir, relPath))
		if err != nil {
			return nil, err
		}

		for _, packageEntry := range packageEntries {
			if !packageEntry.IsDir() || (packageEntry.Type()&os.ModeSymlink) != 0 {
				continue
			}

			packagePath := filepath.ToSlash(filepath.Join(relPath, packageEntry.Name()))
			if shouldSkipDirectory(matcher, packagePath) {
				continue
			}

			directories = append(directories, packagePath)
		}
	}
	slices.Sort(directories[1:])
	return directories, nil
}

func loadIgnoreMatcher(dir string) (ignoreMatcher, error) {
	ignorePath := filepath.Join(dir, ".gitignore")
	raw, err := os.ReadFile(ignorePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	lines := strings.Split(string(raw), "\n")
	patterns := make([]gitignore.Pattern, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		if line == "" {
			continue
		}
		patterns = append(patterns, gitignore.ParsePattern(line, nil))
	}

	if len(patterns) == 0 {
		return nil, nil
	}

	return gitignore.NewMatcher(patterns), nil
}

func shouldSkipDirectory(matcher ignoreMatcher, path string) bool {
	if matcher == nil {
		return false
	}

	parts := strings.Split(filepath.ToSlash(path), "/")
	return matcher.Match(parts, true)
}

func detectDirectory(dirPath string, relativeDir string) ([]DetectionResult, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	filesByEcosystem := map[Ecosystem][]string{
		EcosystemNode:   {},
		EcosystemPHP:    {},
		EcosystemGo:     {},
		EcosystemRust:   {},
		EcosystemPython: {},
		EcosystemGit:    {},
	}
	managersByEcosystem := map[Ecosystem][]string{
		EcosystemNode:   {},
		EcosystemPHP:    {},
		EcosystemGo:     {},
		EcosystemRust:   {},
		EcosystemPython: {},
		EcosystemGit:    {},
	}
	fileSeen := map[Ecosystem]map[string]bool{
		EcosystemNode:   {},
		EcosystemPHP:    {},
		EcosystemGo:     {},
		EcosystemRust:   {},
		EcosystemPython: {},
		EcosystemGit:    {},
	}
	managerSeen := map[Ecosystem]map[string]bool{
		EcosystemNode:   {},
		EcosystemPHP:    {},
		EcosystemGo:     {},
		EcosystemRust:   {},
		EcosystemPython: {},
		EcosystemGit:    {},
	}

	for _, entry := range entries {
		if entry.IsDir() || (entry.Type()&os.ModeSymlink) != 0 {
			continue
		}

		signal, ok := detectSupportedSignal(filepath.Base(entry.Name()))
		if !ok {
			continue
		}
		name := signal.name
		ecosystem := signal.ecosystem

		if !fileSeen[ecosystem][name] {
			fileSeen[ecosystem][name] = true
			if relativeDir == "." {
				filesByEcosystem[ecosystem] = append(filesByEcosystem[ecosystem], name)
			} else {
				filesByEcosystem[ecosystem] = append(filesByEcosystem[ecosystem], filepath.ToSlash(filepath.Join(relativeDir, name)))
			}
		}
		if signal.manager != "" && !managerSeen[ecosystem][signal.manager] {
			managerSeen[ecosystem][signal.manager] = true
			managersByEcosystem[ecosystem] = append(managersByEcosystem[ecosystem], signal.manager)
		}
	}

	var results []DetectionResult
	for _, ecosystem := range ecosystemOrder {
		matched := filesByEcosystem[ecosystem]
		if len(matched) == 0 {
			continue
		}
		slices.Sort(matched)
		slices.Sort(managersByEcosystem[ecosystem])

		result := DetectionResult{
			Ecosystem:    ecosystem,
			Directory:    relativeDir,
			Managers:     managersByEcosystem[ecosystem],
			MatchedFiles: matched,
		}

		if ecosystem == EcosystemNode && len(matched) > 1 {
			result.Warnings = append(result.Warnings, Warning{
				Code:    WarningNodeMultipleLockfiles,
				Message: "multiple Node lockfiles detected",
			})
		}

		results = append(results, result)
	}

	return results, nil
}
