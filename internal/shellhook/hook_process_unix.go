//go:build darwin || linux || freebsd || openbsd

package shellhook

import (
	"os/exec"
	"syscall"
)

func configureBackgroundHookCommand(command *exec.Cmd) {
	if command.SysProcAttr == nil {
		command.SysProcAttr = &syscall.SysProcAttr{}
	}
	command.SysProcAttr.Setsid = true
}
