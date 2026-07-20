//go:build !darwin && !linux

package main

import "os/exec"

func configureInstallCommand(_ *exec.Cmd) {}
