//go:build windows

package agents

import "golang.org/x/sys/windows"

// MoveFile never replaces an existing target, including an empty directory
// created after our preflight check. Both paths are under the same owned root.
func publishManagedCheckout(source, target string) error {
	s, err := windows.UTF16PtrFromString(source)
	if err != nil {
		return err
	}
	t, err := windows.UTF16PtrFromString(target)
	if err != nil {
		return err
	}
	return windows.MoveFile(s, t)
}
