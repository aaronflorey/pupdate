package detection

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func Detect(dir string) ([]DetectionResult, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	filesByEcosystem := map[Ecosystem][]string{
		EcosystemNode:   {},
		EcosystemPHP:    {},
		EcosystemGo:     {},
		EcosystemRust:   {},
		EcosystemPython: {},
	}
	managersByEcosystem := map[Ecosystem][]string{
		EcosystemNode:   {},
		EcosystemPHP:    {},
		EcosystemGo:     {},
		EcosystemRust:   {},
		EcosystemPython: {},
	}
	managerSeen := map[Ecosystem]map[string]bool{
		EcosystemNode:   {},
		EcosystemPHP:    {},
		EcosystemGo:     {},
		EcosystemRust:   {},
		EcosystemPython: {},
	}

	signalToEcosystem := map[string]Ecosystem{}
	for ecosystem, signals := range ecosystemSignals {
		for _, signal := range signals {
			signalToEcosystem[signal] = ecosystem
		}
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

		filesByEcosystem[ecosystem] = append(filesByEcosystem[ecosystem], name)
		if ecosystem == EcosystemNode {
			if manager, has := nodeManagerByLockfile[name]; has && !managerSeen[ecosystem][manager] {
				managerSeen[ecosystem][manager] = true
				managersByEcosystem[ecosystem] = append(managersByEcosystem[ecosystem], manager)
			}
		}
	}

	order := []Ecosystem{
		EcosystemNode,
		EcosystemPHP,
		EcosystemGo,
		EcosystemRust,
		EcosystemPython,
	}

	var results []DetectionResult
	for _, ecosystem := range order {
		matched := filesByEcosystem[ecosystem]
		if len(matched) == 0 {
			continue
		}
		slices.Sort(matched)

		result := DetectionResult{
			Ecosystem:    ecosystem,
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
