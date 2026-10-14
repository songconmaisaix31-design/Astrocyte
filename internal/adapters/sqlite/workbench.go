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
	root             string
	clock            func() time.Time
	zone             *time.Location
	protectedSources []string
}
type workbenchDB struct {
	*sql.DB
	ctx context.Context
}

func (d workbenchDB) query(q string, args ...any) (*sql.Rows, error) {
	return d.QueryContext(d.ctx, q, args...)
}
func (d workbenchDB) row(q string, args ...any) *sql.Row { return d.QueryRowContext(d.ctx, q, args...) }

var workbenchZone = func() *time.Location {
	zone, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}
	return zone
}()

// 保护来源由启动配置明确提供；默认 Reader 只访问自身出版目录。
func NewWorkbenchReader(root string, protectedSources ...string) *WorkbenchReader {
	return &WorkbenchReader{root: root, clock: time.Now, zone: workbenchZone, protectedSources: append([]string(nil), protectedSources...)}
}

func workbenchUnredirected(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if err = workbenchPathUnredirected(abs); err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil || !strings.EqualFold(filepath.Clean(abs), filepath.Clean(resolved)) {
		return "", workbench.ErrUnavailable
	}
	return resolved, nil
}

// 每次请求固定完整的已出版 generation，不执行迁移或打开可写源库。
func (s *WorkbenchReader) open(ctx context.Context) (workbenchDB, workbenchPublication, string, error) {
	var p workbenchPublication
	var empty workbenchDB
	pointer, err := workbenchUnredirected(filepath.Join(s.root, "current.json"))
	if err != nil {
		return empty, p, "", err
	}
	info, err := os.Lstat(pointer)
	if err != nil || !info.Mode().IsRegular() {
		return empty, p, "", workbench.ErrUnavailable
	}
	data, err := os.ReadFile(pointer)
	if err != nil {
		return empty, p, "", err
	}
	if err = json.Unmarshal(data, &p); err != nil {
		return empty, p, "", err
	}
	if p.Generation == "" || filepath.Base(p.Generation) != p.Generation || strings.ContainsAny(p.Generation, "/\\:") || p.Generation == "." || p.Generation == ".." {
		return empty, p, "", workbench.ErrUnavailable
	}
	if _, err = time.Parse(time.RFC3339Nano, p.SnapshotAt); err != nil {
		return empty, p, "", err
	}
	base, err := workbenchUnredirected(filepath.Join(s.root, "snapshots"))
	if err != nil {
		return empty, p, "", err
	}
	dir, err := workbenchUnredirected(filepath.Join(base, p.Generation))
	if err != nil {
		return empty, p, "", err
	}
	file, err := workbenchUnredirected(filepath.Join(dir, "snapshot.sqlite"))
	if err != nil {
		return empty, p, "", err
	}
	snapshot, err := os.Lstat(file)
	if err != nil || !snapshot.Mode().IsRegular() {
		return empty, p, "", workbench.ErrUnavailable
	}
	for _, source := range s.protectedSources {
		if !filepath.IsAbs(source) {
			return empty, p, "", workbench.ErrUnavailable
		}
		original, err := os.Stat(source)
		if err != nil {
			return empty, p, "", err
		}
		if !original.Mode().IsRegular() || os.SameFile(snapshot, original) {
			return empty, p, "", workbench.ErrUnavailable
		}
	}
	uriPath := filepath.ToSlash(file)
	if filepath.VolumeName(file) != "" {
		uriPath = "/" + uriPath
	}
	u := url.URL{Scheme: "file", Path: uriPath}
	query := u.Query()
	query.Set("mode", "ro")
	query.Set("immutable", "1")
	query.Add("_pragma", "query_only(1)")
	u.RawQuery = query.Encode()
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return empty, p, "", err
	}
	db.SetMaxOpenConns(1)
	if err = db.PingContext(ctx); err != nil {
		return empty, p, "", errors.Join(err, db.Close())
	}
	return workbenchDB{db, ctx}, p, dir, nil
}

