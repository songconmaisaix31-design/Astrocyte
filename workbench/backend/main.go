package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	_ "modernc.org/sqlite"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"
)

type Config struct {
	Root string `json:"root"`
	SPA  string `json:"spa"`
	Port int    `json:"port"`
}
type Publication struct {
	Generation string `json:"generation"`
	SnapshotAt string `json:"snapshot_at"`
}
type Event struct {
	ID     string `json:"event_id"`
	Kind   string `json:"kind"`
	Topic  string `json:"topic"`
	Tokens *int64 `json:"tokens"`
	Scope  string `json:"scope"`
	TS     string `json:"ts"`
}
type Topic struct {
	Topic      string  `json:"topic"`
	Attention  float64 `json:"attention"`
	Momentum   float64 `json:"momentum"`
	Confidence float64 `json:"confidence"`
}
type Card struct {
	Topic
	Delta *float64 `json:"attention_delta"`
}
type Point struct {
	TS        string  `json:"ts"`
	Attention float64 `json:"attention"`
}
type Series struct {
	Topic  string  `json:"topic"`
	Points []Point `json:"points"`
}
type HistoryPoint struct {
	TS        int64   `json:"ts"`
	Attention float64 `json:"attention"`
}
type Server struct {
	config Config
	now    func() time.Time
	zone   *time.Location
}

var errInput = errors.New("invalid request")

const projection = `id,kind,topic,CASE WHEN json_type(data,'$.tokens') IN ('integer','real') THEN json_extract(data,'$.tokens') END,COALESCE(json_extract(data,'$.scope'),''),ts`

