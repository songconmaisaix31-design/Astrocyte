//go:build windows

package events

import (
	"fmt"
	"golang.org/x/sys/windows"
	"os"
)

func relaySessionFileIdentity(f *os.File, _ os.FileInfo) (string, error) {
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(windows.Handle(f.Fd()), &info); err != nil {
		return "", sessionFailure("file_identity_unavailable")
	}
	if info.NumberOfLinks != 1 {
		return "", sessionFailure("file_alias_rejected")
	}
	return fmt.Sprintf("%x:%x:%x", info.VolumeSerialNumber, info.FileIndexHigh, info.FileIndexLow), nil
}