func (s *WorkbenchReader) Read(ctx context.Context, name string, q workbench.Query) (v workbench.View, err error) {
	// 请求边界只公开固定错误；内部路径、SQL 和原始内容保留在本机。
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
			if _, _, err := s.bounds(day); err != nil {
				return v, workbench.ErrInvalidQuery
			}
		}
	}
	if q.Start != "" && q.End != "" && q.Start > q.End {
		return v, workbench.ErrInvalidQuery
	}
	db, p, dir, err := s.open(ctx)
	if err != nil {
		return v, err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	v.SnapshotAt = p.SnapshotAt
	switch name {
	case "health":
		v.Health = &workbench.Health{Status: "ok", SnapshotAt: p.SnapshotAt, Timezone: "Asia/Shanghai"}
	case "events":
		x, e := s.events(db, q)
		v.Events, err = &x, e
	case "topics/evidence":
		x, e := s.events(db, q)
		if e != nil {
			return v, e
		}
		a, e := s.aggregates(db, q)
		v.Evidence, err = &workbench.Evidence{Topic: q.Topic, Aggregates: a, Events: x}, e
	case "filters":
		x, e := s.filters(db)
		v.Filters, err = &x, e
	case "focus":
		x, e := s.focus(db, dir)
		v.Focus, err = &x, e
	case "today":
		x, e := s.today(db, dir, q.Source)
		v.Today, err = &x, e
	case "agents":
		x, e := s.agents(db)
		v.Agents, err = &x, e
	default:
		err = workbench.ErrInvalidQuery
	}
	return v, err
}

