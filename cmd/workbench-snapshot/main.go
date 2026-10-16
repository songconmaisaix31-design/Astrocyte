package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"modernc.org/sqlite"
)

type publication struct {
	Generation string `json:"generation"`
	SnapshotAt string `json:"snapshot_at"`
}
type point struct {
	TS        int64   `json:"ts"`
	Attention float64 `json:"attention"`
}

func (p *point) UnmarshalJSON(data []byte) error {
	var input struct {
		TS        *int64   `json:"ts"`
		Attention *float64 `json:"attention"`
	}
	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}
	if input.TS == nil || input.Attention == nil {
		return errors.New("observed history fields required")
	}
	p.TS, p.Attention = *input.TS, *input.Attention
	return nil
}

type history map[string][]point

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, `{"status":"error","code":"SNAPSHOT_UNAVAILABLE"}`)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, out, stderr io.Writer) error {
	f := flag.NewFlagSet("workbench-snapshot", flag.ContinueOnError)
	f.SetOutput(stderr)
	source := f.String("source", "", "explicit source SQLite file")
	root := f.String("root", "", "independent snapshot directory")
	hist := f.String("history", "", "optional explicit actual historical state JSON")
	deadline := f.String("timeout", "", "required positive duration, for example 120s")
	if err := f.Parse(args); err != nil {
		return err
	}
	duration, err := time.ParseDuration(*deadline)
	if err != nil || duration <= 0 || *source == "" || *root == "" || f.NArg() != 0 {
		return errors.New("explicit source, root and positive timeout required")
	}
	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	p, err := publish(ctx, *source, *root, *hist)
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(struct {
		Status string `json:"status"`
		publication
	}{"ok", p})
}

func uri(path string, readonly bool) string {
	p := filepath.ToSlash(path)
	if filepath.VolumeName(path) != "" {
		p = "/" + p
	}
	u := url.URL{Scheme: "file", Path: p}
	q := u.Query()
	if readonly {
		q.Set("mode", "ro")
		q.Add("_pragma", "query_only(1)")
	} else {
		q.Set("mode", "rw")
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func canonical(path string) (string, error) {
	p, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(p)
}

// 检查已存在的祖先，禁止目的目录通过符号链接指向其他写入范围。
func ownedPath(path string) error {
	for p := path; ; p = filepath.Dir(p) {
		_, err := os.Lstat(p)
		if err == nil {
			if err = rejectAlias(p); err != nil {
				return err
			}
			real, err := filepath.EvalSymlinks(p)
			if err != nil {
				return err
			}
			if real != p && !samePath(real, p) {
				return errors.New("destination alias refused")
			}
			return nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if filepath.Dir(p) == p {
			return errors.New("destination ancestor unavailable")
		}
	}
}

var pathEqual = func(a, b string) bool { return a == b }

func samePath(a, b string) bool { return pathEqual(filepath.Clean(a), filepath.Clean(b)) }
func inside(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// 比较实际文件身份，避免 hardlink 或已有 generation 覆盖来源及其 sidecar。
func inspectRoot(ctx context.Context, root string, protected []os.FileInfo) error {
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if e := ctx.Err(); e != nil {
			return e
		}
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return errors.New("destination symlink refused")
		}
		if err = rejectAlias(path); err != nil {
			return err
		}
		real, err := filepath.EvalSymlinks(path)
		if err != nil || !samePath(real, path) {
			return errors.New("destination alias refused")
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		for _, src := range protected {
			if os.SameFile(info, src) {
				return errors.New("source identity overlaps destination")
			}
		}
		return nil
	})
}

func loadJSON(path string, target any) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	d := json.NewDecoder(io.LimitReader(f, 32<<20))
	if err = d.Decode(target); err != nil {
		return err
	}
	var extra any
	if err = d.Decode(&extra); err != io.EOF {
		return errors.New("invalid JSON tail")
	}
	return nil
}
func validateHistory(h history) error {
	if h == nil {
		return errors.New("history object required")
	}
	for topic, points := range h {
		if topic == "" || points == nil {
			return errors.New("invalid history topic")
		}
		for _, p := range points {
			if p.TS <= 0 || math.IsNaN(p.Attention) || math.IsInf(p.Attention, 0) || p.Attention < 0 || p.Attention > 1 {
				return errors.New("invalid history point")
			}
		}
	}
	return nil
}
func mergeHistory(dst, src history) {
	for topic, points := range src {
		known := map[int64]bool{}
		for _, p := range dst[topic] {
			known[p.TS] = true
		}
		for _, p := range points {
			if !known[p.TS] {
				dst[topic] = append(dst[topic], p)
				known[p.TS] = true
			}
		}
	}
}

func writeJSON(path string, value any) (err error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, f.Close()) }()
	if err = json.NewEncoder(f).Encode(value); err != nil {
		return err
	}
	return f.Sync()
}

