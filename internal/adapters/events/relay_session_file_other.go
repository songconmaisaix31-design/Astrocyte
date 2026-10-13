//go:build !windows

package events

import (
	"fmt"
	"os"
	"syscall"
)

func relaySessionFileIdentity(_ *os.File, info os.FileInfo) (string, error) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat.Nlink != 1 {
		return "", sessionFailure("file_identity_unavailable_or_aliased")
	}
	return fmt.Sprintf("%x:%x", stat.Dev, stat.Ino), nil
}
