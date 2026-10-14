package main

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
	_ "modernc.org/sqlite"
)

type options struct {
	base, mode, output, verification, zip, executable, source, method, path, body, db, task, project string
	expect, minimum                                                                                  int
}

type client struct {
	options
	http     *http.Client
	csrf     string
	sequence int
	summary  map[string]any
}

func failure(code string) error { return errors.New(code) }

func main() {
	var o options
	flag.StringVar(&o.base, "base", "", "actual loopback server URL")
	flag.StringVar(&o.mode, "mode", "", "baseline, registry, request, events, or idle")
	flag.StringVar(&o.output, "output", "", "new local evidence directory")
	flag.StringVar(&o.verification, "verification-root", "", "dedicated .verification directory")
	flag.StringVar(&o.zip, "zip", "", "actual P ZIP")
	flag.StringVar(&o.executable, "executable", "", "actual running extracted executable")
	flag.StringVar(&o.source, "package-source", "", "P declared exact package SOURCE")
	flag.StringVar(&o.method, "method", "GET", "GET, POST, or PUT")
	flag.StringVar(&o.path, "path", "", "exact existing API path")
	flag.StringVar(&o.body, "body-file", "", "local JSON request file")
	flag.StringVar(&o.db, "db", "", "actual mutable application SQLite")
	flag.StringVar(&o.task, "task", "", "exact owned task UUID")
	flag.StringVar(&o.project, "project", "", "exact owned project ID")
	flag.IntVar(&o.expect, "expect-status", 200, "required response status")
	flag.IntVar(&o.minimum, "minimum-events", 1, "required event count")
	flag.Parse()
	c := &client{options: o, summary: map[string]any{"result": "FAIL", "mode": o.mode, "native_completion": "NOT_RUN", "DOM": "NOT_RUN", "original_session_comparison": "NOT_RUN"}}
	err := c.run()
	if err != nil {
		c.summary["error_code"] = err.Error()
	} else {
		c.summary["result"] = "PASS"
	}
	encoded, encodeErr := json.MarshalIndent(c.summary, "", "  ")
	if encodeErr != nil {
		fmt.Println("FAIL: summary_encoding")
		os.Exit(1)
	}
	if c.output != "" {
		if err := exclusive(filepath.Join(c.output, "summary.json"), encoded); err != nil {
			fmt.Println("FAIL: summary_write")
			os.Exit(1)
		}
	}
	fmt.Println(string(encoded))
	if err != nil {
		os.Exit(1)
	}
}

func inside(root, path string) bool {
	r, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	p, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(strings.ToLower(r), strings.ToLower(p))
	return err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

func exclusive(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return failure("evidence_create")
	}
	_, writeErr := f.Write(data)
	closeErr := f.Close()
	if writeErr != nil || closeErr != nil {
		return failure("evidence_write")
	}
	return nil
}