func publish(ctx context.Context, source, root, externalHistory string) (p publication, err error) {
	if err = ctx.Err(); err != nil {
		return p, err
	}
	source, err = canonical(source)
	if err != nil {
		return p, err
	}
	srcInfo, err := os.Stat(source)
	if err != nil || !srcInfo.Mode().IsRegular() {
		return p, errors.New("source regular file required")
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return p, err
	}
	if inside(source, root) {
		return p, errors.New("source overlaps destination")
	}
	if err = ownedPath(root); err != nil {
		return p, err
	}
	protected := []os.FileInfo{srcInfo}
	for _, suffix := range []string{"-wal", "-shm"} {
		st, e := os.Stat(source + suffix)
		if e == nil {
			protected = append(protected, st)
		} else if !errors.Is(e, os.ErrNotExist) {
			return p, e
		}
	}
	h := history{}
	if externalHistory != "" {
		externalHistory, err = canonical(externalHistory)
		if err != nil {
			return p, err
		}
		st, e := os.Stat(externalHistory)
		if e != nil {
			return p, e
		}
		protected = append(protected, st)
		if err = loadJSON(externalHistory, &h); err != nil {
			return p, err
		}
		if err = validateHistory(h); err != nil {
			return p, err
		}
	}
	if err = os.MkdirAll(root, 0700); err != nil {
		return p, err
	}
	if err = inspectRoot(ctx, root, protected); err != nil {
		return p, err
	}
	lock, err := os.OpenFile(filepath.Join(root, "snapshot.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return p, err
	}
	defer lock.Close()
	if err = lockFile(lock); err != nil {
		return p, err
	}
	defer unlockFile(lock)
	current := filepath.Join(root, "current.json")
	var old publication
	if err = loadJSON(current, &old); err == nil {
		if old.Generation == "" || old.Generation == "." || old.Generation == ".." || filepath.Base(old.Generation) != old.Generation || strings.ContainsAny(old.Generation, "/\\:") {
			return p, errors.New("invalid generation")
		}
		if _, err = time.Parse(time.RFC3339Nano, old.SnapshotAt); err != nil {
			return p, err
		}
		previous := history{}
		if err = loadJSON(filepath.Join(root, "snapshots", old.Generation, "history.json"), &previous); err != nil {
			return p, err
		}
		if err = validateHistory(previous); err != nil {
			return p, err
		}
		mergeHistory(h, previous)
	} else if !errors.Is(err, os.ErrNotExist) {
		return p, err
	}
	if err = os.MkdirAll(filepath.Join(root, "snapshots"), 0700); err != nil {
		return p, err
	}
	generation, err := os.MkdirTemp(filepath.Join(root, "snapshots"), "snapshot-")
	if err != nil {
		return p, err
	}
	dest := filepath.Join(generation, "snapshot.sqlite")
	if err = backup(ctx, source, dest); err != nil {
		return p, err
	}
	db, err := sql.Open("sqlite", uri(dest, false))
	if err != nil {
		return p, err
	}
	db.SetMaxOpenConns(1)
	defer db.Close()
	var check, mode string
	if err = db.QueryRowContext(ctx, "PRAGMA quick_check").Scan(&check); err != nil || check != "ok" {
		return p, errors.New("invalid backup")
	}
	if err = db.QueryRowContext(ctx, "PRAGMA journal_mode=DELETE").Scan(&mode); err != nil || mode != "delete" {
		return p, errors.New("backup journal mode unavailable")
	}
	if err = validateDatabase(ctx, db); err != nil {
		return p, err
	}
	rows, err := db.QueryContext(ctx, "SELECT topic,updated_at,attention FROM attention_engine_nodes ORDER BY topic")
	if err != nil {
		return p, err
	}
	actual := history{}
	for rows.Next() {
		var topic string
		var pt point
		if err = rows.Scan(&topic, &pt.TS, &pt.Attention); err != nil {
			rows.Close()
			return p, err
		}
		if pt.TS > 0 {
			actual[topic] = []point{pt}
		}
	}
	err = errors.Join(rows.Err(), rows.Close())
	if err != nil {
		return p, err
	}
	if err = validateHistory(actual); err != nil {
		return p, err
	}
	mergeHistory(h, actual)
	for _, points := range h {
		sort.SliceStable(points, func(i, j int) bool { return points[i].TS < points[j].TS })
	}
	if err = db.Close(); err != nil {
		return p, err
	}
	if err = writeJSON(filepath.Join(generation, "history.json"), h); err != nil {
		return p, err
	}
	snapshotFile, err := os.OpenFile(dest, os.O_RDWR, 0)
	if err != nil {
		return p, err
	}
	err = errors.Join(snapshotFile.Sync(), snapshotFile.Close())
	if err != nil {
		return p, err
	}
	p = publication{filepath.Base(generation), time.Now().UTC().Format(time.RFC3339Nano)}
	if err = inspectRoot(ctx, root, protected); err != nil {
		return p, err
	}
	if now, e := os.Stat(source); e != nil || !os.SameFile(now, srcInfo) {
		return p, errors.New("source identity changed")
	}
	if err = ctx.Err(); err != nil {
		return p, err
	}
	temp, err := os.CreateTemp(root, "publication-*.json")
	if err != nil {
		return p, err
	}
	tempName := temp.Name()
	err = json.NewEncoder(temp).Encode(p)
	if err == nil {
		err = temp.Sync()
	}
	err = errors.Join(err, temp.Close())
	if err != nil {
		return p, err
	}
	if err = ctx.Err(); err != nil {
		return p, err
	}
	err = replaceFile(tempName, current)
	return p, err
}

func validateDatabase(ctx context.Context, db *sql.DB) error {
	checks := []struct {
		query string
		scan  func(*sql.Rows) error
	}{
		{`SELECT id,kind,topic,source,ts,CASE WHEN json_type(data,'$.tokens')='integer' THEN json_extract(data,'$.tokens') END,CASE WHEN json_type(data,'$.scope')='text' THEN json_extract(data,'$.scope') ELSE '' END FROM attention_engine_events`, func(r *sql.Rows) error {
			var id, kind, topic, source, scope string
			var ts int64
			var tokens sql.NullInt64
			return r.Scan(&id, &kind, &topic, &source, &ts, &tokens, &scope)
		}},
		{`SELECT topic,attention,momentum,confidence,updated_at FROM attention_engine_nodes`, func(r *sql.Rows) error {
			var topic string
			var a, m, c float64
			var ts int64
			if err := r.Scan(&topic, &a, &m, &c, &ts); err != nil {
				return err
			}
			if math.IsNaN(a) || math.IsInf(a, 0) || a < 0 || a > 1 || math.IsNaN(m) || math.IsInf(m, 0) || math.IsNaN(c) || math.IsInf(c, 0) || c < 0 || c > 1 {
				return errors.New("invalid node scalars")
			}
			return nil
		}},
		{`SELECT day,completed_at,started_at,id FROM attention_engine_sleep_runs`, func(r *sql.Rows) error {
			var day, id string
			var completed sql.NullInt64
			var started int64
			return r.Scan(&day, &completed, &started, &id)
		}},
	}
	for _, check := range checks {
		r, err := db.QueryContext(ctx, check.query)
		if err != nil {
			return err
		}
		for r.Next() {
			if err = check.scan(r); err != nil {
				r.Close()
				return err
			}
		}
		err = errors.Join(r.Err(), r.Close())
		if err != nil {
			return err
		}
	}
	return nil
}

func backup(ctx context.Context, source, dest string) (err error) {
	db, err := sql.Open("sqlite", uri(source, true))
	if err != nil {
		return err
	}
	defer db.Close()
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err = conn.ExecContext(ctx, "BEGIN"); err != nil {
		return err
	}
	defer conn.ExecContext(context.Background(), "ROLLBACK")
	var count int
	if err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master").Scan(&count); err != nil {
		return err
	}
	return conn.Raw(func(raw any) (err error) {
		b, err := raw.(interface {
			NewBackup(string) (*sqlite.Backup, error)
		}).NewBackup(dest)
		if err != nil {
			return err
		}
		defer func() { err = errors.Join(err, b.Finish()) }()
		for {
			if err = ctx.Err(); err != nil {
				return err
			}
			more, e := b.Step(256)
			if e != nil {
				return e
			}
			if !more {
				return ctx.Err()
			}
		}
	})
}
