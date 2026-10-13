package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type relayHTTPSpy struct {
	workspace.RelayRuntime
	actions []string
	caller  domain.Caller
	err     error
}

type relayProjectSpy struct{ workspace.LocalProjects }

func (relayProjectSpy) ListProjects(context.Context, domain.Caller) ([]domain.LocalProject, error) {
	return []domain.LocalProject{{ID: "a", Settings: domain.ProjectSettings{}}, {ID: "b", Settings: domain.ProjectSettings{AllowDirectory: true}}, {ID: "c", Settings: domain.ProjectSettings{ExpandReferences: true}}, {ID: "bc", Settings: domain.ProjectSettings{AllowDirectory: true, ExpandReferences: true}}}, nil
}

func TestRelayProjectScopePreservesIndependentBandC(t *testing.T) {
	h := handler{services: Services{LocalProjects: relayProjectSpy{}}}
	r := httptest.NewRequest("GET", "http://127.0.0.1/", nil)
	v, e := h.relayProjects(r)
	if e != nil {
		t.Fatal("permission projection unavailable")
	}
	b, _ := json.Marshal(v)
	var result struct {
		Projects []relayProject `json:"projects"`
	}
	if json.Unmarshal(b, &result) != nil || len(result.Projects) != 4 {
		t.Fatal("invalid project scope projection")
	}
	for i, want := range []string{"A", "B", "C", "C"} {
		if result.Projects[i].Permission != want {
			t.Fatal("existing A/B/C semantics changed")
		}
	}
	if !result.Projects[3].AllowDirectory || !result.Projects[3].ExpandReferences {
		t.Fatal("independent B/C grant hidden")
	}
}

func (s *relayHTTPSpy) ListAgents(context.Context, domain.Caller) ([]domain.AgentSpec, error) {
	return nil, nil
}
func (s *relayHTTPSpy) ListTasks(context.Context, domain.Caller) ([]domain.RelayTaskView, error) {
	return []domain.RelayTaskView{{Task: domain.RelayTask{ID: "task-a", WorkID: "work-a"}}}, nil
}
func (s *relayHTTPSpy) GetTask(context.Context, domain.Caller, string) (domain.RelayTaskView, error) {
	return domain.RelayTaskView{}, s.err
}
func (s *relayHTTPSpy) ListReceipts(context.Context, domain.Caller, string) ([]domain.RelayReceipt, error) {
	return []domain.RelayReceipt{{ID: "receipt", OperationID: uuid.NewString(), TaskID: "task-a", Action: "dispatch", Status: "delivery_unknown", Error: &domain.RelayReceiptError{Code: "delivery_unknown", Message: "PRIVATE_CANARY prompt/native/sql/path", RequiredAction: "PRIVATE_CANARY"}}}, nil
}
func (s *relayHTTPSpy) ListEvents(context.Context, domain.Caller, string) ([]domain.RelayEvent, error) {
	return []domain.RelayEvent{{ID: "older", Kind: "edit_file", Topic: "topic", Scope: "scope", TS: time.Unix(1, 0), Agent: "PRIVATE_CANARY", TerminalID: "PRIVATE_CANARY"}, {ID: "newer", Kind: "commit", Topic: "topic", Scope: "scope", TS: time.Unix(2, 0), Agent: "PRIVATE_CANARY"}}, nil
}
func (s *relayHTTPSpy) CreateIntent(_ context.Context, c domain.Caller, _ workspace.RelayIntentCommand) (domain.RelayTaskView, error) {
	s.caller = c
	s.actions = append(s.actions, "intent")
	return domain.RelayTaskView{Task: domain.RelayTask{ID: "task-a", State: "ready"}}, nil
}
func (s *relayHTTPSpy) ConfirmIntent(_ context.Context, c domain.Caller, _ string, v workspace.RelayConfirmCommand) (domain.RelayActionResult, error) {
	s.caller = c
	s.actions = append(s.actions, "confirm")
	if !v.PermissionConfirmed {
		return domain.RelayActionResult{}, &apierrors.ServiceError{Code: apierrors.ApprovalRequired}
	}
	return domain.RelayActionResult{}, nil
}
func (s *relayHTTPSpy) Send(_ context.Context, c domain.Caller, _ string, _ workspace.RelaySendCommand) (domain.RelayActionResult, error) {
	s.caller = c
	s.actions = append(s.actions, "send")
	return domain.RelayActionResult{}, &apierrors.ServiceError{Code: apierrors.DeliveryUnknown, Message: "PRIVATE_CANARY"}
}

type relayHTTPClient struct {
	server *httptest.Server
	cookie *http.Cookie
	csrf   string
	t      *testing.T
}

