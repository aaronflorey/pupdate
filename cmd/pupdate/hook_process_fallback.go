//go:build !darwin && !linux && !freebsd && !openbsd

package main

import "os/exec"

func configureBackgroundHookCommand(_ *exec.Cmd) {}
