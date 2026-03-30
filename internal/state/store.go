package state

import (
	"errors"
	"os"
	"path/filepath"
)

const invalidStateWarning = "state file is invalid; treating as empty"

type Store struct {
	Dir string
}

func NewStore(dir string) Store {
	return Store{Dir: dir}
}

func (s Store) Load() (FileState, []string, error) {
	path := filepath.Join(s.Dir, FileName)
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Empty(), nil, nil
		}
		return Empty(), nil, err
	}

	state, warnings, err := Decode(raw)
	if err != nil {
		return Empty(), []string{invalidStateWarning}, nil
	}

	return state, warnings, nil
}

func (s Store) Save(state FileState) error {
	encoded, err := Encode(state)
	if err != nil {
		return err
	}

	dir := s.Dir
	tmp, err := os.CreateTemp(dir, ".pupdate.tmp-*")
	if err != nil {
		return err
	}

	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(encoded); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, filepath.Join(s.Dir, FileName))
}
