package main

import (
	"fmt"
	"strings"
)

const bashInitSnippet = `# pupdate hook
_pupdate_last_pwd=""
_pupdate_hook() {
  if [ "$PWD" != "$_pupdate_last_pwd" ]; then
    _pupdate_last_pwd="$PWD"
    pupdate run --quiet
  fi
}
if [ -n "$PROMPT_COMMAND" ]; then
  PROMPT_COMMAND="_pupdate_hook;$PROMPT_COMMAND"
else
  PROMPT_COMMAND="_pupdate_hook"
fi
`

const zshInitSnippet = `# pupdate hook
autoload -U add-zsh-hook
_pupdate_last_pwd=""
_pupdate_hook() {
  if [ "$PWD" != "$_pupdate_last_pwd" ]; then
    _pupdate_last_pwd="$PWD"
    pupdate run --quiet
  fi
}
add-zsh-hook chpwd _pupdate_hook
add-zsh-hook precmd _pupdate_hook
`

const fishInitSnippet = `# pupdate hook
set -g __pupdate_last_pwd ""
function __pupdate_hook --on-variable PWD
  if test "$PWD" != "$__pupdate_last_pwd"
    set -g __pupdate_last_pwd "$PWD"
    pupdate run --quiet
  end
end
__pupdate_hook
`

var supportedInitShells = []string{"bash", "zsh", "fish"}

func supportedInitShellsText() string {
	return strings.Join(supportedInitShells, ", ")
}

func isSupportedInitShell(shell string) bool {
	for _, supported := range supportedInitShells {
		if shell == supported {
			return true
		}
	}

	return false
}

func initSnippetForShell(shell string) (string, error) {
	switch shell {
	case "bash":
		return bashInitSnippet, nil
	case "zsh":
		return zshInitSnippet, nil
	case "fish":
		return fishInitSnippet, nil
	default:
		return "", fmt.Errorf("unsupported shell %q; supported shells: %s", shell, supportedInitShellsText())
	}
}
