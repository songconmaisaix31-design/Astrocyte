package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	_ "time/tzdata"

	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
)

const WorkbenchFormalSource = "C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v4-w1/formal-20261013-first/state.sqlite"
const workbenchProjection = `id,kind,topic,CASE WHEN json_type(data,'$.tokens')='integer' THEN json_extract(data,'$.tokens') END,CASE WHEN json_type(data,'$.scope')='text' THEN json_extract(data,'$.scope') ELSE '' END,ts`

type workbenchPublication struct {
	Generation string `json:"generation"`
	SnapshotAt string `json:"snapshot_at"`
}
type workbenchPoint struct {
	TS        int64   `json:"ts"`
	Attention float64 `json:"attention"`
}
type WorkbenchReader struct {
	root  string
	clock func() time.Time
	zone  *time.Location
}
type workbenchDB struct {
	*sql.DB
	ctx context.Context
}

func (d workbenchDB) query(q string, args ...any) (*sql.Rows, error) {
	return d.QueryContext(d.ctx, q, args...)
}
func (d workbenchDB) row(q string, args ...any) *sql.Row { return d.QueryRowContext(d.ctx, q, args...) }
func NewWorkbenchReader(root string) *WorkbenchReader {
	z, _ := time.LoadLocation("Asia/Shanghai")
	return &WorkbenchReader{root, time.Now, z}
}

