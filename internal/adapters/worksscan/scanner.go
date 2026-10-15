// Package worksscan 使用固定程序观察明确授权目录的文件事实。
package worksscan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// Scanner 对授权根进行有深度限制的遍历，仅从可选标记读取名称与平台。
type Scanner struct {
	now func() time.Time
}

// NewScanner 使用实际时钟。
func NewScanner() *Scanner { return &Scanner{now: time.Now} }

type marker struct {
	Name     string `json:"name"`
	Platform string `json:"platform"`
}

// ScanWorks 观察各授权根；缺少标记时平台保持未知，来源失败明确报告。
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
		work, err := s.observe(ctx, canonical, info)
		if err != nil {
			return nil, failures, err
		}
		works = append(works, work)
		if !work.Files.Verified {
			failures = append(failures, domain.WorkDiscoveryFailure{Root: canonical, Reason: work.Files.Reason})
		}
	}
	return works, failures, nil
}

func (s *Scanner) observe(ctx context.Context, root string, info os.FileInfo) (domain.Work, error) {
	kind := domain.WorkKindFolder
	if !info.IsDir() {
		kind = domain.WorkKindFile
	}
	name := filepath.Base(root)
	platform := "unknown"
	var m marker
	var ok bool
	var err error
	if info.IsDir() {
		m, ok, err = readMarker(root)
		if err != nil {
			return domain.Work{}, err
		}
	}
	if ok {
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
		mtime, fileCount, reason, err = walkFiles(ctx, root, info.ModTime())
		if err != nil {
			return domain.Work{}, err
		}
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
	}, nil
}

// walkFiles 保留实际 mtime 与计数；深度和数量限制明确标记为未核实。
func walkFiles(ctx context.Context, root string, rootTime time.Time) (time.Time, int, string, error) {
	newest := rootTime
	count := 0
	reason := ""
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if ctx.Err() != nil {
			reason = "scan_cancelled"
			return ctx.Err()
		}
		if err != nil {
			reason = "files_partially_unreadable"
			return err
		}
		if d.IsDir() {
			if path != root {
				name := d.Name()
				if name == ".git" || name == "node_modules" || name == "vendor" || name == ".venv" || (strings.HasPrefix(name, ".") && name != ".") {
					return filepath.SkipDir
				}
				rel, err := filepath.Rel(root, path)
				if err != nil {
					return err
				}
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
			return err
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
	if err != nil {
		return time.Time{}, 0, reason, fmt.Errorf("scan files: %w", err)
	}
	if count > maxFiles {
		count = maxFiles
	}
	return newest.UTC(), count, reason, nil
}

// readMarker 只读取实际标记文件；缺失保持未知，已提供的错误立即返回。
func readMarker(root string) (marker, bool, error) {
	path := filepath.Join(root, markerFileName)
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return marker{}, false, nil
	}
	if err != nil {
		return marker{}, false, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Size() > maxMarkerBytes {
		return marker{}, false, errors.New("work marker must be a bounded ordinary file")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return marker{}, false, err
	}
	var m *marker
	if err := json.Unmarshal(data, &m); err != nil {
		return marker{}, false, fmt.Errorf("work marker JSON: %w", err)
	}
	if m == nil || (m.Name != "" && !domain.ValidWorkName(m.Name)) || (m.Platform != "" && !domain.ValidWorkPlatform(m.Platform)) {
		return marker{}, false, errors.New("invalid work marker fields")
	}
	return *m, true, nil
}

func canonicalRoot(root string) (string, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	clean := filepath.Clean(abs)
	// 先检查根及祖先重定向，再展开 Windows 8.3 名称别名。
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
	// 保留授权根的精确名称和资料身份。
	return clean, nil
}

// SortedRoots 按名称稳定排序。
func SortedRoots(roots []string) []string {
	out := append([]string{}, roots...)
	sort.Strings(out)
	return out
}
