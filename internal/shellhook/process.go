package shellhook

import "os/exec"

func StartBackground(executable string, args []string) (int, error) {
	command := exec.Command(executable, args...)
	command.Stdin = nil
	command.Stdout = nil
	command.Stderr = nil
	configureBackgroundHookCommand(command)
	if err := command.Start(); err != nil {
		return 0, err
	}
	return command.Process.Pid, nil
}
