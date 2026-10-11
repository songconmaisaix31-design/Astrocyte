// Package worksscan observes works only inside explicitly authorized roots
// through fixed programs. It never walks the whole disk, reads credential files
// or executes a project script.
package worksscan

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

const (
	markerFileName = ".astrocyte-works.json"
	maxRoots       = 256
	maxFiles       = 10000
	maxDepth       = 8
	maxMarkerBytes = 64 * 1024
)

// Scanner observes bounded roots and reads file facts directly from disk. It
// is intentionally small: one depth-limited walk per root, no symlink chasing
// and no content read beyond the optional marker manifest.
type Scanner struct {
	now func() time.Time
}

// NewScanner returns a scanner using the real clock.
func NewScanner() *Scanner { return &Scanner{now: time.Now} }

type marker struct {
	Name     string `json:"name"`
	Platform string `json:"platform"`
}

// ScanWorks observes each root as a work folder (or single file) and verifies
// its file facts. A root may carry a marker manifest naming the work and its
// platform; without one the platform stays "unknown". Errors are per root so a
// partial scan is reported, never hidden behind a single failure.
func (s *Scanner) ScanWorks(ctx context.Context, roots []string) ([]domain.Work, []domain.WorkDiscoveryFailure, error) {
	if len(roots) > maxRoots {
		return nil, nil, errors.New("too many authorized roots")
	}
	works := []domain.Work{}
	failures := []domain.WorkDiscoveryFailure{}
	for _, root := range roots {
		if ctx.Err() != nil {
			return works, failures, ctx.Err()
		}
		canonical, err := canonicalRoot(root)
		if err != nil {
			failures = append(failures, domain.WorkDiscoveryFailure{Root: root, Reason: "root_unavailable"})
			continue
		}
		info, err := os.Lstat(canonical)
		if err != nil {
			failures = append(failures, domain.WorkDiscoveryFailure{Root: canonical, Reason: "stat_failed"})
			continue
		}
		if len(works) >= maxRoots {
			failures = append(failures, domain.WorkDiscoveryFailure{Root: canonical, Reason: "work_limit_reached"})
			continue
		}
		work := s.observe(ctx, canonical, info)
		works = append(works, work)
		if !work.Files.Verified {
			failures = append(failures, domain.WorkDiscoveryFailure{Root: canonical, Reason: work.Files.Reason})
		}
	}
	return works, failures, nil
}

func (s *Scanner) observe(ctx context.Context, root string, info os.FileInfo) domain.Work {
	kind := domain.WorkKindFolder
	if !info.IsDir() {
		kind = domain.WorkKindFile
	}
	name := filepath.Base(root)
	platform := "unknown"
	if m, ok := readMarker(root); ok {
		if domain.ValidWorkName(m.Name) && m.Name != "" {
			name = m.Name
		}
		if domain.ValidWorkPlatform(m.Platform) && m.Platform != "" {
			platform = m.Platform
		}
	}
	mtime := info.ModTime().UTC()
	fileCount := 1
	reason := ""
	if info.IsDir() {
		mtime, fileCount, reason = walkFiles(ctx, root, info.ModTime())
	}
	return domain.Work{
		ID:       domain.WorkIdentity(platform, root),
		Platform: platform,
		Name:     name,
		Root:     root,
		Kind:     kind,
		Source:   domain.WorkSourceScan,
		Files: domain.WorkFileFact{
			SourcePath: root,
			Mtime:      &mtime,
			FileCount:  fileCount,
			Verified:   reason == "",
			Reason:     reason,
		},
	}
}

// walkFiles counts regular files and returns the newest mtime, bounded by depth
// and file count. Hidden, VCS and dependency directories are skipped so the
// walk never strays into credentials or vendored content.
func walkFiles(ctx context.Context, root string, fallback time.Time) (time.Time, int, string) {
	newest := fallback
	count := 0
	reason := ""
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if ctx.Err() != nil {
			reason = "scan_cancelled"
			return fs.SkipAll
		}
		if err != nil {
			reason = "files_partially_unreadable"
			return nil
		}
		if d.IsDir() {
			if path != root {
				name := d.Name()
				if name == ".git" || name == "node_modules" || name == "vendor" || name == ".venv" || (strings.HasPrefix(name, ".") && name != ".") {
					return filepath.SkipDir
				}
				rel, _ := filepath.Rel(root, path)
				if len(strings.Split(rel, string(filepath.Separator))) > maxDepth {
					reason = "scan_depth_limit"
					return filepath.SkipDir
				}
			}
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			reason = "files_partially_unreadable"
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		count++
		if count > maxFiles {
			reason = "scan_file_limit"
			return fs.SkipAll
		}
		if info.ModTime().After(newest) {
			newest = info.ModTime()
		}
		return nil
	})
	if count > maxFiles {
		count = maxFiles
	}
	return newest.UTC(), count, reason
}

// readMarker reads a bounded manifest naming the work and platform. It reads
// nothing else in the root and never follows the marker through a symlink.
func readMarker(root string) (marker, bool) {
	path := filepath.Join(root, markerFileName)
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || info.Size() > maxMarkerBytes {
		return marker{}, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return marker{}, false
	}
	var m marker
	if json.Unmarshal(data, &m) != nil {
		return marker{}, false
	}
	return m, true
}

func canonicalRoot(root string) (string, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	clean := filepath.Clean(abs)
	// A naming alias is permitted, but a root or ancestor redirect is not.
	// Check before expanding Windows 8.3 names so aliases cannot hide links.
	for path := clean; ; path = filepath.Dir(path) {
		info, err := os.Lstat(path)
		if err != nil {
			return "", err
		}
		if rootRedirect(info) {
			return "", errors.New("root must not cross a symlink")
		}
		if filepath.Dir(path) == path {
			break
		}
	}
	named, err := expandRootNames(clean)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(clean)
	if err != nil {
		return "", err
	}
	resolved, err = expandRootNames(resolved)
	if err != nil || !strings.EqualFold(resolved, named) {
		return "", errors.New("root must not cross a symlink")
	}
	// Preserve the exact approved root spelling and work identity.
	return clean, nil
}

// SortedRoots is a small helper so callers can present roots deterministically.
func SortedRoots(roots []string) []string {
	out := append([]string{}, roots...)
	sort.Strings(out)
	return out
}
