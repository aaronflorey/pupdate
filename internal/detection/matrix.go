package detection

var ecosystemSignals = map[Ecosystem][]string{
	EcosystemNode: {
		"bun.lock",
		"package-lock.json",
		"pnpm-lock.yaml",
		"yarn.lock",
	},
	EcosystemPHP: {
		"composer.lock",
	},
	EcosystemPython: {
		"uv.lock",
		"poetry.lock",
		"requirements.txt",
	},
	EcosystemGo: {
		"go.mod",
	},
	EcosystemRust: {
		"cargo.lock",
	},
	EcosystemGit: {
		".gitmodules",
	},
}

var nodeManagerByLockfile = map[string]string{
	"bun.lock":          "bun",
	"package-lock.json": "npm",
	"pnpm-lock.yaml":    "pnpm",
	"yarn.lock":         "yarn",
}

var pythonManagerBySignal = map[string]string{
	"uv.lock":          "uv",
	"poetry.lock":      "poetry",
	"requirements.txt": "pip",
}

var managerBySignal = map[Ecosystem]map[string]string{
	EcosystemNode:   nodeManagerByLockfile,
	EcosystemPython: pythonManagerBySignal,
	EcosystemGo: {
		"go.mod": "go",
	},
	EcosystemRust: {
		"cargo.lock": "cargo",
	},
	EcosystemGit: {
		".gitmodules": "git",
	},
}
