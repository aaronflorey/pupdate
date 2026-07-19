package detection

import (
	"strings"
	"sync"

	"github.com/git-pkgs/managers/definitions"
)

// supportedManagers maps manager definition names to pupdate ecosystems.
// Detection filenames are sourced automatically from the managers definitions,
// so adding a manager here picks up its lockfiles without further changes to
// the detection tables.
var supportedManagers = map[string]Ecosystem{
	"bun":      EcosystemNode,
	"npm":      EcosystemNode,
	"pnpm":     EcosystemNode,
	"yarn":     EcosystemNode,
	"composer": EcosystemPHP,
	"uv":       EcosystemPython,
	"poetry":   EcosystemPython,
	"pip":      EcosystemPython,
	"gomod":    EcosystemGo,
	"cargo":    EcosystemRust,
}

// localSignals are detection signals for managers without upstream definitions.
var localSignals = map[string]supportedSignal{
	"kasetto.lock": {ecosystem: EcosystemKasetto, manager: "kst", name: "kasetto.lock"},
	"kasetto.yaml": {ecosystem: EcosystemKasetto, manager: "kst", name: "kasetto.yaml"},
	"kasetto.yml":  {ecosystem: EcosystemKasetto, manager: "kst", name: "kasetto.yml"},
	".gitmodules":  {ecosystem: EcosystemGit, manager: "git", name: ".gitmodules"},
}

// registryEntry pairs a detection signal with its definition priority so that
// higher-priority managers win when two definitions share a filename.
type registryEntry struct {
	signal   supportedSignal
	priority int
}

// signalRegistry maps lowercase filenames to their detection signal. Built
// lazily from managers definitions and local fallbacks.
var signalRegistry = sync.OnceValue(func() map[string]registryEntry {
	registry := make(map[string]registryEntry)

	defs, err := definitions.LoadEmbedded()
	if err != nil {
		panic("failed to load manager definitions: " + err.Error())
	}

	for _, def := range defs {
		ecosystem, ok := supportedManagers[def.Name]
		if !ok {
			continue
		}

		registerDetectionFiles(registry, def.Detection.Lockfiles, ecosystem, def.Name, def.Detection.Priority)

		// gomod detects on go.mod (a manifest), not go.sum (its lockfile).
		// Include manifests for gomod so go.mod is picked up.
		if def.Name == "gomod" {
			registerDetectionFiles(registry, def.Detection.Manifests, ecosystem, def.Name, def.Detection.Priority)
		}
	}

	for name, signal := range localSignals {
		key := strings.ToLower(name)
		if _, exists := registry[key]; !exists {
			registry[key] = registryEntry{signal: signal, priority: 0}
		}
	}

	return registry
})

func registerDetectionFiles(registry map[string]registryEntry, files []string, ecosystem Ecosystem, manager string, priority int) {
	for _, file := range files {
		key := strings.ToLower(file)
		entry := registryEntry{
			signal:   supportedSignal{ecosystem: ecosystem, manager: manager, name: file},
			priority: priority,
		}
		if existing, exists := registry[key]; !exists || priority > existing.priority {
			registry[key] = entry
		}
	}
}