func newRelayHTTPClient(t *testing.T, s workspace.RelayRuntime) *relayHTTPClient {
	t.Helper()
	server := httptest.NewServer(NewServer(Config{AgentToken: "agent-secret", Services: Services{Relay: s}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}).Handler())
	t.Cleanup(server.Close)
	c := &relayHTTPClient{server: server, t: t}
	resp := c.request("GET", "/api/v1/auth/session", "", nil)
	defer resp.Body.Close()
	c.cookie = resp.Cookies()[0]
	var b struct {
		CSRF string `json:"csrf_token"`
	}
	if json.NewDecoder(resp.Body).Decode(&b) != nil {
		t.Fatal("bootstrap JSON")
	}
	c.csrf = b.CSRF
	return c
}
func (c *relayHTTPClient) request(method, path, body string, headers map[string]string) *http.Response {
	c.t.Helper()
	req, err := http.NewRequest(method, c.server.URL+path, strings.NewReader(body))
	if err != nil {
		c.t.Fatal(err)
	}
	if c.cookie != nil {
		req.AddCookie(c.cookie)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", c.csrf)
		req.Header.Set("Idempotency-Key", uuid.NewString())
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	return resp
}
func relayHTTPBody(t *testing.T, r *http.Response, status int) []byte {
	t.Helper()
	defer r.Body.Close()
	b, e := io.ReadAll(r.Body)
	if e != nil || r.StatusCode != status {
		t.Fatalf("HTTP status=%d expected=%d", r.StatusCode, status)
	}
	if strings.Contains(string(b), "PRIVATE_CANARY") {
		t.Fatal("private canary leaked")
	}
	return b
}

func TestRelayRealHTTPGuardAndIntentCannotLaunch(t *testing.T) {
	s := &relayHTTPSpy{}
	c := newRelayHTTPClient(t, s)
	for _, p := range []string{"agents", "tasks", "tasks/task-a", "tasks/task-a/events", "tasks/task-a/receipts"} {
		cookie := c.cookie
		c.cookie = nil
		relayHTTPBody(t, c.request("GET", "/api/v1/relay/"+p, "", nil), 403)
		c.cookie = cookie
	}
	body := `{"work_id":"work-a","project_id":"project-a","agent":"agent-a","task":"human task"}`
	for _, headers := range []map[string]string{{"Origin": "https://foreign.invalid"}, {"Sec-Fetch-Site": "cross-site"}, {"X-CSRF-Token": "wrong"}, {"Authorization": "Bearer agent-secret"}, {"Idempotency-Key": "bad"}} {
		want := 403
		if headers["Idempotency-Key"] == "bad" {
			want = 400
		}
		relayHTTPBody(t, c.request("POST", "/api/v1/relay/intents", body, headers), want)
	}
	if len(s.actions) != 0 {
		t.Fatal("rejected requests called runtime")
	}
	relayHTTPBody(t, c.request("POST", "/api/v1/relay/intents", body, nil), 200)
	if len(s.actions) != 1 || s.actions[0] != "intent" || s.caller.Kind != "human" || s.caller.ID != "local-human" {
		t.Fatal("intent bypassed sole runtime or trusted caller")
	}
	if _, e := uuid.Parse(s.caller.OperationID); e != nil {
		t.Fatal("operation identity missing")
	}
	relayHTTPBody(t, c.request("POST", "/api/v1/relay/intents", `{"work_id":"work-a","actor_kind":"human","cmd":"unsafe"}`, nil), 400)
	relayHTTPBody(t, c.request("POST", "/api/v1/relay/intents/task-a/confirm", `{"permission_confirmed":false,"permission_level":"A","project_settings_revision":1}`, nil), 403)
	if len(s.actions) != 2 {
		t.Fatal("runtime did not receive denial to persist confirmation receipt")
	}
}

func TestRelayRealHTTPPrivacyOrderingAndUnknownNoRetry(t *testing.T) {
	s := &relayHTTPSpy{}
	c := newRelayHTTPClient(t, s)
	b := relayHTTPBody(t, c.request("GET", "/api/v1/relay/tasks/task-a/events", "", nil), 200)
	var events struct {
		Events []map[string]any `json:"events"`
	}
	if json.Unmarshal(b, &events) != nil || len(events.Events) != 2 || len(events.Events[0]) != 6 || events.Events[0]["event_id"] != "newer" || events.Events[0]["tokens"] != nil {
		t.Fatal("closed projection, reverse ordering or unknown token mismatch")
	}
	relayHTTPBody(t, c.request("GET", "/api/v1/relay/tasks/task-a/receipts", "", nil), 200)
	b = relayHTTPBody(t, c.request("GET", "/api/v1/relay/tasks", "", nil), 200)
	var tasks struct {
		Dispatched []string `json:"dispatched_work_ids"`
	}
	if json.Unmarshal(b, &tasks) != nil || len(tasks.Dispatched) != 0 {
		t.Fatal("unknown dispatch counted as confirmed relay")
	}
	relayHTTPBody(t, c.request("POST", "/api/v1/relay/tasks/task-a/send", `{"input":"human input"}`, nil), 409)
	if len(s.actions) != 1 || s.actions[0] != "send" {
		t.Fatal("unknown delivery retried")
	}
	s.err = &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "PRIVATE_CANARY SQL/private path", RequiredAction: "PRIVATE_CANARY"}
	relayHTTPBody(t, c.request("GET", "/api/v1/relay/tasks/task-a", "", nil), 503)
	for _, headers := range []map[string]string{{"Authorization": "Bearer agent-secret"}, {"Origin": "https://foreign.invalid"}} {
		relayHTTPBody(t, c.request("GET", "/api/v1/relay/tasks/task-a/events", "", headers), 403)
	}
}
