package detection

type Ecosystem string

const (
	EcosystemNode   Ecosystem = "node"
	EcosystemPHP    Ecosystem = "php"
	EcosystemGo     Ecosystem = "go"
	EcosystemRust   Ecosystem = "rust"
	EcosystemPython Ecosystem = "python"
	EcosystemGit    Ecosystem = "git"
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
	Ecosystem    Ecosystem
	Directory    string
	Managers     []string
	MatchedFiles []string
	Warnings     []Warning
}

func (r DetectionResult) StateKey() string {
	if r.Directory == "" || r.Directory == "." {
		return string(r.Ecosystem)
	}
	return string(r.Ecosystem) + "@" + r.Directory
}
