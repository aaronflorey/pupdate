package main

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
