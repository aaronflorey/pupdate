package detection

type Ecosystem string

const (
	EcosystemNode   Ecosystem = "node"
	EcosystemPHP    Ecosystem = "php"
	EcosystemGo     Ecosystem = "go"
	EcosystemRust   Ecosystem = "rust"
	EcosystemPython Ecosystem = "python"
)

type WarningCode string

const (
	WarningNodeMultipleLockfiles WarningCode = "node_multiple_lockfiles"
)

type Warning struct {
	Code    WarningCode
	Message string
}

type DetectionResult struct {
	Ecosystem   Ecosystem
	Managers    []string
	MatchedFiles []string
	Warnings    []Warning
}
