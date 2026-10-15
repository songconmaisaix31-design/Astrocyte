package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/repositorycontext"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

var evidence = flag.String("evidence", "", "existing ignored real S4 database and projections")
var sdk = flag.String("sdk", "", "installed official SDK")
var script = flag.String("script", "", "production CodeGraph script")
var aoci = flag.String("aoci", "", "verified AOCI executable")
var node = flag.String("node", "node", "trusted Node executable")

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func require(ok bool, message string) {
	if !ok {
		panic(errors.New(message))
	}
}
func pass(step string) { fmt.Println("PASS", step) }

type endpoint struct {
	base   string
	client *http.Client
	csrf   string
	server *http.Server
}

func start(db *sqlite.DB, options repositorycontext.Options) *endpoint {
	service, err := repositorycontext.New(db, db, options)
	must(err)
	ports := httpapi.RepositoryContextServices{Map: service, Search: service, Impact: service, Indexes: service, MapCommands: service}
	return serve(db, &ports)
}

func serve(db *sqlite.DB, ports *httpapi.RepositoryContextServices) *endpoint {
	projects := app.NewLocalProjectService(db, nil, agents.ProjectFiles{}, nil)
	production := httpapi.NewServer(httpapi.Config{Services: httpapi.Services{LocalProjects: projects, RepositoryContext: ports}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	must(err)
	server := &http.Server{Handler: production.Handler(), ReadHeaderTimeout: 10 * time.Second}
	go func() {
		err := server.Serve(listener)
		if !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()
	jar, err := cookiejar.New(nil)
	must(err)
	return &endpoint{base: "http://" + listener.Addr().String(), client: &http.Client{Jar: jar, Timeout: time.Minute}, server: server}
}

func (e *endpoint) close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	must(e.server.Shutdown(ctx))
	e.client.CloseIdleConnections()
}

func (e *endpoint) request(method, path string, body any, token string, withCSRF bool) (int, []byte) {
	diagnostic := time.AfterFunc(10*time.Second, func() {
		stack := make([]byte, 2*1024*1024)
		n := runtime.Stack(stack, true)
		name := "http-wait-" + time.Now().UTC().Format("20060102T150405.000000000") + ".txt"
		if err := os.WriteFile(filepath.Join(*evidence, name), stack[:n], 0600); err != nil {
			panic("HTTP diagnostic could not be preserved")
		}
	})
	defer diagnostic.Stop()
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		must(err)
	}
	req, err := http.NewRequest(method, e.base+path, bytes.NewReader(payload))
	must(err)
	req.Header.Set("Origin", e.base)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if withCSRF {
		req.Header.Set("X-CSRF-Token", e.csrf)
	}
	client := e.client
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
		client = &http.Client{Timeout: time.Minute}
	}
	response, err := client.Do(req)
	must(err)
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, 16*1024*1024))
	must(err)
	if token != "" {
		client.CloseIdleConnections()
	}
	return response.StatusCode, data
}

func expected(status int, data []byte, want int, code string) {
	require(status == want, fmt.Sprintf("HTTP status: expected %d, received %d", want, status))
	if code != "" {
		var result struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		must(json.Unmarshal(data, &result))
		require(result.Error.Code == code, "HTTP error code differs")
	}
}

func (e *endpoint) bootstrap() {
	status, data := e.request("GET", "/api/v1/auth/session", nil, "", false)
	expected(status, data, 200, "")
	var body struct {
		CSRF string `json:"csrf_token"`
	}
	must(json.Unmarshal(data, &body))
	require(body.CSRF != "", "real session did not issue CSRF")
	e.csrf = body.CSRF
}

