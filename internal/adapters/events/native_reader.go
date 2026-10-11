package events

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const maxReplayBytes = 1024 * 1024
const maxReplayRecords = 128

type nativeLine struct {
	line        int
	raw         []byte
	observation nativeObservation
}

// Read a selected file under its explicitly registered root. Never walk a
// private history directory. Bounds stop the read and are reported to callers.
func readNativeFile(ctx context.Context, source, root, path, project string, limit int) ([]nativeLine, bool, error) {
	if limit < 1 || limit > maxReplayRecords {
		return nil, false, fmt.Errorf("replay limit must be 1..128")
	}
	r, err := filepath.Abs(root)
	if err != nil {
		return nil, false, err
	}
	p, err := filepath.Abs(path)
	if err != nil {
		return nil, false, err
	}
	r, err = unredirectedReplayPath(r)
	if err != nil {
		return nil, false, err
	}
	selected, err := unredirectedReplayPath(p)
	if err != nil {
		return nil, false, err
	}
	rel, err := filepath.Rel(r, selected)
	if err != nil || rel == ".." || filepath.IsAbs(rel) || len(rel) > 3 && rel[:3] == ".."+string(filepath.Separator) {
		return nil, false, fmt.Errorf("replay file leaves selected root")
	}
	f, err := os.Open(p)
	if err != nil {
		return nil, false, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return nil, false, fmt.Errorf("replay source must be a regular file")
	}
	return readNative(ctx, source, project, io.LimitReader(f, maxReplayBytes), limit, info.Size() > maxReplayBytes)
}

func readNative(ctx context.Context, source, project string, r io.Reader, limit int, truncated bool) ([]nativeLine, bool, error) {
	scan := bufio.NewScanner(r)
	scan.Buffer(make([]byte, 4096), maxReplayBytes)
	lines := []nativeLine{}
	association := NativeAssociation{}
	for scan.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, false, err
		}
		if len(lines) == limit {
			return lines, true, nil
		}
		raw := append([]byte(nil), scan.Bytes()...)
		o, err := nativeRecord(source, raw)
		if err != nil {
			return nil, false, fmt.Errorf("replay record %d: %w", len(lines)+1, err)
		}
		if len(lines) == 0 {
			if o.session == "" || project == "" || !sameReplayProject(o.cwd, project) {
				return nil, false, fmt.Errorf("replay header does not identify approved project and session")
			}
		}
		association, err = ObserveNativeAssociation(source, project, association, raw)
		if err != nil {
			return nil, false, err
		}
		o.session, o.cwd = association.Session, project
		lines = append(lines, nativeLine{line: len(lines) + 1, raw: raw, observation: o})
	}
	if err := scan.Err(); err != nil {
		return lines, true, fmt.Errorf("replay read bound or input error: %w", err)
	}
	return lines, truncated, nil
}
