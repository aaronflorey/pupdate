//go:build darwin || linux

package shellhook

import (
	"os/exec"
	"testing"
)

func TestConfigureBackgroundHookCommandSetsNewSession(t *testing.T) {
	cmd := exec.Command("true")
	configureBackgroundHookCommand(cmd)
	if cmd.SysProcAttr == nil {
		t.Fatal("expected SysProcAttr to be set")
	}
	if !cmd.SysProcAttr.Setsid {
		t.Fatal("expected Setsid to be true so the child detaches from the controlling terminal")
	}
}
