//go:build !linux && !darwin

package freshness

import (
	"os"

	"github.com/aaronflorey/pupdate/internal/state"
)

func enrichLockfileMetadata(_ os.FileInfo, _ *state.LockfileMetadata) {
}
