//go:build windows

package main

import "os"

func backgroundHookProcessRunning(pid int) (bool, error) {
	if pid <= 0 {
		return false, nil
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false, nil
	}
	_ = proc.Release()
	return true, nil
}
