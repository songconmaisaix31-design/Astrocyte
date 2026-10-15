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

// 认知快照可使用独立只读根，应用数据和 Relay 根仍由原配置选择。
func workbenchReaderRoot() (string, error) {
	value := os.Getenv("ASTROCYTE_WORKBENCH_READER_ROOT")
	if value == "" {
		return workbenchRoot(), nil
	}
	if !filepath.IsAbs(value) || strings.TrimSpace(value) != value {
		return "", errors.New("workbench reader root must be an explicit absolute directory")
	}
	canonical, err := workbenchCanonical(value)
	if err != nil || !strings.EqualFold(filepath.Clean(value), filepath.Clean(canonical)) {
		return "", errors.New("workbench reader root is unavailable or redirected")
	}
	info, err := os.Stat(canonical)
	if err != nil || !info.IsDir() {
		return "", errors.New("workbench reader root is unavailable")
	}
	return canonical, nil
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

// 先解析已有祖先及 junction，应用数据库不能使用来源或已发布副本。
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

// 接入入口已选定独立目录；复用保护检查并保持该目录选择。
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
	readerRoot, e := workbenchReaderRoot()
	if e != nil {
		return "", e
	}
	directories := []string{filepath.Join(workbenchRoot(), "snapshots"), filepath.Join(readerRoot, "snapshots")}
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
		// 同时核对已有 8.3 别名和祖先目录的实际身份。
		a, e := workbenchExistingInfo(base)
		if e != nil {
			return "", errors.New("protected workbench directory identity unavailable")
		}
		if a != nil {
			for p := resolved; ; p = filepath.Dir(p) {
				b, e := workbenchExistingInfo(p)
				if e != nil {
					return "", errors.New("application directory identity unavailable")
				}
				if b != nil && os.SameFile(a, b) {
					return "", errors.New("application data directory must be separate from workbench source and snapshots")
				}
				if filepath.Dir(p) == p {
					break
				}
			}
		}
	}
	appDB, e := workbenchExistingInfo(filepath.Join(resolved, "state.sqlite"))
	if e != nil {
		return "", errors.New("application database identity unavailable")
	}
	if appDB != nil {
		for _, path := range sources {
			source, e := workbenchExistingInfo(path)
			if e != nil {
				return "", errors.New("protected workbench source identity unavailable")
			}
			if source != nil && os.SameFile(appDB, source) {
				return "", errors.New("application database must be independent")
			}
		}
		for _, snapshotRoot := range []string{filepath.Join(workbenchRoot(), "snapshots"), filepath.Join(readerRoot, "snapshots")} {
			generations, e := os.ReadDir(snapshotRoot)
			if e != nil && !os.IsNotExist(e) {
				return "", errors.New("workbench publication directory unavailable")
			}
			for _, generation := range generations {
				generationPath := filepath.Join(snapshotRoot, generation.Name())
				canonical, e := workbenchCanonical(generationPath)
				if e != nil || !strings.EqualFold(filepath.Clean(generationPath), filepath.Clean(canonical)) {
					return "", errors.New("workbench snapshot directory redirected")
				}
				info, e := workbenchExistingInfo(generationPath)
				if e != nil || info == nil {
					return "", errors.New("workbench snapshot directory identity unavailable")
				}
				if !info.IsDir() {
					continue
				}
				snapshot, e := workbenchExistingInfo(filepath.Join(generationPath, "snapshot.sqlite"))
				if e != nil {
					return "", errors.New("workbench snapshot identity unavailable")
				}
				if snapshot != nil && os.SameFile(appDB, snapshot) {
					return "", errors.New("application database must be independent")
				}
			}
		}
	}
	return dir, nil
}

// 首次启动允许目标尚未创建；其他文件系统错误必须立即返回。
func workbenchExistingInfo(path string) (os.FileInfo, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	return info, err
}