func (c *client) run() error {
	if filepath.Base(c.verification) != ".verification" || !inside(c.verification, c.output) {
		c.output = ""
		return failure("local_evidence_boundary")
	}
	if err := os.Mkdir(c.output, 0700); err != nil {
		c.output = ""
		return failure("output_must_be_new")
	}
	if len(c.source) != 40 {
		return failure("package_source_required")
	}
	if err := c.packageBytes(); err != nil {
		return err
	}
	u, err := url.Parse(c.base)
	if err != nil || u.Scheme != "http" || u.Hostname() != "127.0.0.1" || u.Port() == "" || u.Path != "" || u.RawQuery != "" || u.User != nil || u.Fragment != "" {
		return failure("loopback_url_required")
	}
	if u.Port() == "17321" || u.Port() == "17342" || u.Port() == "17343" {
		return failure("reserved_port")
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return failure("cookie_jar")
	}
	c.http = &http.Client{Jar: jar, Timeout: 20 * time.Second, Transport: &http.Transport{Proxy: nil, DisableKeepAlives: true}, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	defer c.http.CloseIdleConnections()
	if _, err := c.request("GET", "/api/v1/health", nil, 200); err != nil {
		return err
	}
	data, err := c.request("GET", "/api/v1/auth/session", nil, 200)
	if err != nil {
		return err
	}
	var identity struct {
		Kind string `json:"actor_kind"`
		CSRF string `json:"csrf_token"`
	}
	if json.Unmarshal(data, &identity) != nil || identity.Kind != "human" || identity.CSRF == "" || len(jar.Cookies(u)) == 0 {
		return failure("human_bootstrap")
	}
	c.csrf = identity.CSRF
	c.summary["bootstrap"] = "PASS"
	switch c.mode {
	case "baseline":
		return c.baseline(true)
	case "registry":
		return c.baseline(false)
	case "idle":
		return c.idle()
	case "request":
		if !allowed(c.method, c.path) {
			return failure("request_out_of_scope")
		}
		var body []byte
		if c.method != "GET" {
			if !inside(c.verification, c.body) {
				return failure("request_body_boundary")
			}
			body, err = os.ReadFile(c.body)
			if err != nil || !json.Valid(body) {
				return failure("request_body_json")
			}
		}
		data, err = c.request(c.method, c.path, body, c.expect)
		if err != nil {
			return err
		}
		var document map[string]json.RawMessage
		if json.Unmarshal(data, &document) != nil {
			return failure("response_json")
		}
		// 只导出既有资源标识符；完整响应保留在本机目录。
		for _, key := range []string{"project", "work", "view"} {
			var item map[string]json.RawMessage
			if value, ok := document[key]; ok && json.Unmarshal(value, &item) == nil {
				if key == "view" {
					var task map[string]json.RawMessage
					if json.Unmarshal(item["task"], &task) == nil {
						item = task
					}
				}
				for _, field := range []string{"project_id", "work_id", "task_id", "revision", "state"} {
					if value, ok := item[field]; ok {
						var v any
						if json.Unmarshal(value, &v) == nil {
							c.summary[field] = v
						}
					}
				}
			}
		}
		return nil
	case "events":
		return c.events()
	default:
		return failure("mode_required")
	}
}

func (c *client) packageBytes() error {
	if !inside(c.verification, c.executable) {
		return failure("executable_boundary")
	}
	archive, err := zip.OpenReader(c.zip)
	if err != nil {
		return failure("zip_open")
	}
	defer archive.Close()
	actual, err := os.ReadFile(c.executable)
	if err != nil {
		return failure("executable_read")
	}
	matched := 0
	for _, entry := range archive.File {
		if entry.Name != "Astrocyte-v0.1-windows-x64/astrocyte.exe" {
			continue
		}
		matched++
		reader, err := entry.Open()
		if err != nil {
			return failure("zip_executable_open")
		}
		packaged, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil || closeErr != nil || !bytes.Equal(actual, packaged) {
			return failure("zip_executable_mismatch")
		}
	}
	if matched != 1 {
		return failure("zip_executable_count")
	}
	c.summary["package_source"] = c.source
	c.summary["zip"] = c.zip
	c.summary["executable"] = c.executable
	c.summary["zip_executable_bytes"] = "PASS"
	return nil
}

func allowed(method, path string) bool {
	u, err := url.Parse(path)
	if err != nil || u.IsAbs() || u.Host != "" || u.Fragment != "" || strings.Contains(path, "..") {
		return false
	}
	if method == "GET" {
		return strings.HasPrefix(path, "/api/v1/relay/") || path == "/api/v1/local-projects" || path == "/api/v1/works" || path == "/api/v1/local-projects/registered"
	}
	if u.RawQuery != "" {
		return false
	}
	if method == "POST" {
		return path == "/api/v1/local-projects" || path == "/api/v1/works/refresh" || path == "/api/v1/works/delivery" || path == "/api/v1/relay/intents" || (strings.HasPrefix(path, "/api/v1/relay/intents/") && strings.HasSuffix(path, "/confirm")) || (strings.HasPrefix(path, "/api/v1/relay/tasks/") && strings.HasSuffix(path, "/stop"))
	}
	return method == "PUT" && strings.HasPrefix(path, "/api/v1/local-projects/") && strings.HasSuffix(path, "/settings")
}

func (c *client) request(method, path string, body []byte, expected int) ([]byte, error) {
	c.sequence++
	req, err := http.NewRequest(method, c.base+path, bytes.NewReader(body))
	if err != nil {
		return nil, failure("request_construct")
	}
	if method != "GET" {
		req.Header.Set("Origin", c.base)
		req.Header.Set("X-CSRF-Token", c.csrf)
		req.Header.Set("Content-Type", "application/json")
		operation := uuid.NewString()
		var envelope map[string]json.RawMessage
		if json.Unmarshal(body, &envelope) == nil {
			if id, ok := envelope["request_id"]; ok {
				var value string
				if json.Unmarshal(id, &value) != nil {
					return nil, failure("request_id_invalid")
				}
				parsed, e := uuid.Parse(value)
				if e != nil || parsed == uuid.Nil {
					return nil, failure("request_id_invalid")
				}
				operation = value
			}
		}
		req.Header.Set("Idempotency-Key", operation)
		// 在发送前封存写入信息，传输失败后禁止再次发送。
		metadata, _ := json.Marshal(map[string]any{"method": method, "path": path, "operation_id": operation, "status": "NOT_SENT"})
		if err := exclusive(filepath.Join(c.output, fmt.Sprintf("%03d-write.json", c.sequence)), metadata); err != nil {
			return nil, err
		}
	}
	res, err := c.http.Do(req)
	if err != nil {
		if method != "GET" {
			c.summary["delivery"] = "UNKNOWN_DO_NOT_RESEND"
		}
		return nil, failure("http_transport")
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, (8<<20)+1))
	if err != nil || len(data) > 8<<20 {
		return nil, failure("response_read")
	}
	if err := exclusive(filepath.Join(c.output, fmt.Sprintf("%03d-response-local.json", c.sequence)), data); err != nil {
		return nil, err
	}
	c.summary["last_http_status"] = res.StatusCode
	if res.StatusCode != expected {
		if method != "GET" {
			c.summary["delivery"] = "UNKNOWN_INSPECT_RECEIPT"
		}
		return nil, failure("http_status")
	}
	return data, nil
}

