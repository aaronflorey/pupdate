//go:build !darwin && !linux

package install

import "os/exec"

func configureInstallCommand(_ *exec.Cmd) {}