func sourceResult(data []byte, roots ...string) domain.RepositoryResult {
	var envelope struct {
		Result domain.RepositoryResult `json:"repository_context"`
	}
	must(json.Unmarshal(data, &envelope))
	var fields any
	must(json.Unmarshal(data, &fields))
	var inspect func(any)
	inspect = func(value any) {
		switch item := value.(type) {
		case map[string]any:
			for key, v := range item {
				for _, forbidden := range []string{"root", "raw_text", "raw_vector", "vector", "overview_text", "runtime_repository_root", "native_id", "provider_payload", "section"} {
					require(key != forbidden, "private field escaped repository HTTP projection")
				}
				inspect(v)
			}
		case []any:
			for _, v := range item {
				inspect(v)
			}
		case string:
			for _, root := range roots {
				require(!strings.Contains(strings.ToLower(filepath.ToSlash(item)), strings.ToLower(filepath.ToSlash(root))), "absolute source root escaped repository HTTP projection")
			}
		}
	}
	inspect(fields)
	return envelope.Result
}

func main() {
	flag.Parse()
	defer func() {
		if failure := recover(); failure != nil {
			if *evidence != "" {
				_ = os.WriteFile(filepath.Join(*evidence, "http-failure-"+time.Now().UTC().Format("20060102T150405.000000000")+".txt"), []byte(fmt.Sprint(failure)), 0600)
			}
			fmt.Fprintln(os.Stderr, "FAIL actual_repository_http; inspect ignored evidence")
			os.Exit(1)
		}
	}()
	require(*evidence != "" && *sdk != "" && *script != "" && *aoci != "", "explicit real S4 configuration is required")
	ctx := context.Background()
	db, err := sqlite.Open(filepath.Join(*evidence, "state.sqlite"))
	must(err)
	must(db.RunMigrations(ctx, migrations.FS))
	p, err := db.LoadProject(ctx, "s4-public-a")
	must(err)
	pb, err := db.LoadProject(ctx, "s4-public-b")
	must(err)
	for _, root := range []string{p.Root, pb.Root} {
		remote, err := exec.Command("git", "-C", root, "remote", "get-url", "origin").Output()
		must(err)
		require(strings.TrimSpace(string(remote)) == "https://github.com/songconmaisaix31-design/astrocyte-s4-public-context-20261015.git", "HTTP verification is restricted to the owned public project")
	}
	budget := domain.RepositoryBudget{MaxFiles: 100, MaxBytes: 8 * 1024 * 1024, MaxResults: 100, MaxDepth: 3}
	options := repositorycontext.Options{Git: repositorycontext.Git{EnvironmentID: "s4-public-windows"}, Graph: repositorycontext.CodeGraph{Node: *node, SDK: *sdk, Script: *script}, AOCI: *aoci, CacheRoot: filepath.Join(*evidence, "projections"), SnapshotBudget: budget}
	e := start(db, options)
	defer func() { e.close(); must(db.Close()) }()
	prefix := "/api/v1/local-projects/s4-public-a/repository-context"
	other := "/api/v1/local-projects/s4-public-b/repository-context"
	query := func() map[string]any {
		return map[string]any{"schema_version": 1, "purpose": "verify_owned_public_http", "roots": []string{"."}, "budget": budget, "query": "Count"}
	}
	params := url.Values{"purpose": {"verify_owned_public_http"}, "root": {"."}, "max_files": {"100"}, "max_bytes": {"8388608"}, "max_results": {"100"}, "max_depth": {"3"}}
	func() {
		unconfigured := serve(db, &httpapi.RepositoryContextServices{})
		defer unconfigured.close()
		status, data := unconfigured.request("GET", prefix+"/map?"+params.Encode(), nil, "", false)
		expected(status, data, 403, "")
		unconfigured.bootstrap()
		status, data = unconfigured.request("POST", prefix+"/search", query(), "", false)
		expected(status, data, 403, "")
		status, data = unconfigured.request("POST", prefix+"/search", query(), "", true)
		expected(status, data, 503, "provider_unavailable")
		status, data = unconfigured.request("GET", prefix+"/map?"+params.Encode(), nil, "", false)
		expected(status, data, 503, "provider_unavailable")
		pass("unconfigured_capabilities_preserve_real_auth_and_report_unavailable")
	}()
	status, data := e.request("POST", prefix+"/search", query(), "", true)
	expected(status, data, 403, "")
	e.bootstrap()
	status, data = e.request("POST", prefix+"/search", query(), "", false)
	expected(status, data, 403, "")
	pass("production_session_and_csrf_enforced")
	status, data = e.request("POST", prefix+"/search", query(), "", true)
	expected(status, data, 200, "")
	result := sourceResult(data, p.Root, pb.Root)
	symbol := ""
	for _, node := range result.Symbols {
		if node.Name == "Count" && node.Span.Path == "textstats/stats.go" {
			symbol = node.ID
		}
	}
	require(symbol != "", "actual HTTP Count symbol is absent")
	impact := query()
	delete(impact, "query")
	impact["symbol_id"] = symbol
	status, data = e.request("POST", prefix+"/impact", impact, "", true)
	expected(status, data, 200, "")
	result = sourceResult(data, p.Root, pb.Root)
	require(len(result.Relations) > 0, "actual HTTP callers are absent")
	status, data = e.request("GET", prefix+"/map?"+params.Encode(), nil, "", false)
	expected(status, data, 200, "")
	result = sourceResult(data, p.Root, pb.Root)
	require(len(result.MapEntries) == 6, "actual HTTP authored map is incomplete")
	pass("real_map_search_impact_and_safe_http_projection")
	bad := query()
	bad["head"] = "untrusted-head"
	status, data = e.request("POST", prefix+"/search", bad, "", true)
	expected(status, data, 400, "validation_failed")
	bad = query()
	delete(bad, "query")
	status, data = e.request("POST", prefix+"/search", bad, "", true)
	expected(status, data, 400, "validation_failed")
	status, data = e.request("POST", prefix+"/impact", bad, "", true)
	expected(status, data, 400, "validation_failed")
	bad = query()
	bad["purpose"] = ""
	status, data = e.request("POST", prefix+"/search", bad, "", true)
	expected(status, data, 400, "validation_failed")
	bad = query()
	bad["query"] = strings.Repeat("C", 1025)
	status, data = e.request("POST", prefix+"/search", bad, "", true)
	expected(status, data, 400, "validation_failed")
	bad = query()
	delete(bad, "query")
	bad["symbol_id"] = strings.Repeat("C", 513)
	status, data = e.request("POST", prefix+"/impact", bad, "", true)
	expected(status, data, 400, "validation_failed")
	bad = query()
	bad["roots"] = []string{p.Root}
	status, data = e.request("POST", prefix+"/search", bad, "", true)
	expected(status, data, 403, "scope_denied")
	bad = query()
	limited := budget
	limited.MaxResults = 1
	bad["budget"] = limited
	status, data = e.request("POST", prefix+"/search", bad, "", true)
	expected(status, data, 409, "budget_exhausted")
	bad = query()
	bad["roots"] = []string{"../worktree-b"}
	status, data = e.request("POST", prefix+"/search", bad, "", true)
	expected(status, data, 403, "scope_denied")
	pass("http_strict_dto_relative_scope_and_budget_error_codes")
	status, data = e.request("GET", prefix+"/map?"+params.Encode()+"&unknown=%zz", nil, "", false)
	expected(status, data, 400, "validation_failed")
	status, data = e.request("GET", prefix+"/map?"+params.Encode()+"&unknown=x;y", nil, "", false)
	expected(status, data, 400, "validation_failed")
	status, data = e.request("GET", prefix+"/map?"+params.Encode()+"&max_depth=3", nil, "", false)
	expected(status, data, 400, "validation_failed")
	status, data = e.request("GET", prefix+"/map?"+params.Encode()+"&unknown=value", nil, "", false)
	expected(status, data, 400, "validation_failed")
	pass("malformed_map_encoding_unknown_parameters_and_duplicate_scalars_rejected")
	meta := func() map[string]any {
		return map[string]any{"schema_version": 1, "request_id": uuid.NewString(), "expected_version": p.Settings.Revision, "agent_id": "s4-http-reader"}
	}
	// 既有项目控制面使用真实Idempotency-Key，凭据只驻留内存。
	control := func(route string, body any) (int, []byte) {
		raw, err := json.Marshal(body)
		must(err)
		req, err := http.NewRequest("POST", e.base+"/api/v1/local-projects/s4-public-a/"+route, bytes.NewReader(raw))
		must(err)
		req.Header.Set("Origin", e.base)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", e.csrf)
		req.Header.Set("Idempotency-Key", uuid.NewString())
		response, err := e.client.Do(req)
		must(err)
		defer response.Body.Close()
		data, err := io.ReadAll(response.Body)
		must(err)
		return response.StatusCode, data
	}
	grant := meta()
	grant["actions"] = []string{"read_context"}
	status, data = control("grants", grant)
	expected(status, data, 200, "")
	status, data = control("agent-token", meta())
	expected(status, data, 200, "")
	var issued struct {
		Token domain.AgentToken `json:"credential"`
	}
	must(json.Unmarshal(data, &issued))
	require(issued.Token.Token != "", "real project token is absent")
	token := issued.Token.Token
	status, data = e.request("POST", prefix+"/search", query(), token, false)
	expected(status, data, 200, "")
	sourceResult(data, p.Root, pb.Root)
	status, data = e.request("GET", prefix+"/map?"+params.Encode(), nil, token, false)
	expected(status, data, 200, "")
	status, data = e.request("POST", prefix+"/impact", impact, token, false)
	expected(status, data, 200, "")
	index := query()
	delete(index, "query")
	index["action"] = "sync"
	status, data = e.request("POST", prefix+"/index", index, token, false)
	expected(status, data, 403, "")
	status, data = e.request("POST", prefix+"/map/refresh", query(), token, false)
	expected(status, data, 403, "")
	status, data = e.request("POST", other+"/search", query(), token, false)
	expected(status, data, 403, "")
	status, data = control("grants/revoke", meta())
	expected(status, data, 200, "")
	status, data = e.request("POST", prefix+"/search", query(), token, false)
	expected(status, data, 403, "")
	pass("real_project_token_reads_cross_scope_denial_human_only_maintenance_and_revoke")
	cmd := exec.Command("git", "-C", p.Root, "add", "textstats/stats.go")
	must(cmd.Run())
	status, data = e.request("POST", prefix+"/search", query(), "", true)
	expected(status, data, 409, "index_stale")
	status, data = e.request("GET", prefix+"/map?"+params.Encode(), nil, "", false)
	expected(status, data, 409, "index_stale")
	status, data = e.request("POST", other+"/search", query(), "", true)
	expected(status, data, 200, "")
	status, data = e.request("POST", prefix+"/index", index, "", true)
	expected(status, data, 200, "")
	status, data = e.request("POST", prefix+"/map/refresh", query(), "", true)
	expected(status, data, 200, "")
	status, data = e.request("POST", prefix+"/search", query(), "", true)
	expected(status, data, 200, "")
	before := sourceResult(data, p.Root, pb.Root).Index.Generation
	pass("actual_index_change_http_stale_other_worktree_and_explicit_recovery")
	e.close()
	must(db.Close())
	db, err = sqlite.Open(filepath.Join(*evidence, "state.sqlite"))
	must(err)
	must(db.RunMigrations(ctx, migrations.FS))
	e = start(db, options)
	e.bootstrap()
	status, data = e.request("POST", prefix+"/search", query(), "", true)
	expected(status, data, 200, "")
	result = sourceResult(data, p.Root, pb.Root)
	require(result.Index.Generation == before, "HTTP restart changed the saved generation")
	status, data = e.request("GET", prefix+"/map?"+params.Encode(), nil, "", false)
	expected(status, data, 200, "")
	sourceResult(data, p.Root, pb.Root)
	pass("actual_http_sqlite_restart_preserves_map_and_graph_generation")
	fmt.Println("PASS actual_repository_http")
}