func localRoot() string {
	return filepath.Join(os.Getenv("LOCALAPPDATA"), "astrocyte-workbench", "v0.1")
}
func main() {
	// A local config can select SPA and port, never a source database.
	b, e := os.ReadFile(filepath.Join(localRoot(), "config.json"))
	if e != nil {
		log.Fatal("configuration unavailable")
	}
	var c Config
	if json.Unmarshal(bytes.TrimPrefix(b, []byte{239,187,191}), &c) != nil {
		log.Fatal("configuration invalid")
	}
	c.Root = localRoot()
	if c.Port < 1 || c.Port > 65535 {
		log.Fatal("port invalid")
	}
	z, e := time.LoadLocation("Asia/Shanghai")
	if e != nil {
		log.Fatal("timezone unavailable")
	}
	s := &Server{c, time.Now, z}
	l, e := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", c.Port))
	if e != nil {
		log.Fatal("listener unavailable")
	}
	log.Print("workbench listening on loopback")
	h := &http.Server{Handler: s, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 8192}
	if h.Serve(l) != nil {
		log.Fatal("server stopped")
	}
}
func stamp(ms int64) string { return time.UnixMilli(ms).UTC().Format(time.RFC3339Nano) }
func (s *Server) bounds(day string) (int64, int64, error) {
	t, e := time.ParseInLocation("2006-01-02", day, s.zone)
	if e != nil {
		return 0, 0, errInput
	}
	return t.UnixMilli(), t.AddDate(0, 0, 1).UnixMilli(), nil
}
func (s *Server) open() (*sql.DB, Publication, string, error) {
	var p Publication
	b, e := os.ReadFile(filepath.Join(s.config.Root, "current.json"))
	if e != nil {
		return nil, p, "", e
	}
	if e = json.Unmarshal(b, &p); e != nil {
		return nil, p, "", e
	}
	if p.Generation == "" || filepath.Base(p.Generation) != p.Generation || strings.ContainsAny(p.Generation, "/\\:") || p.Generation == "." || p.Generation == ".." {
		return nil, p, "", errInput
	}
	dir := filepath.Join(s.config.Root, "snapshots", p.Generation)
	resolved, e := filepath.EvalSymlinks(dir)
	if e != nil {
		return nil, p, "", e
	}
	base, e := filepath.EvalSymlinks(filepath.Join(s.config.Root, "snapshots"))
	if e != nil {
		return nil, p, "", e
	}
	rel, e := filepath.Rel(base, resolved)
	if e != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return nil, p, "", errInput
	}
	file := filepath.Join(resolved, "snapshot.sqlite")
	real, e := filepath.EvalSymlinks(file)
	if e != nil || real != file {
		return nil, p, "", errInput
	}
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(real)}
	q := u.Query()
	q.Set("mode", "ro")
	q.Set("immutable", "1")
	q.Add("_pragma", "query_only(1)")
	u.RawQuery = q.Encode()
	db, e := sql.Open("sqlite", u.String())
	if e != nil {
		return nil, p, "", e
	}
	db.SetMaxOpenConns(1)
	if e = db.Ping(); e != nil {
		db.Close()
		return nil, p, "", e
	}
	return db, p, dir, nil
}
func send(w http.ResponseWriter, status int, v any) {
	b, e := json.Marshal(v)
	if e != nil {
		status = 503
		b = []byte(`{"code":"UNAVAILABLE","message":"Snapshot unavailable"}`)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	w.Write(b)
}
func fail(w http.ResponseWriter, status int, code, message string) {
	send(w, status, map[string]string{"code": code, "message": message})
}
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if r.Method != "GET" {
		w.Header().Set("Allow", "GET")
		fail(w, 405, "METHOD_NOT_ALLOWED", "Use GET")
		return
	}
	if !strings.HasPrefix(r.URL.Path, "/api/") {
		s.static(w, r)
		return
	}
	allowed := map[string]bool{"/api/health": true, "/api/today": true, "/api/focus": true, "/api/topics/evidence": true, "/api/agents": true, "/api/events": true, "/api/filters": true}
	if !allowed[r.URL.Path] {
		fail(w, 404, "NOT_FOUND", "Endpoint unavailable")
		return
	}
	db, p, dir, e := s.open()
	if e != nil {
		fail(w, 503, "UNAVAILABLE", "Snapshot unavailable")
		return
	}
	defer db.Close()
	var v any
	switch r.URL.Path {
	case "/api/health":
		v = map[string]any{"status": "ok", "snapshot_at": p.SnapshotAt, "timezone": "Asia/Shanghai"}
	case "/api/events":
		v, e = s.events(db, r.URL.Query(), false)
	case "/api/topics/evidence":
		v, e = s.events(db, r.URL.Query(), true)
	case "/api/filters":
		v, e = s.filters(db)
	case "/api/focus":
		v, e = s.focus(db, dir)
	case "/api/today":
		v, e = s.today(db, dir, r.URL.Query().Get("source"))
	case "/api/agents":
		v, e = s.agents(db)
	}
	if e != nil {
		if errors.Is(e, errInput) {
			fail(w, 400, "INVALID_QUERY", "Check filters and pagination")
		} else {
			fail(w, 503, "UNAVAILABLE", "Snapshot unavailable")
		}
		return
	}
	w.Header().Set("X-Snapshot-At", p.SnapshotAt)
	send(w, 200, v)
}
func (s *Server) static(w http.ResponseWriter, r *http.Request) {
	if s.config.SPA == "" {
		fail(w, 503, "SPA_UNAVAILABLE", "Application unavailable")
		return
	}
	clean := filepath.Clean(filepath.FromSlash(r.URL.Path))
	clean = strings.TrimLeft(clean, "/\\")
	path := filepath.Join(s.config.SPA, clean)
	rel, e := filepath.Rel(s.config.SPA, path)
	if e != nil || strings.HasPrefix(rel, "..") {
		fail(w, 404, "NOT_FOUND", "Page unavailable")
		return
	}
	info, e := os.Stat(path)
	if e != nil || info.IsDir() {
		path = filepath.Join(s.config.SPA, "index.html")
	}
	if _, e = os.Stat(path); e != nil {
		fail(w, 503, "SPA_UNAVAILABLE", "Application unavailable")
		return
	}
	// ServeContent avoids FileServer redirects for SPA entry points.
	f, e := os.Open(path)
	if e != nil {
		fail(w, 503, "SPA_UNAVAILABLE", "Application unavailable")
		return
	}
	defer f.Close()
	info, e = f.Stat()
	if e != nil {
		fail(w, 503, "SPA_UNAVAILABLE", "Application unavailable")
		return
	}
	http.ServeContent(w, r, info.Name(), info.ModTime(), f)
}
func eventRows(db *sql.DB, where string, args []any, limit, offset int) ([]Event, error) {
	rows, e := db.Query(`SELECT `+projection+` FROM attention_engine_events WHERE `+where+` ORDER BY ts DESC,id ASC LIMIT ? OFFSET ?`, append(append([]any{}, args...), limit, offset)...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Event{}
	for rows.Next() {
		var x Event
		var n sql.NullInt64
		var ms int64
		if e = rows.Scan(&x.ID, &x.Kind, &x.Topic, &n, &x.Scope, &ms); e != nil {
			return nil, e
		}
		if n.Valid {
			x.Tokens = &n.Int64
		}
		x.TS = stamp(ms)
		out = append(out, x)
	}
	return out, rows.Err()
}
func page(q url.Values) (int, int, error) {
	p, n := 1, 50
	var e error
	if q.Get("page") != "" {
		p, e = strconv.Atoi(q.Get("page"))
		if e != nil {
			return 0, 0, errInput
		}
	}
	if q.Get("page_size") != "" {
		n, e = strconv.Atoi(q.Get("page_size"))
		if e != nil {
			return 0, 0, errInput
		}
	}
	if p < 1 || p > 10000000 || n < 1 || n > 100 {
		return 0, 0, errInput
	}
	return p, n, nil
}
func (s *Server) events(db *sql.DB, q url.Values, evidence bool) (any, error) {
	p, n, e := page(q)
	if e != nil {
		return nil, e
	}
	where := "1=1"
	args := []any{}
	for _, k := range []string{"topic", "kind", "source"} {
		if v := q.Get(k); v != "" {
			where += " AND " + k + "=?"
			args = append(args, v)
		}
	}
	for _, k := range []string{"start", "end"} {
		if v := q.Get(k); v != "" {
			a, b, e := s.bounds(v)
			if e != nil {
				return nil, e
			}
			if k == "start" {
				where += " AND ts>=?"
				args = append(args, a)
			} else {
				where += " AND ts<?"
				args = append(args, b)
			}
		}
	}
	if q.Get("start") != "" && q.Get("end") != "" && q.Get("start") > q.Get("end") {
		return nil, errInput
	}
	var total int
	if e = db.QueryRow(`SELECT COUNT(*) FROM attention_engine_events WHERE `+where, args...).Scan(&total); e != nil {
		return nil, e
	}
	ev, e := eventRows(db, where, args, n, (p-1)*n)
	if e != nil {
		return nil, e
	}
	result := map[string]any{"events": ev, "total": total, "page": p, "page_size": n}
	if evidence {
		result["topic"] = q.Get("topic")
		rows, e := db.Query(`SELECT strftime('%Y-%m-%d',ts/1000.0,'unixepoch','+8 hours'),kind,COUNT(*) FROM attention_engine_events WHERE `+where+` GROUP BY 1,2 ORDER BY 1,2`, args...)
		if e != nil {
			return nil, e
		}
		defer rows.Close()
		ag := []map[string]any{}
		for rows.Next() {
			var day, k string
			var c int
			if e = rows.Scan(&day, &k, &c); e != nil {
				return nil, e
			}
			ag = append(ag, map[string]any{"day": day, "kind": k, "count": c})
		}
		if e = rows.Err(); e != nil {
			return nil, e
		}
		result["aggregates"] = ag
	}
	return result, nil
}
func (s *Server) filters(db *sql.DB) (any, error) {
	out := map[string]any{}
	for _, k := range []string{"topic", "kind", "source"} {
		rows, e := db.Query(`SELECT DISTINCT ` + k + ` FROM attention_engine_events UNION SELECT ` + k + ` FROM attention_engine_events WHERE 0 ORDER BY 1`)
		if e != nil {
			return nil, e
		}
		v := []string{}
		for rows.Next() {
			var x string
			if e = rows.Scan(&x); e != nil {
				rows.Close()
				return nil, e
			}
			v = append(v, x)
		}
		e = rows.Err()
		rows.Close()
		if e != nil {
			return nil, e
		}
		out[k+"s"] = v
	}
	return out, nil
}
func topics(db *sql.DB, source string) ([]Topic, error) {
	where := ""
	args := []any{}
	if source != "" {
		where = ` WHERE topic IN (SELECT DISTINCT topic FROM attention_engine_events WHERE source=?)`
		args = append(args, source)
	}
	rows, e := db.Query(`SELECT topic,attention,momentum,confidence FROM attention_engine_nodes`+where+` ORDER BY attention DESC,topic ASC`, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Topic{}
	for rows.Next() {
		var t Topic
		if e = rows.Scan(&t.Topic, &t.Attention, &t.Momentum, &t.Confidence); e != nil {
			return nil, e
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
func history(dir string) (map[string][]HistoryPoint, error) {
	b, e := os.ReadFile(filepath.Join(dir, "history.json"))
	if e != nil {
		return nil, e
	}
	out := map[string][]HistoryPoint{}
	e = json.Unmarshal(b, &out)
	return out, e
}
func (s *Server) focus(db *sql.DB, dir string) (any, error) {
	ts, e := topics(db, "")
	if e != nil {
		return nil, e
	}
	hist, e := history(dir)
	if e != nil {
		return nil, e
	}
	// Only persisted post-update A values qualify. Applied contribution/strength are not A.
	rows, e := db.Query(`SELECT topic,json_extract(j.value,'$.signal.ts'),json_extract(j.value,'$.attention') FROM attention_engine_nodes,json_each(data,'$.history.applied') j WHERE json_type(j.value,'$.attention') IN ('real','integer') AND json_type(j.value,'$.signal.ts')='integer' ORDER BY topic,2,j.key`)
	if e != nil {
		return nil, e
	}
	for rows.Next() {
		var topic string
		var p HistoryPoint
		if e = rows.Scan(&topic, &p.TS, &p.Attention); e != nil {
			rows.Close()
			return nil, e
		}
		hist[topic] = append(hist[topic], p)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	series := []Series{}
	for _, t := range ts {
		pts := hist[t.Topic]
		sort.SliceStable(pts, func(i, j int) bool { return pts[i].TS < pts[j].TS })
		ss := Series{t.Topic, []Point{}}
		for i, p := range pts {
			if i > 0 && pts[i-1].TS == p.TS {
				continue
			}
			ss.Points = append(ss.Points, Point{stamp(p.TS), p.Attention})
		}
		series = append(series, ss)
	}
	return map[string]any{"topics": ts, "series": series}, nil
}
func (s *Server) today(db *sql.DB, dir, source string) (any, error) {
	day := s.now().In(s.zone).Format("2006-01-02")
	start, end, _ := s.bounds(day)
	yday := s.now().In(s.zone).AddDate(0, 0, -1).Format("2006-01-02")
	ts, e := topics(db, source)
	if e != nil {
		return nil, e
	}
	hist, e := history(dir)
	if e != nil {
		return nil, e
	}
	cards := []Card{}
	// Delta is available only with confirmed observations on both calendar days.
	for _, t := range ts {
		var a, b *float64
		for _, p := range hist[t.Topic] {
			d := time.UnixMilli(p.TS).In(s.zone).Format("2006-01-02")
			v := p.Attention
			if d == day {
				a = &v
			}
			if d == yday {
				b = &v
			}
		}
		var delta *float64
		if a != nil && b != nil {
			v := *a - *b
			delta = &v
		}
		cards = append(cards, Card{t, delta})
	}
	where := "ts>=? AND ts<?"
	args := []any{start, end}
	if source != "" {
		where += " AND source=?"
		args = append(args, source)
	}
	activity, e := eventRows(db, where, args, 50, 0)
	if e != nil {
		return nil, e
	}
	sleep := map[string]any{"day": yday, "status": "not_produced", "completed_at": nil}
	var completed sql.NullInt64
	e = db.QueryRow(`SELECT completed_at FROM attention_engine_sleep_runs WHERE day=? ORDER BY started_at DESC,id LIMIT 1`, yday).Scan(&completed)
	if e != nil && e != sql.ErrNoRows {
		return nil, e
	}
	if e == nil {
		sleep["status"] = "running"
		if completed.Valid && completed.Int64 > 0 {
			sleep["status"] = "completed"
			sleep["completed_at"] = stamp(completed.Int64)
		}
	}
	counts := []map[string]any{}
	for i := 6; i >= 0; i-- {
		d := s.now().In(s.zone).AddDate(0, 0, -i).Format("2006-01-02")
		a, b, _ := s.bounds(d)
		args := []any{a, b}
		w := "ts>=? AND ts<?"
		if source != "" {
			w += " AND source=?"
			args = append(args, source)
		}
		var c int
		if e = db.QueryRow(`SELECT COUNT(*) FROM attention_engine_events WHERE `+w, args...).Scan(&c); e != nil {
			return nil, e
		}
		counts = append(counts, map[string]any{"day": d, "count": c})
	}
	rows, e := db.Query(`SELECT source,COUNT(*) FROM attention_engine_events WHERE ts>=? AND ts<? GROUP BY source ORDER BY source`, start, end)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	sources := []map[string]any{}
	for rows.Next() {
		var x string
		var c int
		if e = rows.Scan(&x, &c); e != nil {
			return nil, e
		}
		sources = append(sources, map[string]any{"source": x, "count": c})
	}
	if e = rows.Err(); e != nil {
		return nil, e
	}
	return map[string]any{"day": day, "focus": cards, "activity": activity, "sleep": sleep, "daily_counts": counts, "sources": sources}, nil
}
func (s *Server) agents(db *sql.DB) (any, error) {
	d := s.now().In(s.zone)
	a, b, _ := s.bounds(d.Format("2006-01-02"))
	a7, _, _ := s.bounds(d.AddDate(0, 0, -6).Format("2006-01-02"))
	a30, _, _ := s.bounds(d.AddDate(0, 0, -29).Format("2006-01-02"))
	rows, e := db.Query(`SELECT source,SUM(CASE WHEN ts>=? AND ts<? THEN 1 ELSE 0 END),SUM(CASE WHEN ts>=? AND ts<? THEN 1 ELSE 0 END),SUM(CASE WHEN ts>=? AND ts<? THEN 1 ELSE 0 END) FROM attention_engine_events GROUP BY source ORDER BY source`, a, b, a7, b, a30, b)
	if e != nil {
		return nil, e
	}
	groups := []map[string]any{}
	for rows.Next() {
		var x string
		var n, n7, n30 int
		if e = rows.Scan(&x, &n, &n7, &n30); e != nil {
			rows.Close()
			return nil, e
		}
		groups = append(groups, map[string]any{"source": x, "today": n, "days7": n7, "days30": n30})
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	for _, g := range groups {
		ev, e := eventRows(db, "source=?", []any{g["source"]}, 5, 0)
		if e != nil {
			return nil, e
		}
		g["recent"] = ev
	}
	return map[string]any{"groups": groups}, nil
}
