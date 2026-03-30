package state

import (
	"encoding/json"
	"fmt"
	"time"
)

const FileName = ".pupdate"

const SchemaVersion = 1

type FileState struct {
	Version    int                       `json:"version"`
	Ecosystems map[string]EcosystemState `json:"ecosystems"`
}

type EcosystemState struct {
	LastSuccessAt string            `json:"last_success_at"`
	Lockfiles     map[string]string `json:"lockfiles,omitempty"`
}

func Empty() FileState {
	return FileState{
		Version:    SchemaVersion,
		Ecosystems: map[string]EcosystemState{},
	}
}

func Decode(raw []byte) (FileState, []string, error) {
	var decoded FileState
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return Empty(), nil, err
	}

	if decoded.Version != SchemaVersion {
		return Empty(), []string{
			fmt.Sprintf("state schema version mismatch: got %d expected %d; treating as empty state", decoded.Version, SchemaVersion),
		}, nil
	}

	if decoded.Ecosystems == nil {
		decoded.Ecosystems = map[string]EcosystemState{}
	}
	for key, ecosystem := range decoded.Ecosystems {
		if ecosystem.Lockfiles == nil {
			ecosystem.Lockfiles = map[string]string{}
			decoded.Ecosystems[key] = ecosystem
		}
	}

	return decoded, nil, nil
}

func Encode(state FileState) ([]byte, error) {
	if state.Version == 0 {
		state.Version = SchemaVersion
	}
	if state.Ecosystems == nil {
		state.Ecosystems = map[string]EcosystemState{}
	}
	for key, ecosystem := range state.Ecosystems {
		if ecosystem.Lockfiles == nil {
			ecosystem.Lockfiles = map[string]string{}
			state.Ecosystems[key] = ecosystem
		}
	}
	return json.Marshal(state)
}

func FormatRFC3339UTC(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func ParseRFC3339UTC(raw string) (time.Time, error) {
	return time.Parse(time.RFC3339, raw)
}
