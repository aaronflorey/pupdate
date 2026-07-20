package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type repositoryLock struct {
	file *os.File
	path string
}

func acquireRepositoryLock(root string) (*repositoryLock, bool, error) {
	path := backgroundHookLockPath(root)
	file, acquired, err := tryAcquireRepositoryLock(path)
	if err != nil {
		return nil, false, fmt.Errorf("failed to acquire repository lock %s: %w", path, err)
	}
	if !acquired {
		return nil, false, nil
	}
	return &repositoryLock{file: file, path: path}, true, nil
}

func (lock *repositoryLock) release() error {
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

func repositoryLockActive(root string) (bool, error) {
	path := filepath.Join(root, backgroundHookLockFileName)
	active, err := repositoryLockFileActive(path)
	if err != nil {
		return false, fmt.Errorf("failed to inspect repository lock %s: %w", path, err)
	}
	return active, nil
}
