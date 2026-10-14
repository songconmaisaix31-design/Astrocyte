package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/songconmaisaix31-design/Astrocyte/cmd/internal/attentionruntime"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	adapter "github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "PUBLIC_SOURCE_VERIFICATION_FAILED")
		os.Exit(1)
	}
}
func write(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), "safe-*.json")
	if err != nil {
		return err
	}
	if _, err = f.Write(b); err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
func git(ctx context.Context, repo string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", repo}, args...)...)
	b, err := cmd.Output()
	return strings.TrimSpace(string(b)), err
}
func run() (finalErr error) {
	f := flag.NewFlagSet("actual-public-cognition", flag.ContinueOnError)
	mode := f.String("mode", "", "collect or inspect")
	source := f.String("source", "", "explicit server-initialized independent SQLite")
	repo := f.String("repository", "", "explicit public bare Git checkout")
	ref := f.String("ref", "", "explicit actual reachable public Git ref")
	scope := f.String("scope", "", "actual source project scope")
	root := f.String("root", "", "published snapshot root")
	out := f.String("output", "", "owned verification directory")
	strength := f.Float64("strength", -1, "explicit approved signal strength")
	decay := f.Float64("decay", -1, "explicit approved decay")
	confidence := f.Float64("initial-confidence", -1, "explicit approved initial confidence")
	seed := f.Bool("authorized-public-replay", false, "coordinator-authorized manual entry at actual earliest source time")
	pace := f.Duration("pace", 0, "explicit local ingestion pacing; source event times remain unchanged")
	if err := f.Parse(os.Args[1:]); err != nil {
		return err
	}
	defer func() {
		if finalErr != nil && *out != "" {
			_ = os.WriteFile(filepath.Join(*out, "verification-error.txt"), []byte(finalErr.Error()), 0600)
		}
	}()
	if *source == "" || *out == "" || f.NArg() != 0 {
		return errors.New("explicit paths required")
	}
	if err := os.MkdirAll(*out, 0700); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	if *mode == "inspect" {
		return inspect(ctx, *source, *root, *out)
	}
	if *mode != "collect" || *repo == "" || *ref == "" || *scope == "" || !*seed || *strength < 0 || *decay < 0 || *confidence < 0 || *confidence > 1 {
		return errors.New("explicit approved collection configuration required")
	}
	if _, err := os.Stat(*source); err != nil {
		return err
	}
	remote, err := git(ctx, *repo, "remote", "get-url", "origin")
	if err != nil {
		return err
	}
	if remote != "https://github.com/songconmaisaix31-design/Astrocyte" {
		return errors.New("only approved public repository allowed")
	}
	ids, err := git(ctx, *repo, "rev-list", *ref)
	if err != nil {
		return err
	}
	var records []app.StoredEvent
	for _, id := range strings.Fields(ids) {
		stamp, e := git(ctx, *repo, "show", "-s", "--format=%ct", id)
		if e != nil {
			return e
		}
		seconds, e := strconv.ParseInt(stamp, 10, 64)
		if e != nil {
			return e
		}
		subject, e := git(ctx, *repo, "show", "-s", "--format=%s", id)
		if e != nil {
			return e
		}
		raw, e := json.Marshal(struct {
			events.GitLogRecord
			PublicRepository string `json:"public_repository"`
		}{events.GitLogRecord{Commit: id, TS: seconds * 1000, Subject: subject}, remote})
		if e != nil {
			return e
		}
		row, e := events.MapGitLogRecord(*scope, raw, time.Now().UnixMilli())
		if e != nil {
			return e
		}
		records = append(records, row)
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].Observation.TS == records[j].Observation.TS {
			return records[i].OriginKey < records[j].OriginKey
		}
		return records[i].Observation.TS < records[j].Observation.TS
	})
	db, err := adapter.Open(*source)
	if err != nil {
		return err
	}
	defer db.Close()
	var initial int
	if err = db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM attention_engine_events").Scan(&initial); err != nil {
		return err
	}
	if initial != 0 {
		return errors.New("collection requires real initialized empty source")
	}
	// 只在本次自有库关闭自动 checkpoint，便于实际并发 WAL 备份核查。
	if _, err = db.Conn().ExecContext(ctx, "PRAGMA wal_autocheckpoint=0"); err != nil {
		return err
	}
	engine := app.NewDynamicsEngine(adapter.NewDynamicsRepository(db), app.DynamicsOptions{Decay: decay, InitialConfidence: confidence})
	repository := adapter.NewEventRepository(db)
	recorder := attentionruntime.EventRecorder{Events: app.NewEventService(repository, *strength), Repository: repository, Engine: engine}
	seeded := map[string]int64{}
	for _, row := range records {
		if row.Event.Topic == "" {
			continue
		}
		if _, ok := seeded[row.Event.Topic]; ok {
			continue
		}
		// 本次获准人工入口使用实际公开来源时间，保留历史重放说明。
		if _, err = engine.SeedTopic(ctx, app.Principal{ID: "authorized-public-replay", Kind: "human"}, row.Event.Topic, row.Event.TS); err != nil {
			return err
		}
		seeded[row.Event.Topic] = row.Event.TS
	}
	if err = write(filepath.Join(*out, "writer-started.json"), map[string]any{"pid": os.Getpid(), "actual_commits": len(records), "seeded_source_times": seeded, "source": remote, "ref": *ref, "scope": *scope}); err != nil {
		return err
	}
	for i, row := range records {
		if _, err = recorder.Record(ctx, []app.StoredEvent{row}); err != nil {
			return err
		}
		if err = write(filepath.Join(*out, "progress.json"), map[string]any{"recorded": i + 1, "actual_commits": len(records), "pid": os.Getpid()}); err != nil {
			return err
		}
		if *pace > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(*pace):
			}
		}
	}
	return write(filepath.Join(*out, "writer-completed.json"), map[string]any{"recorded": len(records), "source": remote, "ref": *ref, "scope": *scope, "seeded_source_times": seeded})
}

