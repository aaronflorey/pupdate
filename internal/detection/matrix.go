package detection

import (
	"strings"
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
	EcosystemKasetto,
	EcosystemGit,
}

func detectSupportedSignal(name string) (supportedSignal, bool) {
	registry := signalRegistry()
	entry, ok := registry[strings.ToLower(name)]
	if !ok {
		return supportedSignal{}, false
	}
	return entry.signal, true
}
