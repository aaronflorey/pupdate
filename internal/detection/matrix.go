package detection

import (
	"strings"

	"github.com/git-pkgs/manifests"
)

type supportedSignal struct {
	ecosystem Ecosystem
	manager   string
	name      string
}

var ecosystemOrder = []Ecosystem{
	EcosystemNode,
	EcosystemPHP,
	EcosystemGo,
	EcosystemRust,
	EcosystemPython,
	EcosystemGit,
}

var supportedSignalNames = []string{
	"bun.lock",
	"package-lock.json",
	"pnpm-lock.yaml",
	"yarn.lock",
	"composer.lock",
	"uv.lock",
	"poetry.lock",
	"requirements.txt",
	"go.mod",
	"Cargo.lock",
	".gitmodules",
}

var signalManagerByName = map[string]string{
	"bun.lock":          "bun",
	"package-lock.json": "npm",
	"pnpm-lock.yaml":    "pnpm",
	"yarn.lock":         "yarn",
	"composer.lock":     "composer",
	"uv.lock":           "uv",
	"poetry.lock":       "poetry",
	"requirements.txt":  "pip",
	"go.mod":            "go",
	"Cargo.lock":        "cargo",
	".gitmodules":       "git",
}

var canonicalSignalByLowerName = buildCanonicalSignalNames()

func buildCanonicalSignalNames() map[string]string {
	canonical := make(map[string]string, len(supportedSignalNames))
	for _, name := range supportedSignalNames {
		canonical[strings.ToLower(name)] = name
	}
	return canonical
}

func detectSupportedSignal(name string) (supportedSignal, bool) {
	canonicalName, ok := canonicalSignalByLowerName[strings.ToLower(name)]
	if !ok {
		return supportedSignal{}, false
	}

	ecosystemName, _, ok := manifests.Identify(canonicalName)
	if !ok {
		return supportedSignal{}, false
	}

	ecosystem, ok := manifestEcosystemToDetection(ecosystemName)
	if !ok {
		return supportedSignal{}, false
	}

	return supportedSignal{
		ecosystem: ecosystem,
		manager:   signalManagerByName[canonicalName],
		name:      strings.ToLower(canonicalName),
	}, true
}

func manifestEcosystemToDetection(ecosystem string) (Ecosystem, bool) {
	switch ecosystem {
	case "npm":
		return EcosystemNode, true
	case "composer":
		return EcosystemPHP, true
	case "golang":
		return EcosystemGo, true
	case "cargo":
		return EcosystemRust, true
	case "pypi":
		return EcosystemPython, true
	case "git":
		return EcosystemGit, true
	default:
		return "", false
	}
}
