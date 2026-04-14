package main

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
