package agents

import (
	"path/filepath"

	"golang.org/x/sys/windows"
)

// pathSpelling expands filesystem 8.3 names without resolving symlinks. Keep
// this separate from EvalSymlinks: comparing the two still detects redirected
// native roots and ancestors. No file contents are read or roots authorized.
func pathSpelling(path string) string {
	clean := filepath.Clean(path)
	if !filepath.IsAbs(clean) {
		return clean
	}
	ptr, err := windows.UTF16PtrFromString(clean)
	if err != nil {
		return clean
	}
	buf := make([]uint16, 32768)
	n, err := windows.GetLongPathName(ptr, &buf[0], uint32(len(buf)))
	if err == nil && n > 0 && n < uint32(len(buf)) {
		return filepath.Clean(windows.UTF16ToString(buf[:n]))
	}
	// Missing checkouts retain their registered association metadata. Expand
	// only existing ancestors, leaving the missing suffix literal; permission
	// errors do not authorize a fallback search or a replacement directory.
	if err == windows.ERROR_FILE_NOT_FOUND || err == windows.ERROR_PATH_NOT_FOUND {
		parent := filepath.Dir(clean)
		if parent != clean {
			return filepath.Join(pathSpelling(parent), filepath.Base(clean))
		}
	}
	return clean
}

func pathIsRedirect(path string) bool {
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return true
	}
	attrs, err := windows.GetFileAttributes(ptr)
	return err == nil && attrs&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0
}
