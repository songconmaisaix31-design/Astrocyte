package events

import (
	"fmt"
	"os"
	"path/filepath"
)

// ValidateLocalReplayPath reuses the existing component/reparse/8.3 checks at
// command entrypoints. The caller still restricts which roots are authorized.
func ValidateLocalReplayPath(path string) error {
	_, err := unredirectedReplayPath(path)
	return err
}

// Check every selected path component independently of name expansion. This
// also rejects redirects within the approved root, not only outside escapes.
func unredirectedReplayPath(path string) (string, error) {
	for current := filepath.Clean(path); ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err != nil {
			return "", err
		}
		redirect, err := replayPathRedirect(current, info)
		if err != nil {
			return "", err
		}
		if redirect {
			return "", fmt.Errorf("replay file uses symbolic links")
		}
		if filepath.Dir(current) == current {
			break
		}
	}
	names, err := expandReplayNames(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	resolvedNames, err := expandReplayNames(resolved)
	if err != nil || !replayNamesEqual(names, resolvedNames) {
		return "", fmt.Errorf("replay file uses symbolic links")
	}
	return names, nil
}

func sameReplayProject(a, b string) bool {
	a, b = filepath.Clean(a), filepath.Clean(b)
	if replayNamesEqual(a, b) {
		return true
	}
	// Relative fixture headers and unavailable projects keep their literal
	// association. Only existing absolute names can establish an 8.3 alias.
	if !filepath.IsAbs(a) || !filepath.IsAbs(b) {
		return false
	}
	left, err := expandReplayNames(a)
	if err != nil {
		return false
	}
	right, err := expandReplayNames(b)
	return err == nil && replayNamesEqual(left, right)
}
