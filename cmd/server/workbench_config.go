package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
)

func workbenchRoot() string {
	return filepath.Join(os.Getenv("LOCALAPPDATA"), "astrocyte-workbench", "v0.1")
}

// Resolve existing ancestors before constructing a new directory, including
// junctions. A source or published copy can never become the migrated app DB.
func workbenchCanonical(path string) (string, error) {
	path, e := filepath.Abs(path)
	if e != nil {
		return "", e
	}
	real, e := filepath.EvalSymlinks(path)
	if e == nil {
		return real, nil
	}
	if !os.IsNotExist(e) {
		return "", e
	}
	parent := filepath.Dir(path)
	if parent == path {
		return path, nil
	}
	real, e = workbenchCanonical(parent)
	return filepath.Join(real, filepath.Base(path)), e
}
func guardWorkbenchDataDir(dir string) (string, error) {
	if os.Getenv("ASTROCYTE_WORKBENCH") == "1" && os.Getenv("ASTROCYTE_DATA_DIR") == "" {
		dir = filepath.Join(workbenchRoot(), "app-state")
	}
	return guardWorkbenchExplicitDataDir(dir)
}

// The relay entrypoint has already selected its isolated data directory. Reuse
// every protected source/snapshot/identity check without applying v0.1 defaults.
func guardWorkbenchExplicitDataDir(dir string) (string, error) {
	resolved, e := workbenchCanonical(dir)
	if e != nil {
		return "", errors.New("application data directory unavailable")
	}
	resolvedDB, e := workbenchCanonical(filepath.Join(resolved, "state.sqlite"))
	if e != nil {
		return "", errors.New("application database unavailable")
	}
	for _, protected := range []string{filepath.Dir(sqlite.WorkbenchFormalSource), filepath.Join(workbenchRoot(), "snapshots")} {
		base, e := workbenchCanonical(protected)
		if e != nil {
			return "", errors.New("application data directory unavailable")
		}
		rel, e := filepath.Rel(strings.ToLower(base), strings.ToLower(resolved))
		if e == nil && (rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel))) {
			return "", errors.New("application data directory must be separate from workbench source and snapshots")
		}
		rel, e = filepath.Rel(strings.ToLower(base), strings.ToLower(resolvedDB))
		if e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel) {
			return "", errors.New("application database must be separate from workbench source and snapshots")
		}
		// Existing 8.3 aliases and parent identity must also fail closed.
		if a, e := os.Stat(base); e == nil {
			for p := resolved; ; p = filepath.Dir(p) {
				if b, e := os.Stat(p); e == nil && os.SameFile(a, b) {
					return "", errors.New("application data directory must be separate from workbench source and snapshots")
				}
				if filepath.Dir(p) == p {
					break
				}
			}
		}
	}
	appDB, e := os.Stat(filepath.Join(resolved, "state.sqlite"))
	if e == nil {
		if source, e := os.Stat(sqlite.WorkbenchFormalSource); e == nil && os.SameFile(appDB, source) {
			return "", errors.New("application database must be independent")
		}
		generations, _ := filepath.Glob(filepath.Join(workbenchRoot(), "snapshots", "*", "snapshot.sqlite"))
		for _, path := range generations {
			if snapshot, e := os.Stat(path); e == nil && os.SameFile(appDB, snapshot) {
				return "", errors.New("application database must be independent")
			}
		}
	}
	return dir, nil
}