func (c *client) baseline(empty bool) error {
	data, err := c.request("GET", "/api/v1/relay/agents", nil, 200)
	if err != nil {
		return err
	}
	var registry struct {
		Version int               `json:"schema_version"`
		Agents  []json.RawMessage `json:"agents"`
	}
	if json.Unmarshal(data, &registry) != nil || registry.Version != 1 || registry.Agents == nil {
		return failure("agent_registry_schema")
	}
	if empty && len(registry.Agents) != 0 {
		return failure("baseline_agents_not_empty")
	}
	if !empty && len(registry.Agents) != 1 {
		return failure("expected_one_registered_agent")
	}
	c.summary["registered_agents"] = len(registry.Agents)
	return c.idle()
}

func (c *client) idle() error {
	data, err := c.request("GET", "/api/v1/relay/tasks", nil, 200)
	if err != nil {
		return err
	}
	var state struct {
		Version    int               `json:"schema_version"`
		Tasks      []json.RawMessage `json:"tasks"`
		Dispatched []string          `json:"dispatched_work_ids"`
	}
	if json.Unmarshal(data, &state) != nil || state.Version != 1 || state.Tasks == nil || state.Dispatched == nil {
		return failure("task_list_schema")
	}
	c.summary["tasks"] = len(state.Tasks)
	c.summary["dispatched_work_ids"] = len(state.Dispatched)
	if len(state.Tasks) != 0 || len(state.Dispatched) != 0 {
		return failure("unexpected_automatic_dispatch")
	}
	return nil
}

func decodeEvent(raw []byte) (workbench.Event, error) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || len(fields) != 6 {
		return workbench.Event{}, failure("six_fields_required")
	}
	for _, key := range []string{"event_id", "kind", "topic", "tokens", "scope", "ts"} {
		if _, ok := fields[key]; !ok {
			return workbench.Event{}, failure("six_fields_required")
		}
	}
	var event workbench.Event
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&event) != nil || decoder.Decode(new(any)) != io.EOF || event.ID == "" || event.Kind == "" || (event.Tokens != nil && *event.Tokens < 0) {
		return event, failure("event_projection")
	}
	if _, err := time.Parse(time.RFC3339Nano, event.TS); err != nil {
		return event, failure("event_timestamp")
	}
	return event, nil
}