// Every request pins a complete published generation. No migration or mutable
// database constructor is called. Atomic pointer changes affect the next read.
func (s *WorkbenchReader) open(ctx context.Context) (workbenchDB, workbenchPublication, string, error) {
	var p workbenchPublication
	var empty workbenchDB
	b, e := os.ReadFile(filepath.Join(s.root, "current.json"))
	if e != nil {
		return empty, p, "", e
	}
	if json.Unmarshal(b, &p) != nil {
		return empty, p, "", workbench.ErrUnavailable
	}
	if p.Generation == "" || filepath.Base(p.Generation) != p.Generation || strings.ContainsAny(p.Generation, "/\\:") || p.Generation == "." || p.Generation == ".." {
		return empty, p, "", workbench.ErrUnavailable
	}
	if _, e = time.Parse(time.RFC3339Nano, p.SnapshotAt); e != nil {
		return empty, p, "", e
	}
	base, e := filepath.EvalSymlinks(filepath.Join(s.root, "snapshots"))
	if e != nil {
		return empty, p, "", e
	}
	dir := filepath.Join(base, p.Generation)
	real, e := filepath.EvalSymlinks(dir)
	if e != nil || !strings.EqualFold(real, dir) {
		return empty, p, "", workbench.ErrUnavailable
	}
	file := filepath.Join(dir, "snapshot.sqlite")
	real, e = filepath.EvalSymlinks(file)
	if e != nil || !strings.EqualFold(real, file) {
		return empty, p, "", workbench.ErrUnavailable
	}
	st, e := os.Stat(file)
	if e != nil || !st.Mode().IsRegular() {
		return empty, p, "", workbench.ErrUnavailable
	}
	if src, e := os.Stat(WorkbenchFormalSource); e == nil && os.SameFile(st, src) {
		return empty, p, "", workbench.ErrUnavailable
	}
	uriPath := filepath.ToSlash(real)
	if filepath.VolumeName(real) != "" {
		uriPath = "/" + uriPath
	}
	u := url.URL{Scheme: "file", Path: uriPath}
	q := u.Query()
	q.Set("mode", "ro")
	q.Set("immutable", "1")
	q.Add("_pragma", "query_only(1)")
	u.RawQuery = q.Encode()
	db, e := sql.Open("sqlite", u.String())
	if e != nil {
		return empty, p, "", e
	}
	db.SetMaxOpenConns(1)
	if e = db.PingContext(ctx); e != nil {
		db.Close()
		return empty, p, "", e
	}
	return workbenchDB{db, ctx}, p, dir, nil
}
func (s *WorkbenchReader) Read(ctx context.Context, name string, q workbench.Query) (v workbench.View, err error) {
	// Collapse all internal paths/SQL/payload errors before crossing app boundary.
	defer func() {
		if err != nil && !errors.Is(err, workbench.ErrInvalidQuery) {
			err = workbench.ErrUnavailable
		}
	}()
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 50
	}
	if q.Page < 1 || q.Page > 10000000 || q.PageSize < 1 || q.PageSize > 100 {
		return v, workbench.ErrInvalidQuery
	}
	for _, day := range []string{q.Start, q.End} {
		if day != "" {
			if _, _, e := s.bounds(day); e != nil {
				return v, workbench.ErrInvalidQuery
			}
		}
	}
	if q.Start != "" && q.End != "" && q.Start > q.End {
		return v, workbench.ErrInvalidQuery
	}
	db, p, dir, e := s.open(ctx)
	if e != nil {
		return v, e
	}
	defer db.Close()
	v.SnapshotAt = p.SnapshotAt
	switch name {
	case "health":
		v.Health = &workbench.Health{Status: "ok", SnapshotAt: p.SnapshotAt, Timezone: "Asia/Shanghai"}
	case "events":
		x, e := s.events(db, q)
		v.Events = &x
		err = e
	case "topics/evidence":
		x, e := s.events(db, q)
		if e != nil {
			return v, e
		}
		a, e := s.aggregates(db, q)
		v.Evidence = &workbench.Evidence{Topic: q.Topic, Aggregates: a, Events: x}
		err = e
	case "filters":
		x, e := s.filters(db)
		v.Filters = &x
		err = e
	case "focus":
		x, e := s.focus(db, dir)
		v.Focus = &x
		err = e
	case "today":
		x, e := s.today(db, dir, q.Source)
		v.Today = &x
		err = e
	case "agents":
		x, e := s.agents(db)
		v.Agents = &x
		err = e
	default:
		err = workbench.ErrInvalidQuery
	}
	return v, err
}
func workbenchStamp(ms int64) string { return time.UnixMilli(ms).UTC().Format(time.RFC3339Nano) }
func (s *WorkbenchReader) bounds(day string) (int64, int64, error) {
	t, e := time.ParseInLocation("2006-01-02", day, s.zone)
	if e != nil {
		return 0, 0, e
	}
	return t.UnixMilli(), t.AddDate(0, 0, 1).UnixMilli(), nil
}
func (s *WorkbenchReader) where(q workbench.Query) (string, []any) {
	w := "1=1"
	a := []any{}
	for _, p := range []struct{ k, v string }{{"topic", q.Topic}, {"kind", q.Kind}, {"source", q.Source}} {
		if p.v != "" {
			w += " AND " + p.k + "=?"
			a = append(a, p.v)
		}
	}
	if q.Start != "" {
		start, _, _ := s.bounds(q.Start)
		w += " AND ts>=?"
		a = append(a, start)
	}
	if q.End != "" {
		_, end, _ := s.bounds(q.End)
		w += " AND ts<?"
		a = append(a, end)
	}
	return w, a
}
func workbenchEvents(db workbenchDB, w string, args []any, n, offset int) ([]workbench.Event, error) {
	rows, e := db.query(`SELECT `+workbenchProjection+` FROM attention_engine_events WHERE `+w+` ORDER BY ts DESC,id ASC LIMIT ? OFFSET ?`, append(append([]any{}, args...), n, offset)...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []workbench.Event{}
	for rows.Next() {
		var x workbench.Event
		var n sql.NullInt64
		var ms int64
		if e = rows.Scan(&x.ID, &x.Kind, &x.Topic, &n, &x.Scope, &ms); e != nil {
			return nil, e
		}
		if n.Valid {
			x.Tokens = &n.Int64
		}
		x.TS = workbenchStamp(ms)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (s *WorkbenchReader) events(db workbenchDB, q workbench.Query) (workbench.Events, error) {
	out := workbench.Events{Page: q.Page, PageSize: q.PageSize}
	w, a := s.where(q)
	if e := db.row(`SELECT COUNT(*) FROM attention_engine_events WHERE `+w, a...).Scan(&out.Total); e != nil {
		return out, e
	}
	ev, e := workbenchEvents(db, w, a, q.PageSize, (q.Page-1)*q.PageSize)
	out.Events = ev
	return out, e
}
func (s *WorkbenchReader) aggregates(db workbenchDB, q workbench.Query) ([]workbench.Aggregate, error) {
	w, a := s.where(q)
	rows, e := db.query(`SELECT strftime('%Y-%m-%d',ts/1000.0,'unixepoch','+8 hours'),kind,COUNT(*) FROM attention_engine_events WHERE `+w+` GROUP BY 1,2 ORDER BY 1,2`, a...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []workbench.Aggregate{}
	for rows.Next() {
		var x workbench.Aggregate
		if e = rows.Scan(&x.Day, &x.Kind, &x.Count); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func workbenchTopics(db workbenchDB, source string) ([]workbench.Topic, error) {
	w := ""
	a := []any{}
	if source != "" {
		w = ` WHERE topic IN (SELECT DISTINCT topic FROM attention_engine_events WHERE source=?)`
		a = append(a, source)
	}
	rows, e := db.query(`SELECT topic,attention,momentum,confidence FROM attention_engine_nodes`+w+` ORDER BY attention DESC,topic ASC`, a...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []workbench.Topic{}
	for rows.Next() {
		var x workbench.Topic
		if e = rows.Scan(&x.Topic, &x.Attention, &x.Momentum, &x.Confidence); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func workbenchHistory(dir string) (map[string][]workbenchPoint, error) {
	f := filepath.Join(dir, "history.json")
	real, e := filepath.EvalSymlinks(f)
	if e != nil || !strings.EqualFold(real, f) {
		return nil, workbench.ErrUnavailable
	}
	b, e := os.ReadFile(f)
	if e != nil {
		return nil, e
	}
	out := map[string][]workbenchPoint{}
	e = json.Unmarshal(b, &out)
	if e != nil {
		return nil, e
	}
	for _, points := range out {
		for _, p := range points {
			if p.TS <= 0 || p.Attention < 0 || p.Attention > 1 {
				return nil, workbench.ErrUnavailable
			}
		}
	}
	return out, nil
}
func (s *WorkbenchReader) focus(db workbenchDB, dir string) (workbench.Focus, error) {
	out := workbench.Focus{Series: []workbench.Series{}}
	ts, e := workbenchTopics(db, "")
	if e != nil {
		return out, e
	}
	out.Topics = ts
	hist, e := workbenchHistory(dir)
	if e != nil {
		return out, e
	}
	for _, t := range ts {
		pts := hist[t.Topic]
		sort.SliceStable(pts, func(i, j int) bool { return pts[i].TS < pts[j].TS })
		ss := workbench.Series{Topic: t.Topic, Points: []workbench.Point{}}
		for i, p := range pts {
			if i > 0 && pts[i-1].TS == p.TS {
				continue
			}
			ss.Points = append(ss.Points, workbench.Point{TS: workbenchStamp(p.TS), Attention: p.Attention})
		}
		out.Series = append(out.Series, ss)
	}
	return out, nil
}
func (s *WorkbenchReader) filters(db workbenchDB) (workbench.Filters, error) {
	out := workbench.Filters{}
	for _, p := range []struct {
		sql  string
		dest *[]string
	}{{`SELECT DISTINCT topic FROM attention_engine_events UNION SELECT topic FROM attention_engine_nodes ORDER BY 1`, &out.Topics}, {`SELECT DISTINCT kind FROM attention_engine_events ORDER BY kind`, &out.Kinds}, {`SELECT DISTINCT source FROM attention_engine_events ORDER BY source`, &out.Sources}} {
		rows, e := db.query(p.sql)
		if e != nil {
			return out, e
		}
		*p.dest = []string{}
		for rows.Next() {
			var x string
			if e = rows.Scan(&x); e != nil {
				rows.Close()
				return out, e
			}
			*p.dest = append(*p.dest, x)
		}
		e = rows.Err()
		rows.Close()
		if e != nil {
			return out, e
		}
	}
	return out, nil
}
func (s *WorkbenchReader) today(db workbenchDB, dir, source string) (workbench.Today, error) {
	now := s.clock().In(s.zone)
	day := now.Format("2006-01-02")
	yday := now.AddDate(0, 0, -1).Format("2006-01-02")
	start, end, _ := s.bounds(day)
	out := workbench.Today{Day: day, Focus: []workbench.Card{}, Sleep: workbench.Sleep{Day: yday, Status: "not_produced"}, DailyCounts: []workbench.DayCount{}, Sources: []workbench.SourceCount{}}
	ts, e := workbenchTopics(db, source)
	if e != nil {
		return out, e
	}
	hist, e := workbenchHistory(dir)
	if e != nil {
		return out, e
	}
	// Compare current confirmed state with last confirmed point on the previous
	// Shanghai day. Current state must itself belong to today, else delta is null.
	for _, t := range ts {
		var ms int64
		if e = db.row(`SELECT updated_at FROM attention_engine_nodes WHERE topic=?`, t.Topic).Scan(&ms); e != nil {
			return out, e
		}
		var previous *workbenchPoint
		for i, p := range hist[t.Topic] {
			if time.UnixMilli(p.TS).In(s.zone).Format("2006-01-02") == yday && (previous == nil || p.TS > previous.TS) {
				x := hist[t.Topic][i]
				previous = &x
			}
		}
		var delta *float64
		if previous != nil && time.UnixMilli(ms).In(s.zone).Format("2006-01-02") == day {
			v := t.Attention - previous.Attention
			delta = &v
		}
		out.Focus = append(out.Focus, workbench.Card{Topic: t, Delta: delta})
	}
	q := workbench.Query{Source: source, Start: day, End: day}
	w, a := s.where(q)
	out.Activity, e = workbenchEvents(db, w, a, 50, 0)
	if e != nil {
		return out, e
	}
	var completed sql.NullInt64
	e = db.row(`SELECT completed_at FROM attention_engine_sleep_runs WHERE day=? ORDER BY started_at DESC,id ASC LIMIT 1`, yday).Scan(&completed)
	if e != nil && e != sql.ErrNoRows {
		return out, e
	}
	if e == nil {
		out.Sleep.Status = "running"
		if completed.Valid && completed.Int64 > 0 {
			out.Sleep.Status = "completed"
			v := workbenchStamp(completed.Int64)
			out.Sleep.CompletedAt = &v
		}
	}
	for i := 6; i >= 0; i-- {
		d := now.AddDate(0, 0, -i).Format("2006-01-02")
		w, a := s.where(workbench.Query{Source: source, Start: d, End: d})
		var c int
		if e = db.row(`SELECT COUNT(*) FROM attention_engine_events WHERE `+w, a...).Scan(&c); e != nil {
			return out, e
		}
		out.DailyCounts = append(out.DailyCounts, workbench.DayCount{Day: d, Count: c})
	}
	rows, e := db.query(`SELECT source,COUNT(*) FROM attention_engine_events WHERE ts>=? AND ts<? GROUP BY source ORDER BY source`, start, end)
	if e != nil {
		return out, e
	}
	defer rows.Close()
	for rows.Next() {
		var x workbench.SourceCount
		if e = rows.Scan(&x.Source, &x.Count); e != nil {
			return out, e
		}
		out.Sources = append(out.Sources, x)
	}
	return out, rows.Err()
}
func (s *WorkbenchReader) agents(db workbenchDB) (workbench.Agents, error) {
	out := workbench.Agents{Groups: []workbench.AgentGroup{}}
	d := s.clock().In(s.zone)
	a, b, _ := s.bounds(d.Format("2006-01-02"))
	a7, _, _ := s.bounds(d.AddDate(0, 0, -6).Format("2006-01-02"))
	a30, _, _ := s.bounds(d.AddDate(0, 0, -29).Format("2006-01-02"))
	rows, e := db.query(`SELECT source,SUM(CASE WHEN ts>=? AND ts<? THEN 1 ELSE 0 END),SUM(CASE WHEN ts>=? AND ts<? THEN 1 ELSE 0 END),SUM(CASE WHEN ts>=? AND ts<? THEN 1 ELSE 0 END) FROM attention_engine_events GROUP BY source ORDER BY source`, a, b, a7, b, a30, b)
	if e != nil {
		return out, e
	}
	for rows.Next() {
		var x workbench.AgentGroup
		if e = rows.Scan(&x.Source, &x.Today, &x.Days7, &x.Days30); e != nil {
			rows.Close()
			return out, e
		}
		out.Groups = append(out.Groups, x)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return out, e
	}
	for i := range out.Groups {
		ev, e := workbenchEvents(db, "source=?", []any{out.Groups[i].Source}, 5, 0)
		if e != nil {
			return out, e
		}
		out.Groups[i].Recent = ev
	}
	return out, nil
}
