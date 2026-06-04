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

type Options struct {
	WorkspaceGlobs  []string
	FolderBlacklist []string
}

func Detect(dir string) ([]DetectionResult, error) {
	return DetectWithOptions(dir, Options{})
}

func DetectWithWorkspaceGlobs(dir string, workspaceGlobs []string) ([]DetectionResult, error) {
	return DetectWithOptions(dir, Options{WorkspaceGlobs: workspaceGlobs})
}

func DetectWithOptions(dir string, options Options) ([]DetectionResult, error) {
	directories, err := scanDirectories(dir, options)
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

func scanDirectories(dir string, options Options) ([]string, error) {
	matcher, err := loadIgnoreMatcher(dir)
	if err != nil {
		return nil, err
	}

	folderBlacklist := makeFolderBlacklistSet(options.FolderBlacklist)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	directories := []string{"."}
	directorySet := map[string]struct{}{".": {}}
	for _, entry := range entries {
		if !entry.IsDir() || (entry.Type()&os.ModeSymlink) != 0 {
			continue
		}

		relPath := entry.Name()
		if shouldSkipDirectory(matcher, folderBlacklist, relPath) {
			continue
		}

		directories = appendDirectory(directories, directorySet, relPath)

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
			if shouldSkipDirectory(matcher, folderBlacklist, packagePath) {
				continue
			}

			directories = appendDirectory(directories, directorySet, packagePath)
		}
	}

	for _, workspaceGlob := range options.WorkspaceGlobs {
		matchedDirectories, err := scanWorkspaceGlob(dir, matcher, folderBlacklist, workspaceGlob)
		if err != nil {
			return nil, err
		}
		for _, matchedDirectory := range matchedDirectories {
			directories = appendDirectory(directories, directorySet, matchedDirectory)
		}
	}

	slices.Sort(directories[1:])
	return directories, nil
}

func appendDirectory(directories []string, seen map[string]struct{}, relPath string) []string {
	if _, ok := seen[relPath]; ok {
		return directories
	}

	seen[relPath] = struct{}{}
	return append(directories, relPath)
}

func makeFolderBlacklistSet(entries []string) map[string]struct{} {
	if len(entries) == 0 {
		return nil
	}

	set := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		set[entry] = struct{}{}
	}
	return set
}

func scanWorkspaceGlob(dir string, matcher ignoreMatcher, folderBlacklist map[string]struct{}, workspaceGlob string) ([]string, error) {
	segments := strings.Split(filepath.ToSlash(workspaceGlob), "/")
	directories := []string{}

	var walk func(string, int) error
	walk = func(relativeDir string, segmentIndex int) error {
		if segmentIndex == len(segments) {
			if relativeDir != "." {
				directories = append(directories, relativeDir)
			}
			return nil
		}

		dirPath := dir
		if relativeDir != "." {
			dirPath = filepath.Join(dir, filepath.FromSlash(relativeDir))
		}

		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			if !entry.IsDir() || (entry.Type()&os.ModeSymlink) != 0 {
				continue
			}

			relPath := entry.Name()
			if relativeDir != "." {
				relPath = filepath.ToSlash(filepath.Join(relativeDir, entry.Name()))
			}
			if shouldSkipDirectory(matcher, folderBlacklist, relPath) {
				continue
			}

			matched, err := filepath.Match(segments[segmentIndex], entry.Name())
			if err != nil {
				return err
			}
			if !matched {
				continue
			}

			if err := walk(relPath, segmentIndex+1); err != nil {
				return err
			}
		}

		return nil
	}

	if err := walk(".", 0); err != nil {
		return nil, err
	}

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

func shouldSkipDirectory(matcher ignoreMatcher, folderBlacklist map[string]struct{}, path string) bool {
	if folderBlacklist != nil {
		normalizedPath := filepath.ToSlash(path)
		directoryName := normalizedPath[strings.LastIndex(normalizedPath, "/")+1:]
		if _, ok := folderBlacklist[directoryName]; ok {
			return true
		}
	}

	if matcher != nil {
		parts := strings.Split(filepath.ToSlash(path), "/")
		if matcher.Match(parts, true) {
			return true
		}
	}

	return false
}

func detectDirectory(dirPath string, relativeDir string) ([]DetectionResult, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	filesByEcosystem := map[Ecosystem][]string{
		EcosystemNode:    {},
		EcosystemPHP:     {},
		EcosystemGo:      {},
		EcosystemRust:    {},
		EcosystemPython:  {},
		EcosystemKasetto: {},
		EcosystemGit:     {},
	}
	managersByEcosystem := map[Ecosystem][]string{
		EcosystemNode:    {},
		EcosystemPHP:     {},
		EcosystemGo:      {},
		EcosystemRust:    {},
		EcosystemPython:  {},
		EcosystemKasetto: {},
		EcosystemGit:     {},
	}
	fileSeen := map[Ecosystem]map[string]bool{
		EcosystemNode:    {},
		EcosystemPHP:     {},
		EcosystemGo:      {},
		EcosystemRust:    {},
		EcosystemPython:  {},
		EcosystemKasetto: {},
		EcosystemGit:     {},
	}
	managerSeen := map[Ecosystem]map[string]bool{
		EcosystemNode:    {},
		EcosystemPHP:     {},
		EcosystemGo:      {},
		EcosystemRust:    {},
		EcosystemPython:  {},
		EcosystemKasetto: {},
		EcosystemGit:     {},
	}

	for _, entry := range entries {
		if entry.IsDir() || (entry.Type()&os.ModeSymlink) != 0 {
			continue
		}

		matchedName := filepath.Base(entry.Name())
		signal, ok := detectSupportedSignal(matchedName)
		if !ok {
			continue
		}
		name := signal.name
		ecosystem := signal.ecosystem

		if !fileSeen[ecosystem][name] {
			fileSeen[ecosystem][name] = true
			if relativeDir == "." {
				filesByEcosystem[ecosystem] = append(filesByEcosystem[ecosystem], matchedName)
			} else {
				filesByEcosystem[ecosystem] = append(filesByEcosystem[ecosystem], filepath.ToSlash(filepath.Join(relativeDir, matchedName)))
			}
		} else if matchedName == name {
			preferredPath := matchedName
			if relativeDir != "." {
				preferredPath = filepath.ToSlash(filepath.Join(relativeDir, matchedName))
			}

			for i, existing := range filesByEcosystem[ecosystem] {
				if strings.EqualFold(filepath.Base(existing), name) {
					filesByEcosystem[ecosystem][i] = preferredPath
					break
				}
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
