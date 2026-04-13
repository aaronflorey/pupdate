package detection

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func Detect(dir string) ([]DetectionResult, error) {
	directories, err := scanDirectories(dir)
	if err != nil {
		return nil, err
	}

	signalToEcosystem := map[string]Ecosystem{}
	for ecosystem, signals := range ecosystemSignals {
		for _, signal := range signals {
			signalToEcosystem[signal] = ecosystem
		}
	}

	order := []Ecosystem{
		EcosystemNode,
		EcosystemPHP,
		EcosystemGo,
		EcosystemRust,
		EcosystemPython,
		EcosystemGit,
	}

	results := []DetectionResult{}
	for _, directory := range directories {
		dirPath := dir
		if directory != "." {
			dirPath = filepath.Join(dir, directory)
		}

		directoryResults, err := detectDirectory(dirPath, directory, signalToEcosystem, order)
		if err != nil {
			return nil, err
		}
		results = append(results, directoryResults...)
	}

	return results, nil
}

func scanDirectories(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	directories := []string{"."}
	for _, entry := range entries {
		if !entry.IsDir() || (entry.Type()&os.ModeSymlink) != 0 {
			continue
		}
		directories = append(directories, entry.Name())
	}
	slices.Sort(directories[1:])
	return directories, nil
}

func detectDirectory(dirPath string, relativeDir string, signalToEcosystem map[string]Ecosystem, order []Ecosystem) ([]DetectionResult, error) {
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

		name := strings.ToLower(filepath.Base(entry.Name()))
		ecosystem, ok := signalToEcosystem[name]
		if !ok {
			continue
		}

		if !fileSeen[ecosystem][name] {
			fileSeen[ecosystem][name] = true
			if relativeDir == "." {
				filesByEcosystem[ecosystem] = append(filesByEcosystem[ecosystem], name)
			} else {
				filesByEcosystem[ecosystem] = append(filesByEcosystem[ecosystem], filepath.ToSlash(filepath.Join(relativeDir, name)))
			}
		}
		if ecosystemManagers, hasEcosystem := managerBySignal[ecosystem]; hasEcosystem {
			if manager, hasManager := ecosystemManagers[name]; hasManager && !managerSeen[ecosystem][manager] {
				managerSeen[ecosystem][manager] = true
				managersByEcosystem[ecosystem] = append(managersByEcosystem[ecosystem], manager)
			}
		}
	}

	var results []DetectionResult
	for _, ecosystem := range order {
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
