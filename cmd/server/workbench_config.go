package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func workbenchRoot() string {
	return filepath.Join(os.Getenv("LOCALAPPDATA"), "astrocyte-workbench", "v0.1")
}

func workbenchProtectedSources() ([]string, error) {
	value := os.Getenv("ASTROCYTE_WORKBENCH_PROTECTED_SOURCES")
	if value == "" {
		return []string{}, nil
	}
	decoder := json.NewDecoder(strings.NewReader(value))
	var paths []string
	if decoder.Decode(&paths) != nil || decoder.Decode(new(any)) != io.EOF || paths == nil {
		return nil, errors.New("protected workbench sources must be an explicit JSON array")
	}
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			return nil, errors.New("protected workbench source must be absolute")
		}
		if _, err := workbenchCanonical(path); err != nil {
			return nil, errors.New("protected workbench source unavailable")
		}
	}
	return paths, nil
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
	sources, e := workbenchProtectedSources()
	if e != nil {
		return "", e
	}
	directories := []string{filepath.Join(workbenchRoot(), "snapshots")}
	for _, source := range sources {
		directories = append(directories, filepath.Dir(source))
	}
	for _, protected := range directories {
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
		for _, path := range sources {
			if source, e := os.Stat(path); e == nil && os.SameFile(appDB, source) {
				return "", errors.New("application database must be independent")
			}
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