func readonly(path string) (*sql.DB, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	p := filepath.ToSlash(absolute)
	if filepath.VolumeName(absolute) != "" {
		p = "/" + p
	}
	u := url.URL{Scheme: "file", Path: p, RawQuery: "mode=ro&_pragma=query_only(1)"}
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return db, nil
}
func inspect(ctx context.Context, source, root, out string) error {
	db, err := readonly(source)
	if err != nil {
		return err
	}
	defer db.Close()
	counts := map[string]int{}
	for _, table := range []string{"attention_engine_events", "attention_engine_signals", "attention_engine_nodes", "attention_engine_sleep_runs"} {
		var n int
		if err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&n); err != nil {
			return err
		}
		counts[table] = n
	}
	var origins, signals int
	if err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM attention_engine_events WHERE origin_key LIKE 'git:%'`).Scan(&origins); err != nil {
		return err
	}
	if err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM attention_engine_signals s JOIN attention_engine_events e ON e.id=s.event_id WHERE e.origin_key LIKE 'git:%'`).Scan(&signals); err != nil {
		return err
	}
	if err = write(filepath.Join(out, "actual-counts.json"), map[string]any{"counts": counts, "actual_git_origins": origins, "signals_linked_to_git": signals}); err != nil {
		return err
	}
	if root == "" {
		return nil
	}
	reader := adapter.NewWorkbenchReader(root)
	views := map[string]workbench.View{}
	for _, name := range []string{"health", "today", "focus", "agents", "events", "filters", "topics/evidence"} {
		v, e := reader.Read(ctx, name, workbench.Query{})
		if e != nil {
			return e
		}
		views[name] = v
	}
	if err = write(filepath.Join(out, "production-reader-safe.json"), views); err != nil {
		return err
	}
	// 独立 SQL 只读取六字段及所需聚合，不输出 observation/raw/data。
	rows, err := db.QueryContext(ctx, `SELECT id,kind,topic,CASE WHEN json_type(data,'$.tokens')='integer' THEN json_extract(data,'$.tokens') END,CASE WHEN json_type(data,'$.scope')='text' THEN json_extract(data,'$.scope') ELSE '' END,ts FROM attention_engine_events ORDER BY ts DESC,id ASC`)
	if err != nil {
		return err
	}
	safe := []workbench.Event{}
	for rows.Next() {
		var x workbench.Event
		var tokens sql.NullInt64
		var ts int64
		if err = rows.Scan(&x.ID, &x.Kind, &x.Topic, &tokens, &x.Scope, &ts); err != nil {
			rows.Close()
			return err
		}
		if tokens.Valid {
			v := tokens.Int64
			x.Tokens = &v
		}
		x.TS = time.UnixMilli(ts).UTC().Format(time.RFC3339Nano)
		safe = append(safe, x)
	}
	err = errors.Join(rows.Err(), rows.Close())
	if err != nil {
		return err
	}
	if err = write(filepath.Join(out, "events-six-fields.json"), safe); err != nil {
		return err
	}
	for page := 1; page <= (len(safe)+99)/100; page++ {
		v, e := reader.Read(ctx, "events", workbench.Query{Page: page, PageSize: 100})
		if e != nil {
			return e
		}
		for i, x := range v.Events.Events {
			expected := safe[(page-1)*100+i]
			a, _ := json.Marshal(x)
			b, _ := json.Marshal(expected)
			if string(a) != string(b) {
				return errors.New("Reader six fields differ from source SQL")
			}
		}
	}
	if len(views["focus"].Focus.Topics) < 2 {
		return errors.New("fewer than two actual topics")
	}
	topic := views["focus"].Focus.Topics[0].Topic
	var total int
	if err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM attention_engine_events WHERE topic=?", topic).Scan(&total); err != nil {
		return err
	}
	if total <= 20 {
		return errors.New("highest attention topic has insufficient actual events")
	}
	values := []map[string]any{}
	zone, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return err
	}
	today := time.Now().In(zone)
	day := today.Format(time.DateOnly)
	for i := 2; i >= 0; i-- {
		d := today.AddDate(0, 0, -i)
		start := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, zone)
		var n int
		if err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM attention_engine_events WHERE ts>=? AND ts<?", start.UnixMilli(), start.AddDate(0, 0, 1).UnixMilli()).Scan(&n); err != nil {
			return err
		}
		values = append(values, map[string]any{"route": "/cognition/today", "metric": "daily_count", "day": start.Format(time.DateOnly), "value": n, "query": map[string]any{}})
	}
	for _, metric := range []string{"attention", "momentum", "confidence"} {
		var n float64
		if err = db.QueryRowContext(ctx, "SELECT "+metric+" FROM attention_engine_nodes WHERE topic=?", topic).Scan(&n); err != nil {
			return err
		}
		values = append(values, map[string]any{"route": "/cognition/focus", "metric": metric, "topic": topic, "value": n, "query": map[string]any{}})
	}
	var src string
	if err = db.QueryRowContext(ctx, "SELECT source FROM attention_engine_events GROUP BY source ORDER BY COUNT(*) DESC,source LIMIT 1").Scan(&src); err != nil {
		return err
	}
	for _, item := range []struct {
		metric string
		days   int
	}{{"today", 1}, {"days7", 7}, {"days30", 30}} {
		start := today.AddDate(0, 0, -item.days+1)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, zone)
		end := time.Date(today.Year(), today.Month(), today.Day()+1, 0, 0, 0, 0, zone)
		var n int
		if err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM attention_engine_events WHERE source=? AND ts>=? AND ts<?", src, start.UnixMilli(), end.UnixMilli()).Scan(&n); err != nil {
			return err
		}
		values = append(values, map[string]any{"route": "/cognition/agents", "metric": item.metric, "source": src, "value": n, "query": map[string]any{}})
	}
	var all int
	if err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM attention_engine_events").Scan(&all); err != nil {
		return err
	}
	values = append(values, map[string]any{"route": "/cognition/evidence", "metric": "total", "value": all, "query": map[string]any{}})
	return write(filepath.Join(out, "sql-expected.json"), map[string]any{"snapshot_at": views["health"].SnapshotAt, "day": day, "values": values})
}
