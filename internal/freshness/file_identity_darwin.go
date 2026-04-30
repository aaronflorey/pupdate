//go:build darwin

package freshness

import (
	"fmt"
	"os"
	"syscall"

	"github.com/aaronflorey/pupdate/internal/state"
)

func enrichLockfileMetadata(info os.FileInfo, metadata *state.LockfileMetadata) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return
	}

	metadata.FileID = fmt.Sprintf("%d:%d", stat.Dev, stat.Ino)
	metadata.ChangeTimeUnixNano = stat.Ctimespec.Sec*1_000_000_000 + stat.Ctimespec.Nsec
}