func (c *client) events() error {
	id, err := uuid.Parse(c.task)
	if err != nil || id == uuid.Nil || id.String() != c.task || c.project == "" || c.minimum < 1 {
		return failure("owned_event_scope_required")
	}
	if !inside(c.verification, c.db) {
		return failure("database_boundary")
	}
	if _, err := os.Stat(c.db); err != nil {
		return failure("database_missing")
	}
	dbURL := url.URL{Scheme: "file", Path: filepath.ToSlash(c.db)}
	query := url.Values{"mode": {"ro"}, "_pragma": {"query_only(1)", "busy_timeout(5000)"}}
	dbURL.RawQuery = query.Encode()
	db, err := sql.Open("sqlite", dbURL.String())
	if err != nil {
		return failure("database_open")
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var readonly int
	if db.QueryRowContext(ctx, "PRAGMA query_only").Scan(&readonly) != nil || readonly != 1 {
		return failure("database_query_only")
	}
	seen := map[string]bool{}
	total := -1
	count := 0
	originCounts := map[string]int{}
	var coverage []workbench.SessionCoverage
	for page := 1; page <= 1000; page++ {
		path := fmt.Sprintf("/api/v1/relay/session-events?task_id=%s&page=%d&page_size=100", c.task, page)
		data, err := c.request("GET", path, nil, 200)
		if err != nil {
			return err
		}
		var result struct {
			Events   []json.RawMessage           `json:"events"`
			Coverage []workbench.SessionCoverage `json:"coverage"`
			Total    int                         `json:"total"`
			Page     int                         `json:"page"`
			PageSize int                         `json:"page_size"`
		}
		if json.Unmarshal(data, &result) != nil || result.Events == nil || result.Page != page || result.PageSize != 100 || result.Total < 0 {
			return failure("event_page_schema")
		}
		if total == -1 {
			total = result.Total
			coverage = result.Coverage
		} else if total != result.Total {
			return failure("mutable_event_population_changed")
		}
		for _, raw := range result.Events {
			event, err := decodeEvent(raw)
			if err != nil {
				return err
			}
			if seen[event.ID] || event.Topic != "relay:"+c.task || event.Scope != c.project {
				return failure("event_scope_or_duplicate")
			}
			seen[event.ID] = true
			var storedID, origin, observation, status string
			var ts int64
			var stored []byte
			err = db.QueryRowContext(ctx, "SELECT id,origin_key,ts,raw_bytes,observation,mapping_status FROM attention_engine_events WHERE id=?", event.ID).Scan(&storedID, &origin, &ts, &stored, &observation, &status)
			if err != nil {
				return failure("event_sql_locator")
			}
			projection, err := decodeEvent(stored)
			if err != nil {
				return err
			}
			projection.ID = storedID
			timestamp, _ := time.Parse(time.RFC3339Nano, projection.TS)
			projection.TS = timestamp.UTC().Format(time.RFC3339Nano)
			if !reflect.DeepEqual(event, projection) || timestamp.UnixMilli() != ts || status != "pending" {
				return failure("http_sql_six_field_mismatch")
			}
			var metadata struct {
				Session string `json:"session"`
				Scope   string `json:"scope"`
			}
			if json.Unmarshal([]byte(observation), &metadata) != nil || metadata.Session == "" || metadata.Scope != c.project {
				return failure("sql_observation_scope")
			}
			originType := ""
			for _, prefix := range []string{"relay:", "relay-session:", "relay-native-session:"} {
				if strings.HasPrefix(origin, prefix) {
					originType = prefix
					break
				}
			}
			if originType == "" {
				return failure("sql_origin_scope")
			}
			originCounts[originType]++
			count++
		}
		if count == total {
			break
		}
		if len(result.Events) == 0 || count > total || page == 1000 {
			return failure("event_pagination")
		}
	}
	if count < c.minimum {
		return failure("required_events_missing")
	}
	for _, item := range coverage {
		if item.TaskID != c.task || item.ProjectID != c.project {
			return failure("coverage_scope")
		}
	}
	c.summary["http_event_count"] = count
	c.summary["sql_located_count"] = count
	c.summary["six_field_match_count"] = count
	c.summary["origin_counts"] = originCounts
	c.summary["coverage"] = coverage
	return nil
}
