package detection

var ecosystemSignals = map[Ecosystem][]string{
	EcosystemNode: {
		"bun.lock",
		"pnpm-lock.yaml",
		"pnpm-lock.yml",
		"yarn.lock",
		"package-lock.json",
	},
	EcosystemPHP: {
		"composer.lock",
	},
	EcosystemGo: {
		"go.mod",
	},
	EcosystemRust: {
		"cargo.toml",
	},
	EcosystemPython: {
		"pyproject.toml",
		"poetry.lock",
		"uv.lock",
		"pipfile.lock",
		"requirements.txt",
	},
}

var nodeManagerByLockfile = map[string]string{
	"bun.lock":          "bun",
	"pnpm-lock.yaml":    "pnpm",
	"pnpm-lock.yml":     "pnpm",
	"yarn.lock":         "yarn",
	"package-lock.json": "npm",
}