func workbenchStamp(ms int64) string { return time.UnixMilli(ms).UTC().Format(time.RFC3339Nano) }
func (s *WorkbenchReader) bounds(day string) (int64, int64, error) {
	t, err := time.ParseInLocation("2006-01-02", day, s.zone)
	if err != nil {
		return 0, 0, err
	}
	return t.UnixMilli(), t.AddDate(0, 0, 1).UnixMilli(), nil
}
func (s *WorkbenchReader) where(q workbench.Query) (string, []any, error) {
	w, args := "1=1", []any{}
	for _, field := range []struct{ key, value string }{{"topic", q.Topic}, {"kind", q.Kind}, {"source", q.Source}} {
		if field.value != "" {
			w += " AND " + field.key + "=?"
			args = append(args, field.value)
		}
	}
	if q.Start != "" {
		start, _, err := s.bounds(q.Start)
		if err != nil {
			return "", nil, err
		}
		w += " AND ts>=?"
		args = append(args, start)
	}
	if q.End != "" {
		_, end, err := s.bounds(q.End)
		if err != nil {
			return "", nil, err
		}
		w += " AND ts<?"
		args = append(args, end)
	}
	return w, args, nil
}
func workbenchEvents(db workbenchDB, where string, args []any, count, offset int) ([]workbench.Event, error) {
	rows, err := db.query(`SELECT `+workbenchProjection+` FROM attention_engine_events WHERE `+where+` ORDER BY ts DESC,id ASC LIMIT ? OFFSET ?`, append(append([]any{}, args...), count, offset)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []workbench.Event{}
	for rows.Next() {
		var event workbench.Event
		var tokens sql.NullInt64
		var ms int64
		if err = rows.Scan(&event.ID, &event.Kind, &event.Topic, &tokens, &event.Scope, &ms); err != nil {
			return nil, err
		}
		if tokens.Valid {
			event.Tokens = &tokens.Int64
		}
		event.TS = workbenchStamp(ms)
		result = append(result, event)
	}
	return result, rows.Err()
}
func (s *WorkbenchReader) events(db workbenchDB, q workbench.Query) (workbench.Events, error) {
	result := workbench.Events{Page: q.Page, PageSize: q.PageSize}
	where, args, err := s.where(q)
	if err != nil {
		return result, err
	}
	if err = db.row(`SELECT COUNT(*) FROM attention_engine_events WHERE `+where, args...).Scan(&result.Total); err != nil {
		return result, err
	}
	result.Events, err = workbenchEvents(db, where, args, q.PageSize, (q.Page-1)*q.PageSize)
	return result, err
}
func (s *WorkbenchReader) aggregates(db workbenchDB, q workbench.Query) ([]workbench.Aggregate, error) {
	where, args, err := s.where(q)
	if err != nil {
		return nil, err
	}
	rows, err := db.query(`SELECT strftime('%Y-%m-%d',ts/1000.0,'unixepoch','+8 hours'),kind,COUNT(*) FROM attention_engine_events WHERE `+where+` GROUP BY 1,2 ORDER BY 1,2`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []workbench.Aggregate{}
	for rows.Next() {
		var aggregate workbench.Aggregate
		if err = rows.Scan(&aggregate.Day, &aggregate.Kind, &aggregate.Count); err != nil {
			return nil, err
		}
		result = append(result, aggregate)
	}
	return result, rows.Err()
}
func workbenchTopics(db workbenchDB, source string) ([]workbench.Topic, error) {
	where, args := "", []any{}
	if source != "" {
		where = ` WHERE topic IN (SELECT DISTINCT topic FROM attention_engine_events WHERE source=?)`
		args = append(args, source)
	}
	rows, err := db.query(`SELECT topic,attention,momentum,confidence FROM attention_engine_nodes`+where+` ORDER BY attention DESC,topic ASC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []workbench.Topic{}
	for rows.Next() {
		var topic workbench.Topic
		if err = rows.Scan(&topic.Topic, &topic.Attention, &topic.Momentum, &topic.Confidence); err != nil {
			return nil, err
		}
		result = append(result, topic)
	}
	return result, rows.Err()
}
func workbenchHistory(dir string) (map[string][]workbenchPoint, error) {
	file, err := workbenchUnredirected(filepath.Join(dir, "history.json"))
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	result := map[string][]workbenchPoint{}
	if err = json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	for _, points := range result {
		for _, point := range points {
			if point.TS <= 0 || point.Attention < 0 || point.Attention > 1 {
				return nil, workbench.ErrUnavailable
			}
		}
	}
	return result, nil
}
func (s *WorkbenchReader) focus(db workbenchDB, dir string) (workbench.Focus, error) {
	result := workbench.Focus{Series: []workbench.Series{}}
	topics, err := workbenchTopics(db, "")
	if err != nil {
		return result, err
	}
	result.Topics = topics
	history, err := workbenchHistory(dir)
	if err != nil {
		return result, err
	}
	for _, topic := range topics {
		result.Series = append(result.Series, workbenchSeries(topic.Topic, history[topic.Topic]))
	}
	return result, nil
}

func workbenchSeries(topic string, points []workbenchPoint) workbench.Series {
	ordered := append([]workbenchPoint(nil), points...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].TS < ordered[j].TS })
	series := workbench.Series{Topic: topic, Points: []workbench.Point{}}
	for i, point := range ordered {
		if i > 0 && ordered[i-1].TS == point.TS {
			continue
		}
		series.Points = append(series.Points, workbench.Point{TS: workbenchStamp(point.TS), Attention: point.Attention})
	}
	return series
}
func (s *WorkbenchReader) filters(db workbenchDB) (workbench.Filters, error) {
	result := workbench.Filters{}
	for _, projection := range []struct {
		sql  string
		dest *[]string
	}{
		{`SELECT DISTINCT topic FROM attention_engine_events UNION SELECT topic FROM attention_engine_nodes ORDER BY 1`, &result.Topics},
		{`SELECT DISTINCT kind FROM attention_engine_events ORDER BY kind`, &result.Kinds},
		{`SELECT DISTINCT source FROM attention_engine_events ORDER BY source`, &result.Sources},
	} {
		rows, err := db.query(projection.sql)
		if err != nil {
			return result, err
		}
		*projection.dest = []string{}
		for rows.Next() {
			var value string
			if err = rows.Scan(&value); err != nil {
				rows.Close()
				return result, err
			}
			*projection.dest = append(*projection.dest, value)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return result, err
		}
	}
	return result, nil
}
func (s *WorkbenchReader) today(db workbenchDB, dir, source string) (workbench.Today, error) {
	now := s.clock().In(s.zone)
	day, yesterday := now.Format("2006-01-02"), now.AddDate(0, 0, -1).Format("2006-01-02")
	start, end, err := s.bounds(day)
	if err != nil {
		return workbench.Today{}, err
	}
	result := workbench.Today{Day: day, Focus: []workbench.Card{}, Sleep: workbench.Sleep{Day: yesterday, Status: "not_produced"}, DailyCounts: []workbench.DayCount{}, Sources: []workbench.SourceCount{}}
	topics, err := workbenchTopics(db, source)
	if err != nil {
		return result, err
	}
	history, err := workbenchHistory(dir)
	if err != nil {
		return result, err
	}
	// 只比较昨天与今天已确认的实际观测，缺失时 delta 保持 null。
	for _, topic := range topics {
		var ms int64
		if err = db.row(`SELECT updated_at FROM attention_engine_nodes WHERE topic=?`, topic.Topic).Scan(&ms); err != nil {
			return result, err
		}
		var previous *workbenchPoint
		for i, point := range history[topic.Topic] {
			if time.UnixMilli(point.TS).In(s.zone).Format("2006-01-02") == yesterday && (previous == nil || point.TS > previous.TS) {
				p := history[topic.Topic][i]
				previous = &p
			}
		}
		var delta *float64
		if previous != nil && time.UnixMilli(ms).In(s.zone).Format("2006-01-02") == day {
			value := topic.Attention - previous.Attention
			delta = &value
		}
		result.Focus = append(result.Focus, workbench.Card{Topic: topic, Delta: delta})
	}
	where, args, err := s.where(workbench.Query{Source: source, Start: day, End: day})
	if err != nil {
		return result, err
	}
	result.Activity, err = workbenchEvents(db, where, args, 50, 0)
	if err != nil {
		return result, err
	}
	var completed sql.NullInt64
	err = db.row(`SELECT completed_at FROM attention_engine_sleep_runs WHERE day=? ORDER BY started_at DESC,id ASC LIMIT 1`, yesterday).Scan(&completed)
	if err != nil && err != sql.ErrNoRows {
		return result, err
	}
	if err == nil {
		result.Sleep.Status = "running"
		if completed.Valid && completed.Int64 > 0 {
			result.Sleep.Status = "completed"
			value := workbenchStamp(completed.Int64)
			result.Sleep.CompletedAt = &value
		}
	}
	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -i).Format("2006-01-02")
		where, args, err := s.where(workbench.Query{Source: source, Start: day, End: day})
		if err != nil {
			return result, err
		}
		var count int
		if err = db.row(`SELECT COUNT(*) FROM attention_engine_events WHERE `+where, args...).Scan(&count); err != nil {
			return result, err
		}
		result.DailyCounts = append(result.DailyCounts, workbench.DayCount{Day: day, Count: count})
	}
	rows, err := db.query(`SELECT source,COUNT(*) FROM attention_engine_events WHERE ts>=? AND ts<? GROUP BY source ORDER BY source`, start, end)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var source workbench.SourceCount
		if err = rows.Scan(&source.Source, &source.Count); err != nil {
			return result, err
		}
		result.Sources = append(result.Sources, source)
	}
	return result, rows.Err()
}
func (s *WorkbenchReader) agents(db workbenchDB) (workbench.Agents, error) {
	result := workbench.Agents{Groups: []workbench.AgentGroup{}}
	now := s.clock().In(s.zone)
	start, end, err := s.bounds(now.Format("2006-01-02"))
	if err != nil {
		return result, err
	}
	start7, _, err := s.bounds(now.AddDate(0, 0, -6).Format("2006-01-02"))
	if err != nil {
		return result, err
	}
	start30, _, err := s.bounds(now.AddDate(0, 0, -29).Format("2006-01-02"))
	if err != nil {
		return result, err
	}
	rows, err := db.query(`SELECT source,SUM(CASE WHEN ts>=? AND ts<? THEN 1 ELSE 0 END),SUM(CASE WHEN ts>=? AND ts<? THEN 1 ELSE 0 END),SUM(CASE WHEN ts>=? AND ts<? THEN 1 ELSE 0 END) FROM attention_engine_events GROUP BY source ORDER BY source`, start, end, start7, end, start30, end)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var group workbench.AgentGroup
		if err = rows.Scan(&group.Source, &group.Today, &group.Days7, &group.Days30); err != nil {
			rows.Close()
			return result, err
		}
		result.Groups = append(result.Groups, group)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, err
	}
	for i := range result.Groups {
		recent, err := workbenchEvents(db, "source=?", []any{result.Groups[i].Source}, 5, 0)
		if err != nil {
			return result, err
		}
		result.Groups[i].Recent = recent
	}
	return result, nil
}
