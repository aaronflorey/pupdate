package main

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
