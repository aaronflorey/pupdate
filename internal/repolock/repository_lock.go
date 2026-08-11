package repolock

import (
	"fmt"
	"os"
	"path/filepath"
)

const FileName = ".pupdate.hook.lock"

type Lock struct {
	file *os.File
	path string
}

func Path(root string) string {
	return filepath.Join(root, FileName)
}

func Acquire(root string) (*Lock, bool, error) {
	path := Path(root)
	file, acquired, err := tryAcquireRepositoryLock(path)
	if err != nil {
		return nil, false, fmt.Errorf("failed to acquire repository lock %s: %w", path, err)
	}
	if !acquired {
		return nil, false, nil
	}
	return &Lock{file: file, path: path}, true, nil
}

func (lock *Lock) Release() error {
	if lock == nil || lock.file == nil {
		return nil
	}
	err := releaseRepositoryLock(lock.file, lock.path)
	lock.file = nil
	if err != nil {
		return fmt.Errorf("failed to release repository lock %s: %w", lock.path, err)
	}
	return nil
}

func Active(root string) (bool, error) {
	path := Path(root)
	active, err := repositoryLockFileActive(path)
	if err != nil {
		return false, fmt.Errorf("failed to inspect repository lock %s: %w", path, err)
	}
	return active, nil
}
