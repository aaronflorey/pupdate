//go:build !darwin && !linux && !freebsd && !openbsd

package shellhook

import "os/exec"

func configureBackgroundHookCommand(_ *exec.Cmd) {}
